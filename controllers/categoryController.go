package controllers

import (
	"errors"
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CategoryResponse
type CategoryResponse struct {
	ID          uint64              `json:"id"`
	Name        string              `json:"name"`
	IsCourse    bool                `json:"isCourse"`
	FileCounts  int64               `json:"fileCounts"`
	ReadCounts  int64               `json:"readCounts"`
	Description string              `json:"description,omitempty"`
	ParentID    *uint64             `json:"parentId,omitempty"`
	Children    []*CategoryResponse `json:"children,omitempty"`
}

// buildCategoryTree 构建分类树
func buildCategoryTree(categories []models.Category, fileCounts map[uint64]int64, readCounts map[uint64]int64) []*CategoryResponse {
	// 分类ID对应分类映射
	categoryMap := make(map[uint64]*CategoryResponse)
	var rootCategories []*CategoryResponse

	for _, category := range categories {
		categoryResp := &CategoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			IsCourse:    category.IsCourse,
			FileCounts:  fileCounts[category.ID],
			ReadCounts:  readCounts[category.ID],
			Description: category.Description,
			ParentID:    category.ParentID,
			Children:    make([]*CategoryResponse, 0),
		}
		categoryMap[category.ID] = categoryResp
	}

	for _, category := range categories {
		categoryResp := categoryMap[category.ID]
		if category.ParentID == nil {
			// 顶级分类
			rootCategories = append(rootCategories, categoryResp)
		} else {
			// 将子分类添加到父分类的children中
			if parent, ok := categoryMap[*category.ParentID]; ok {
				parent.Children = append(parent.Children, categoryResp)
			}
		}
	}

	return rootCategories
}

// GetCategoriesAndCourses 获取所有分类和课程
// GET /api/category
func GetCategoriesAndCourses(c *gin.Context) {
	// 获取所有分类和课程
	categories, err := dao.GetAllCategories()
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	// 统计每类文档的数量和浏览量
	fileCounts := make(map[uint64]int64)
	readCounts := make(map[uint64]int64)

	for _, category := range categories {
		// 统计文档数量
		fileCount, err := dao.CountDocumentsByCategory(category.ID)
		if err != nil {
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
		fileCounts[category.ID] = fileCount

		// 统计浏览量
		readCount, err := dao.GetDocumentReadCountsByCategory(category.ID)
		if err != nil {
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
		readCounts[category.ID] = readCount
	}

	// 构建分类树
	categoryTree := buildCategoryTree(categories, fileCounts, readCounts)

	response.SuccessWithData(c, categoryTree, constant.MsgGetCategoriesSuccess)
}

// SearchCategoriesAndCourses 搜索分类和课程
// GET /api/searchcat?name=xxx
func SearchCategoriesAndCourses(c *gin.Context) {
	// 获取分类和课程
	name := c.Query("name")
	if name == "" {
		// 如果没有搜索关键词，返回所有分类和课程
		GetCategoriesAndCourses(c)
		return
	}

	// 搜索分类和课程
	categories, err := dao.SearchCategoriesByName(name)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	// 统计每类文档的数量和浏览量
	fileCounts := make(map[uint64]int64)
	readCounts := make(map[uint64]int64)

	for _, category := range categories {
		// 统计文档数量
		fileCount, err := dao.CountDocumentsByCategory(category.ID)
		if err != nil {
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCategoryCountFailed)
			return
		}
		fileCounts[category.ID] = fileCount

		// 统计浏览量
		readCount, err := dao.GetDocumentReadCountsByCategory(category.ID)
		if err != nil {
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCategoryReadCountFailed)
			return
		}
		readCounts[category.ID] = readCount
	}

	// 构建分类树
	categoryTree := buildCategoryTree(categories, fileCounts, readCounts)

	response.SuccessWithData(c, categoryTree, constant.MsgGetCategoriesSuccess)
}

// AddCategory 添加分类/课程接口
func AddCategory(c *gin.Context) {
	var req dto.AddCategoryDTO

	// 1. 绑定参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 业务校验
	// 2.1 如果指定了父分类，检查父分类是否存在
	if req.ParentCatID != nil && *req.ParentCatID != 0 {
		_, err := dao.GetCategoryByID(*req.ParentCatID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusBadRequest, nil, constant.ParentCategoryNotExist)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
		//可选：检查父分类是否允许有子分类（例如，如果父分类本身是课程，不允许再有子分类）
		//if parentCategory.IsCourse {
		// response.Fail(c, http.StatusBadRequest, nil, "课程不能作为父分类")
		// return
		//}
	}

	// 2.2 检查同名分类 (防止重复)
	existingCats, _ := dao.GetCategoryByName(req.Name)
	if len(existingCats) > 0 {
		// 简单的重名检查，更严格的检查应该看是否在同一个父分类下重名
		response.Fail(c, http.StatusUnprocessableEntity, nil, constant.CategoryNameAlreadyExist)
		return
	}

	// 3. 构建模型
	category := models.Category{
		Name:        req.Name,
		IsCourse:    *req.IsCourse,
		Description: req.Description,
		ParentID:    req.ParentCatID,
		// 其他三个时间相关字段在dao.CreateCategory(&category)这一步会由gorm自动处理
	}

	// 4. 保存到数据库
	if err := dao.CreateCategory(&category); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.MsgCategoryCreateFailed)
		return
	}

	// 5. 返回成功响应
	response.Success(c, gin.H{"success": true}, constant.MsgCategoryCreateSuccess)
}
