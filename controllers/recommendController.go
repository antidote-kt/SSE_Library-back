package controllers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

// GetBookRecommendations 获取书籍推荐
func GetBookRecommendations(c *gin.Context) {
	// 1. 获取当前用户 ID
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	userID := userClaims.UserID

	// 2. 获取用户最近浏览的书籍记录 (最多取出10本book)
	histories, _, err := dao.GetUserViewHistory(userID, "document", 1, 20)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	var recentBooks []string
	var recentBookTexts []string

	count := 0
	for _, h := range histories {
		if count >= 10 {
			break
		}
		doc, err := dao.GetDocumentByID(h.SourceID)
		if err == nil && doc.Type == "book" {
			category, _ := dao.GetCategoryByID(doc.CategoryID)
			// 用户历史浏览信息用于 prompt
			recentBooks = append(recentBooks, fmt.Sprintf("%d、《%s》——作者：%s", count+1, doc.Name, doc.Author))
			// 用于向量化
			recentBookTexts = append(recentBookTexts, fmt.Sprintf("%s %s %s", doc.Name, category.Name, doc.Introduction))
			count++
		}
	}

	var recommendIds []int64

	if len(recentBookTexts) == 0 {
		// 冷启动：获取阅读量最高的10本书
		topBooks, err := dao.GetTopReadBooks(10)
		if err != nil || len(topBooks) == 0 {
			response.SuccessWithData(c, []response.DocumentDetailResponse{}, "获取推荐成功")
			return
		}
		for _, b := range topBooks {
			recommendIds = append(recommendIds, int64(b.ID))
		}
	} else {
		// 3. 计算用户兴趣向量 (求均值)
		vectors, err := utils.GetEmbeddings(recentBookTexts)
		if err != nil || len(vectors) == 0 {
			response.Fail(c, http.StatusInternalServerError, nil, "兴趣向量计算失败")
			return
		}

		dim := len(vectors[0])
		avgVector := make([]float32, dim)
		for _, vec := range vectors {
			for i := 0; i < dim; i++ {
				avgVector[i] += vec[i]
			}
		}
		for i := 0; i < dim; i++ {
			avgVector[i] /= float32(len(vectors))
		}

		// 4. 从 Milvus 检索最相似的 10 本书
		recommendIds, err = utils.SearchBooks(avgVector, 10)
		if err != nil || len(recommendIds) == 0 {
			// 退级容错：取最热书籍
			topBooks, _ := dao.GetTopReadBooks(10)
			for _, b := range topBooks {
				recommendIds = append(recommendIds, int64(b.ID))
			}
		}
	}

	// Milvus中找不到相似推荐书籍，直接返回空数据（无需再ai重排序）
	if len(recommendIds) == 0 {
		response.SuccessWithData(c, []response.DocumentDetailResponse{}, "当前没有匹配用户浏览偏好的书籍")
		return
	}

	// 5. 准备大模型重排的 Prompt
	candidates, err := dao.GetDocumentsByIDs(recommendIds) // 注意这个函数只会提取通过审核的书籍
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	var candidateStrs []string
	for _, cand := range candidates {
		candidateStrs = append(candidateStrs, fmt.Sprintf("[%d] 《%s》——作者：%s", cand.ID, cand.Name, cand.Author))
	}
	log.Printf("推荐候选书籍如下：%s ,即将由ai重排推荐", candidateStrs)

	prompt := "你是一个专业的图书推荐助手。\n"
	if len(recentBooks) > 0 {
		prompt += "用户最近关注：\n" + strings.Join(recentBooks, "\n") + "\n\n"
	} else {
		prompt += "该用户是新用户，暂无浏览记录。\n\n"
	}

	prompt += "请从以下备选库中，挑选出最适合该用户的几本书并进行重排：\n"
	prompt += strings.Join(candidateStrs, "\n") + "\n\n"
	prompt += "请直接回复推荐的书籍ID，使用逗号分隔，不要输出其他废话。"

	// 调用大模型
	messages := []utils.Message{
		{Role: "user", Content: prompt},
	}
	aiResponse, err := utils.Chat(messages)
	if err != nil {
		// 如果AI调用失败，直接返回初筛结果
		aiResponse = ""
		for _, id := range recommendIds {
			aiResponse += fmt.Sprintf("%d,", id)
		}
	}

	// 6. 解析 AI 返回的 IDs
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(aiResponse, -1)

	finalIds := make([]int64, 0)
	idMap := make(map[int64]bool)
	for _, m := range matches {
		id, _ := strconv.ParseInt(m, 10, 64)
		if !idMap[id] {
			finalIds = append(finalIds, id)
			idMap[id] = true
		}
	}

	// 如果解析为空，退回到初筛
	if len(finalIds) == 0 {
		finalIds = recommendIds
	}

	// 获取最终书籍详情并构建响应
	finalDocs, _ := dao.GetDocumentsByIDs(finalIds)
	// 按 AI 返回顺序排序
	docMap := make(map[uint64]models.Document)
	for _, doc := range finalDocs {
		docMap[doc.ID] = doc
	}

	var respList []response.DocumentDetailResponse
	for _, id := range finalIds {
		if doc, ok := docMap[uint64(id)]; ok {
			detailResp, err := response.BuildDocumentDetailResponse(doc)
			if err == nil {
				respList = append(respList, detailResp)
			}
		}
	}

	// 如果最后仍然为空（例如id无效），则直接将初筛书籍构建返回
	if len(respList) == 0 {
		for _, doc := range candidates {
			detailResp, err := response.BuildDocumentDetailResponse(doc)
			if err == nil {
				respList = append(respList, detailResp)
			}
		}
	}

	response.SuccessWithData(c, respList, "获取推荐成功")
}
