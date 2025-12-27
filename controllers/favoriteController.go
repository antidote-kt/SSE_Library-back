package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

// CollectDocumentOrPost 通用收藏接口
// 允许用户将文档或帖子添加到收藏夹，包括验证用户身份、文档/帖子存在性、防止重复收藏等
func CollectDocumentOrPost(c *gin.Context) {
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

	// 使用通用处理函数处理收藏
	responseData, err := handleCollectAction(userClaims.UserID, request.SourceID, request.Type)
	if err != nil {
		// 如果文档不是公开状态，返回403禁止访问错误
		if err.Error() == constant.DocumentNotOpen {
			response.Fail(c, http.StatusForbidden, nil, constant.DocumentNotOpen)
			return
		}
		// 其他错误返回500
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 返回成功响应，携带用户收藏的相应资源列表
	response.SuccessWithData(c, responseData, constant.FavoriteSuccessMsg)
}

// 通用处理收藏操作的函数
func handleCollectAction(userID, sourceID uint64, sourceType string) (interface{}, error) {
	switch sourceType {
	case constant.DocumentType:
		return handleDocumentCollection(userID, sourceID)
	case constant.PostType:
		return handlePostCollection(userID, sourceID)
	default:
		return nil, errors.New(constant.FavoriteTypeNotAllow)
	}
}

// 处理文档收藏逻辑
func handleDocumentCollection(userID, documentID uint64) (interface{}, error) {
	// 验证文档是否存在
	document, err := dao.GetDocumentByID(documentID)
	if err != nil {
		// 如果文档不存在，返回错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(constant.DocumentNotExist)
		}
		// 其他数据库错误
		return nil, errors.New(constant.DatabaseError)
	}

	// 检查文档状态是否为open，如果不是open状态则不能收藏
	if document.Status != constant.DocumentStatusOpen {
		return nil, errors.New(constant.DocumentNotOpen)
	}

	// 检查该用户是否已经收藏了该文档（防止重复收藏）
	exists, err := dao.CheckFavoriteDocumentExist(userID, documentID)
	if err != nil {
		// 如果检查收藏状态失败，返回错误
		return nil, errors.New(constant.FavoriteStatusCheckFailed)
	}
	if exists {
		// 如果已经收藏，返回错误提示
		return nil, errors.New(constant.FavoriteAlreadyExistsMsg)
	}

	// 使用数据库事务确保操作的原子性
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建新的收藏记录
		favorite := models.Favorite{
			UserID:     userID,
			SourceID:   documentID,
			SourceType: constant.DocumentType,
		}
		// 将收藏记录插入数据库
		if err := tx.Create(&favorite).Error; err != nil {
			return fmt.Errorf(constant.FavoriteCreateFailed)
		}

		// 更新文档的收藏数（增加1）
		document.Collections++
		// 保存更新后的文档信息到数据库
		if err := tx.Save(&document).Error; err != nil {
			return fmt.Errorf(constant.CollectionUpdateFailed)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 获取用户收藏的所有文档列表
	favoriteDocuments, err := dao.GetFavoriteDocumentsByUserID(userID)
	if err != nil {
		// 如果获取收藏文档失败，返回错误
		return nil, errors.New(constant.GetFavoriteDocumentFailed)
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

	return responseData, nil
}

// 处理帖子收藏逻辑
func handlePostCollection(userID, postID uint64) (interface{}, error) {
	// 验证帖子是否存在
	post, err := dao.GetPostByID(postID)
	if err != nil {
		// 如果帖子不存在，返回错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(constant.PostNotExist)
		}
		// 其他数据库错误
		return nil, errors.New(constant.DatabaseError)
	}

	// 检查该用户是否已经收藏了该帖子（防止重复收藏）
	exists, err := dao.CheckFavoritePostExist(userID, postID)
	if err != nil {
		// 如果检查收藏状态失败，返回错误
		return nil, errors.New(constant.FavoriteStatusCheckFailed)
	}
	if exists {
		// 如果已经收藏，返回错误提示
		return nil, errors.New(constant.FavoriteAlreadyExistsMsg)
	}

	// 使用数据库事务确保操作的原子性
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建新的收藏记录
		favorite := models.Favorite{
			UserID:     userID,
			SourceID:   postID,
			SourceType: constant.PostType,
		}
		// 将收藏记录插入数据库
		if err := tx.Create(&favorite).Error; err != nil {
			return fmt.Errorf(constant.FavoriteCreateFailed)
		}

		// 更新帖子的收藏数（增加1）
		post.CollectCount++
		// 保存更新后的帖子信息到数据库
		if err := tx.Save(&post).Error; err != nil {
			return fmt.Errorf(constant.CollectionUpdateFailed)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 获取用户收藏的所有帖子列表
	favoritePosts, err := dao.GetFavoritePostsByUserID(userID)
	if err != nil {
		// 如果获取收藏帖子失败，返回错误
		return nil, errors.New(constant.GetFavoritePostFailed)
	}

	// 构造返回数据数组，包含用户所有收藏帖子的简要信息
	var responseData []response.PostBriefResponse
	// 遍历用户收藏的帖子，构建每个帖子的简要响应信息
	for _, favPost := range favoritePosts {
		// 构建单个帖子的简要信息
		postBriefResponse := response.BuildPostBriefResponse(favPost)

		// 将简要信息添加到响应数据数组
		responseData = append(responseData, postBriefResponse)
	}

	return responseData, nil
}

// WithdrawCollection 通用取消收藏接口
// 允许用户将文档或帖子从收藏夹中移除，包括验证用户身份、文档/帖子存在性、确认收藏状态等
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

	// 使用通用处理函数处理取消收藏
	responseData, err := handleUnCollectAction(userClaims.UserID, request.SourceID, request.Type)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 返回成功响应，携带用户取消收藏后的相应资源列表
	response.SuccessWithData(c, responseData, constant.UnfavoriteSuccessMsg)
}

// 通用处理取消收藏操作的函数
func handleUnCollectAction(userID, sourceID uint64, sourceType string) (interface{}, error) {
	switch sourceType {
	case constant.DocumentType:
		return handleDocumentUnCollection(userID, sourceID)
	case constant.PostType:
		return handlePostUnCollection(userID, sourceID)
	default:
		return nil, errors.New(constant.FavoriteTypeNotAllow)
	}
}

// 处理文档取消收藏逻辑
func handleDocumentUnCollection(userID, documentID uint64) (interface{}, error) {
	// 验证文档是否存在
	document, err := dao.GetDocumentByID(documentID)
	if err != nil {
		// 如果文档不存在，返回错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(constant.DocumentNotExist)
		}
		// 其他数据库错误
		return nil, errors.New(constant.DatabaseError)
	}

	// 检查是否已收藏（必须已收藏才能取消收藏）
	exists, err := dao.CheckFavoriteDocumentExist(userID, documentID)
	if err != nil {
		// 如果检查收藏状态失败，返回错误
		return nil, errors.New(constant.FavoriteStatusCheckFailed)
	}
	if !exists {
		// 如果没有收藏，返回错误提示
		return nil, errors.New(constant.FavoriteNotExistMsg)
	}

	// 使用数据库事务确保操作的原子性
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 根据用户ID和文档ID删除对应的收藏记录
		if err := tx.Where("user_id = ? AND source_id = ? AND source_type = ?", userID, documentID, constant.DocumentType).Delete(&models.Favorite{}).Error; err != nil {
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

	if err != nil {
		return nil, err
	}

	// 获取用户收藏的所有文档列表
	favoriteDocuments, err := dao.GetFavoriteDocumentsByUserID(userID)
	if err != nil {
		// 如果获取收藏文档失败，返回错误
		return nil, errors.New(constant.GetFavoriteDocumentFailed)
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

	return responseData, nil
}

// 处理帖子取消收藏逻辑
func handlePostUnCollection(userID, postID uint64) (interface{}, error) {
	// 验证帖子是否存在
	post, err := dao.GetPostByID(postID)
	if err != nil {
		// 如果帖子不存在，返回错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(constant.PostNotExist)
		}
		// 其他数据库错误
		return nil, errors.New(constant.DatabaseError)
	}

	// 检查是否已收藏（必须已收藏才能取消收藏）
	exists, err := dao.CheckFavoritePostExist(userID, postID)
	if err != nil {
		// 如果检查收藏状态失败，返回错误
		return nil, errors.New(constant.FavoriteStatusCheckFailed)
	}
	if !exists {
		// 如果没有收藏，返回错误提示
		return nil, errors.New(constant.FavoriteNotExistMsg)
	}

	// 使用数据库事务确保操作的原子性
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 根据用户ID和帖子ID删除对应的收藏记录
		if err := tx.Where("user_id = ? AND source_id = ? AND source_type = ?", userID, postID, constant.PostType).Delete(&models.Favorite{}).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.FavoriteDeleteFailed, err)
		}

		// 更新帖子的收藏数（减少1）
		post.CollectCount--
		if post.CollectCount < 0 {
			post.CollectCount = 0 // 确保收藏数不会小于0，防止出现负数
		}
		// 保存更新后的帖子信息到数据库
		if err := tx.Save(&post).Error; err != nil {
			return fmt.Errorf("%v: %v", constant.CollectionUpdateFailed, err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 获取用户收藏的所有帖子列表
	favoritePosts, err := dao.GetFavoritePostsByUserID(userID)
	if err != nil {
		// 如果获取收藏帖子失败，返回错误
		return nil, errors.New(constant.GetFavoritePostFailed)
	}

	// 构造返回数据数组，包含用户剩余收藏帖子的简要信息
	var responseData []response.PostBriefResponse
	// 遍历用户收藏的帖子，构建每个帖子的简要响应信息
	for _, favPost := range favoritePosts {
		// 构建单个帖子的简要信息
		postBriefResponse := response.BuildPostBriefResponse(favPost)

		// 将简要信息添加到响应数据数组
		responseData = append(responseData, postBriefResponse)
	}

	return responseData, nil
}

func CheckFavorite(c *gin.Context) {
	// 获取请求参数
	userIdStr := c.Query("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDFormatError)
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
	if userClaims.UserID != userId {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 验证用户是否存在
	_, err = dao.GetUserByID(userId)
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

	// 获取资源ID和类型
	sourceIdStr := c.Query("sourceId")
	sourceId, err := strconv.ParseUint(sourceIdStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgDocumentIDFormatError) // 使用通用ID格式错误
		return
	}

	sourceType := c.Query("type")
	if sourceType == "" {
		// 如果没有提供type参数，默认为document类型以保持兼容性
		sourceType = constant.DocumentType
	}

	// 根据类型检查收藏状态
	var isFavorite bool
	switch sourceType {
	case constant.DocumentType:
		isFavorite, err = dao.CheckFavoriteDocumentExist(userId, sourceId)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
	case constant.PostType:
		isFavorite, err = dao.CheckFavoritePostExist(userId, sourceId)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
	default:
		response.Fail(c, http.StatusBadRequest, nil, constant.FavoriteTypeNotAllow)
		return
	}

	// 构造返回数据
	result := gin.H{
		"judgement": isFavorite,
	}
	// 返回成功响应
	response.Success(c, result, constant.FavoriteGetSuccess)
}
