package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PostAISummary 查看/生成 AI 摘要：未要求 regenerate 且 Redis 中已有与当前正文 hash 一致的缓存则直接返回；否则调用 Chat 生成并写入 Redis（不新增业务表）。
func PostAISummary(c *gin.Context) {
	contentType := strings.ToLower(strings.TrimSpace(c.Param("contentType")))
	if contentType != "document" && contentType != "post" {
		response.Fail(c, http.StatusBadRequest, nil, constant.AISummaryInvalidContentType)
		return
	}

	idStr := strings.TrimSpace(c.Param("contentId"))
	contentID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || contentID == 0 {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	var reqBody struct {
		Regenerate bool `json:"regenerate"`
	}
	_ = c.ShouldBindJSON(&reqBody)

	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	var sourceText string

	switch contentType {
	case "document":
		doc, err := dao.GetDocumentByID(contentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
		can := doc.Status == constant.DocumentStatusOpen ||
			doc.UploaderID == userClaims.UserID ||
			userClaims.Role == "admin"
		if !can {
			response.Fail(c, http.StatusForbidden, nil, constant.DocumentSummaryAccessDenied)
			return
		}
		if !utils.DocumentURLPathLooksLikePDF(doc.URL) {
			response.Fail(c, http.StatusBadRequest, nil, constant.DocumentSummaryNotPDF)
			return
		}
		sourceText, err = utils.ExtractDocumentPDFPlainText(doc.URL, documentSummaryMaxRunes)
		if err != nil {
			if errors.Is(err, utils.ErrSummaryEmptyText) {
				response.Fail(c, http.StatusBadRequest, nil, constant.DocumentSummaryNoText)
				return
			}
			if errors.Is(err, utils.ErrSummaryNotPDF) {
				response.Fail(c, http.StatusBadRequest, nil, constant.DocumentSummaryNotPDF)
				return
			}
			response.Fail(c, http.StatusBadGateway, nil, err.Error())
			return
		}

	case "post":
		post, err := dao.GetPostByID(contentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.PostNotExist)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
		src := "标题：" + post.Title + "\n\n正文：\n" + post.Content
		sourceText = utils.TruncateRunesForSummary(src, documentSummaryMaxRunes)
		if strings.TrimSpace(sourceText) == "" {
			response.Fail(c, http.StatusBadRequest, nil, constant.DocumentSummaryNoText)
			return
		}
	}

	srcHash := utils.HashSourceText(sourceText)
	ctx := config.Ctx

	if !reqBody.Regenerate {
		cached, err := utils.GetAISummaryFromCache(ctx, contentType, contentID, srcHash)
		if err != nil {
			log.Printf("[AISummary] redis get: %v", err)
		} else if cached != nil {
			response.SuccessWithDataCodeZero(c, response.AISummaryData{
				FromCache:   true,
				ContentType: contentType,
				ContentID:   contentID,
				SummaryID:   cached.SummaryID,
				Summary:     cached.Summary,
			}, constant.AISummarySuccess)
			return
		}
	}

	userBlock := "请根据以下素材输出摘要：\n\n" + sourceText
	msgs := []utils.Message{
		{Role: "system", Content: constant.DocumentSummarySystemPrompt},
		{Role: "user", Content: userBlock},
	}
	summaryText, err := utils.Chat(msgs)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	summaryID := uint64(time.Now().UnixNano())
	cacheVal := &utils.AISummaryCacheValue{
		Summary:   summaryText,
		SummaryID: summaryID,
		SrcHash:   srcHash,
	}
	if err := utils.SetAISummaryCache(ctx, contentType, contentID, cacheVal); err != nil {
		log.Printf("[AISummary] redis set: %v", err)
	}

	response.SuccessWithDataCodeZero(c, response.AISummaryData{
		FromCache:   false,
		ContentType: contentType,
		ContentID:   contentID,
		SummaryID:   summaryID,
		Summary:     summaryText,
	}, constant.AISummarySuccess)
}
