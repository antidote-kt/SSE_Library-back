package controllers

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// CategoryHeatInfo 分类热度信息（用于Redis存储）
type CategoryHeatInfo struct {
	CategoryID uint64  `json:"categoryId"`
	Name       string  `json:"name"`
	IsCourse   bool    `json:"isCourse"`
	ReadCounts float64 `json:"readCounts"`
}

const (
	// 热度计算的天数范围
	HeatDays = 10
	// Redis ZSet的key
	HotCategoriesKey = "hot_categories_zset"
)

// calculateCategoryHeat 计算分类热度值
// 参考算法: heat = log10(readCounts + fileCounts/10) - daysSinceUpdate
func calculateCategoryHeat(readCounts, fileCounts int64, daysSinceUpdate float64) float64 {
	// 避免log10(0)的情况
	value := float64(readCounts) + float64(fileCounts)/10
	if value < 1 {
		value = 1
	}
	return math.Log10(value) - daysSinceUpdate
}

// RefreshCategoryHeat 刷新分类热度并存入Redis ZSet
func RefreshCategoryHeat() {
	ctx := context.Background()
	rdb := config.GetRedisClient()

	// 获取最近10天内有更新的分类
	categories, err := dao.GetRecentCategories(HeatDays)
	if err != nil {
		log.Println("获取最近分类失败:", err)
		return
	}

	// 如果没有最近更新的分类，则获取所有未删除的分类
	if len(categories) == 0 {
		allCategories, err := dao.GetAllCategories()
		if err != nil {
			log.Println("获取所有分类失败:", err)
			return
		}
		categories = allCategories
	}

	// 清空之前的ZSet
	rdb.Del(ctx, HotCategoriesKey)

	now := time.Now()

	// 计算每个分类的热度并存入Redis ZSet
	for _, category := range categories {
		// 统计文档数量
		fileCount, err := dao.CountDocumentsByCategory(category.ID)
		if err != nil {
			log.Println("统计分类文档数量失败:", err)
			continue
		}

		// 统计浏览量
		readCount, err := dao.GetDocumentReadCountsByCategory(category.ID)
		if err != nil {
			log.Println("统计分类浏览量失败:", err)
			continue
		}

		// 计算距离更新的天数
		daysSinceUpdate := math.Floor(now.Sub(category.UpdatedAt).Hours() / 24)

		// 计算热度
		heatScore := calculateCategoryHeat(readCount, fileCount, daysSinceUpdate)

		// 创建分类信息JSON
		categoryInfo := CategoryHeatInfo{
			CategoryID: category.ID,
			Name:       category.Name,
			IsCourse:   category.IsCourse,
			ReadCounts: float64(readCount),
		}
		categoryInfoJSON, err := json.Marshal(categoryInfo)
		if err != nil {
			log.Println("Error marshalling categoryInfo:", err)
			continue
		}

		// 存入Redis ZSet
		rdb.ZAdd(ctx, HotCategoriesKey, &redis.Z{
			Score:  heatScore,
			Member: categoryInfoJSON,
		})
	}

	log.Println("分类热度已刷新")
}

// GetHotCategories 获取热门分类
// GET /api/user/hotCategories?count=10
func GetHotCategories(c *gin.Context) {
	// 获取count参数，默认返回10个
	countStr := c.Query("count")
	count := 10
	if countStr != "" {
		if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	ctx := context.Background()
	rdb := config.GetRedisClient()

	// 初始化响应数组，确保即使没有数据也返回空数组而不是 null
	categoryResponses := make([]*CategoryResponse, 0)

	// 检查Redis中是否有数据，如果没有则刷新
	exists, err := rdb.Exists(ctx, HotCategoriesKey).Result()
	if err != nil || exists == 0 {
		RefreshCategoryHeat()
	}

	// 从Redis ZSet中获取热度最高的分类（降序）
	hotCategoriesJSON, err := rdb.ZRevRange(ctx, HotCategoriesKey, 0, int64(count-1)).Result()
	if err != nil {
		// Redis 查询失败，返回空数组而不是错误
		log.Println("Redis查询失败:", err)
		response.SuccessWithData(c, categoryResponses, constant.MsgGetHotCategoriesSuccess)
		return
	}

	// 解析JSON并构建完整的响应
	for _, jsonStr := range hotCategoriesJSON {
		var heatInfo CategoryHeatInfo
		if err := json.Unmarshal([]byte(jsonStr), &heatInfo); err != nil {
			log.Println("Error unmarshalling categoryInfo:", err)
			continue
		}

		// 获取完整的分类信息
		category, err := dao.GetCategoryByID(heatInfo.CategoryID)
		if err != nil {
			log.Println("获取分类详情失败:", err)
			continue
		}

		// 获取文档数量和浏览量
		fileCount, err := dao.CountDocumentsByCategory(category.ID)
		if err != nil {
			fileCount = 0
		}

		readCount, err := dao.GetDocumentReadCountsByCategory(category.ID)
		if err != nil {
			readCount = 0
		}

		// 构建响应对象
		categoryResp := &CategoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			IsCourse:    category.IsCourse,
			FileCounts:  fileCount,
			ReadCounts:  readCount,
			Description: category.Description,
			ParentID:    category.ParentID,
			Children:    make([]*CategoryResponse, 0),
		}

		categoryResponses = append(categoryResponses, categoryResp)
	}

	response.SuccessWithData(c, categoryResponses, constant.MsgGetHotCategoriesSuccess)
}
