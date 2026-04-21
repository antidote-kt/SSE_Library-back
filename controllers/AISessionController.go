package controllers

import (
	"net/http"
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

func CreateAISession(c *gin.Context) {
	var req dto.CreateAISessionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}

	userClaims := claims.(*utils.MyClaims)

	if userClaims.UserID != req.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	newAISession := models.AISession{
		UserID: userClaims.UserID,
		Title:  "新对话",
	}

	if err := dao.CreateAISession(&newAISession); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.CreateAISessionFailed)
		return
	}

	aiSessionResponse := response.BuildCreateAISessionResponse(newAISession)
	response.SuccessWithData(c, aiSessionResponse, constant.CreateAISessionSuccess)
}

func GetAISessions(c *gin.Context) {
	userIdStr := c.Query("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}

	userClaims := claims.(*utils.MyClaims)

	if userClaims.UserID != userId {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	aiSessions, err := dao.GetAISessionsByUserID(userClaims.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.GetAISessionsFailed)
		return
	}

	aiSessionResponses := response.BuildAISessionListResponses(aiSessions)
	response.SuccessWithData(c, aiSessionResponses, constant.GetAISessionsSuccess)
}

func UpdateAISession(c *gin.Context) {
	var req dto.UpdateAISessionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}

	userClaims := claims.(*utils.MyClaims)

	if userClaims.UserID != req.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	aiSession, err := dao.GetAISessionByID(req.ID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, nil, constant.AISessionNotExist)
		return
	}

	if userClaims.UserID != aiSession.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	if req.Title != "" {
		aiSession.Title = req.Title
	}

	if err := dao.UpdateAISession(&aiSession); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.UpdateAISessionFailed)
		return
	}

	aiSessionResponse := response.BuildAISessionListItemResponse(aiSession)
	response.SuccessWithData(c, aiSessionResponse, constant.UpdateAISessionSuccess)
}
