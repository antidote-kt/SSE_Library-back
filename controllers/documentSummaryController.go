package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const documentSummaryMaxRunes = 80000

// StreamDocumentSummary 对 PDF 文档：下载 → 抽取正文（同 RAG 流程）→ 调用 StreamChat 流式生成中文摘要。
// 需登录；公开文档任意登录用户可摘要，非公开仅上传者或管理员。
func StreamDocumentSummary(c *gin.Context) {
	idStr := c.Param("id")
	documentID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	var body struct {
		IsThink *bool `json:"isThink"`
	}
	_ = c.ShouldBindJSON(&body)
	isThink := false
	if body.IsThink != nil {
		isThink = *body.IsThink
	}

	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	document, err := dao.GetDocumentByID(documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	canAccess := document.Status == constant.DocumentStatusOpen ||
		document.UploaderID == userClaims.UserID ||
		userClaims.Role == "admin"
	if !canAccess {
		response.Fail(c, http.StatusForbidden, nil, constant.DocumentSummaryAccessDenied)
		return
	}

	if !utils.DocumentURLPathLooksLikePDF(document.URL) {
		response.Fail(c, http.StatusBadRequest, nil, constant.DocumentSummaryNotPDF)
		return
	}

	bodyText, err := utils.ExtractDocumentPDFPlainText(document.URL, documentSummaryMaxRunes)
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

	userMsg := "以下为从《" + document.Name + "》提取的 PDF 正文，请按要求输出摘要：\n\n" + bodyText

	_, err = utils.StreamChat(c, []utils.Message{{Role: "user", Content: userMsg}}, isThink, constant.DocumentSummarySystemPrompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
