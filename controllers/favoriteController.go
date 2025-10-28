package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 构造文档信息的公共函数

func CollectDocument(c *gin.Context) {
	var request dto.FavoriteDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	if userClaims.UserID != request.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}
	// 验证用户是否存在
	_, err := dao.GetUserByID(userClaims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 验证文档是否存在
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 检查是否已经收藏（防止重复收藏）
	exists, err = dao.CheckFavoriteExist(userClaims.UserID, request.DocumentID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.FavoriteStatusCheckFailed)
		return
	}
	if exists {
		response.Fail(c, http.StatusBadRequest, nil, constant.FavoriteAlreadyExistsMsg)
		return
	}

	// 使用事务处理收藏操作
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建收藏记录
		favorite := models.Favorite{
			UserID:     userClaims.UserID,
			DocumentID: request.DocumentID,
		}
		if err := tx.Create(&favorite).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.FavoriteCreateFailed, err)
		}

		// 更新文档的收藏数
		document.Collections++
		if err := tx.Save(&document).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.CollectionUpdateFailed, err)
		}

		return nil
	})

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 获取用户收藏的所有文档
	favoriteDocuments, err := dao.GetFavoriteDocumentsByUserID(userClaims.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.GetFavoriteDocumentFailed)
		return
	}

	// 构造返回数据数组
	var responseData []response.InfoBriefResponse
	for _, favDoc := range favoriteDocuments {

		infoBriefResponse, _ := response.BuildInfoBriefResponse(favDoc)

		responseData = append(responseData, infoBriefResponse)
	}

	// 返回收藏的文档列表
	response.SuccessWithData(c, responseData, constant.FavoriteSuccessMsg)
}

func WithdrawCollection(c *gin.Context) {
	var request dto.FavoriteDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	if userClaims.UserID != request.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 验证用户是否存在
	_, err := dao.GetUserByID(request.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 验证文档是否存在
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 检查是否已收藏（必须已收藏才能取消收藏）
	exists, err = dao.CheckFavoriteExist(userClaims.UserID, request.DocumentID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.FavoriteStatusCheckFailed)
		return
	}
	if !exists {
		response.Fail(c, http.StatusBadRequest, nil, constant.FavoriteNotExistMsg)
		return
	}

	// 使用事务处理取消收藏操作
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除收藏记录
		if err := tx.Where("user_id = ? AND document_id = ?", userClaims.UserID, request.DocumentID).Delete(&models.Favorite{}).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.FavoriteDeleteFailed, err)
		}

		// 更新文档的收藏数（减少1）
		document.Collections--
		if document.Collections < 0 {
			document.Collections = 0 // 确保收藏数不会小于0
		}
		if err := tx.Save(&document).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.CollectionUpdateFailed, err)
		}

		return nil
	})

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 获取用户收藏的所有文档
	favoriteDocuments, err := dao.GetFavoriteDocumentsByUserID(userClaims.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.GetFavoriteDocumentFailed)
		return
	}

	// 构造返回数据数组
	var responseData []response.InfoBriefResponse
	for _, favDoc := range favoriteDocuments {
		infoBriefResponse, _ := response.BuildInfoBriefResponse(favDoc)

		responseData = append(responseData, infoBriefResponse)
	}

	// 返回剩余的收藏文档列表
	response.SuccessWithData(c, responseData, constant.UnfavoriteSuccessMsg)
}
