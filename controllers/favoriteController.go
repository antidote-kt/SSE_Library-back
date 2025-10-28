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

// CollectDocument 文档收藏接口
// 允许用户将文档添加到收藏夹，包括验证用户身份、文档存在性、防止重复收藏等
func CollectDocument(c *gin.Context) {
	// 声明收藏请求参数结构体
	var request dto.FavoriteDTO

	// 解析JSON请求参数
	if err := c.ShouldBindJSON(&request); err != nil {
		// 如果参数解析失败，返回400错误
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 从上下文中获取用户JWT声明信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		// 如果无法获取用户信息，返回401未授权错误
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	// 将接口类型转换为具体的声明结构体
	userClaims := claims.(*utils.MyClaims)

	// 验证请求的用户ID是否与JWT中的用户ID一致（防止越权操作）
	if userClaims.UserID != request.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 验证用户是否存在
	_, err := dao.GetUserByID(userClaims.UserID)
	if err != nil {
		// 如果用户不存在，返回404错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
			return
		}
		// 其他数据库错误，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 验证文档是否存在
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		// 如果文档不存在，返回404错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		// 其他数据库错误，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 检查该用户是否已经收藏了该文档（防止重复收藏）
	exists, err = dao.CheckFavoriteExist(userClaims.UserID, request.DocumentID)
	if err != nil {
		// 如果检查收藏状态失败，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.FavoriteStatusCheckFailed)
		return
	}
	if exists {
		// 如果已经收藏，返回400错误提示
		response.Fail(c, http.StatusBadRequest, nil, constant.FavoriteAlreadyExistsMsg)
		return
	}

	// 使用数据库事务确保操作的原子性
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建新的收藏记录
		favorite := models.Favorite{
			UserID:     userClaims.UserID,
			DocumentID: request.DocumentID,
		}
		// 将收藏记录插入数据库
		if err := tx.Create(&favorite).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.FavoriteCreateFailed, err)
		}

		// 更新文档的收藏数（增加1）
		document.Collections++
		// 保存更新后的文档信息到数据库
		if err := tx.Save(&document).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.CollectionUpdateFailed, err)
		}

		return nil
	})

	// 检查事务执行结果
	if err != nil {
		// 如果事务执行失败，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 获取用户收藏的所有文档列表
	favoriteDocuments, err := dao.GetFavoriteDocumentsByUserID(userClaims.UserID)
	if err != nil {
		// 如果获取收藏文档失败，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.GetFavoriteDocumentFailed)
		return
	}

	// 构造返回数据数组，包含用户所有收藏文档的简要信息
	var responseData []response.InfoBriefResponse
	// 遍历用户收藏的文档，构建每个文档的简要响应信息
	for _, favDoc := range favoriteDocuments {
		// 构建单个文档的简要信息
		infoBriefResponse, _ := response.BuildInfoBriefResponse(favDoc)

		// 将简要信息添加到响应数据数组
		responseData = append(responseData, infoBriefResponse)
	}

	// 返回成功响应，携带用户收藏的所有文档列表
	response.SuccessWithData(c, responseData, constant.FavoriteSuccessMsg)
}

// WithdrawCollection 文档取消收藏接口
// 允许用户将文档从收藏夹中移除，包括验证用户身份、文档存在性、确认收藏状态等
func WithdrawCollection(c *gin.Context) {
	// 声明收藏请求参数结构体
	var request dto.FavoriteDTO

	// 解析JSON请求参数
	if err := c.ShouldBindJSON(&request); err != nil {
		// 如果参数解析失败，返回400错误
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 从上下文中获取用户JWT声明信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		// 如果无法获取用户信息，返回401未授权错误
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	// 将接口类型转换为具体的声明结构体
	userClaims := claims.(*utils.MyClaims)

	// 验证请求的用户ID是否与JWT中的用户ID一致（防止越权操作）
	if userClaims.UserID != request.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 验证用户是否存在
	_, err := dao.GetUserByID(request.UserID)
	if err != nil {
		// 如果用户不存在，返回404错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
			return
		}
		// 其他数据库错误，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 验证文档是否存在
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		// 如果文档不存在，返回404错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		// 其他数据库错误，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 检查是否已收藏（必须已收藏才能取消收藏）
	exists, err = dao.CheckFavoriteExist(userClaims.UserID, request.DocumentID)
	if err != nil {
		// 如果检查收藏状态失败，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.FavoriteStatusCheckFailed)
		return
	}
	if !exists {
		// 如果没有收藏，返回400错误提示
		response.Fail(c, http.StatusBadRequest, nil, constant.FavoriteNotExistMsg)
		return
	}

	// 使用数据库事务确保操作的原子性
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 根据用户ID和文档ID删除对应的收藏记录
		if err := tx.Where("user_id = ? AND document_id = ?", userClaims.UserID, request.DocumentID).Delete(&models.Favorite{}).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.FavoriteDeleteFailed, err)
		}

		// 更新文档的收藏数（减少1）
		document.Collections--
		if document.Collections < 0 {
			document.Collections = 0 // 确保收藏数不会小于0，防止出现负数
		}
		// 保存更新后的文档信息到数据库
		if err := tx.Save(&document).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.CollectionUpdateFailed, err)
		}

		return nil
	})

	// 检查事务执行结果
	if err != nil {
		// 如果事务执行失败，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 获取用户收藏的所有文档列表（移除当前文档后剩余的收藏）
	favoriteDocuments, err := dao.GetFavoriteDocumentsByUserID(userClaims.UserID)
	if err != nil {
		// 如果获取收藏文档失败，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.GetFavoriteDocumentFailed)
		return
	}

	// 构造返回数据数组，包含用户剩余收藏文档的简要信息
	var responseData []response.InfoBriefResponse
	// 遍历用户收藏的文档，构建每个文档的简要响应信息
	for _, favDoc := range favoriteDocuments {
		// 构建单个文档的简要信息
		infoBriefResponse, _ := response.BuildInfoBriefResponse(favDoc)

		// 将简要信息添加到响应数据数组
		responseData = append(responseData, infoBriefResponse)
	}

	// 返回成功响应，携带用户剩余收藏的所有文档列表
	response.SuccessWithData(c, responseData, constant.UnfavoriteSuccessMsg)
}
