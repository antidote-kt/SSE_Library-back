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

// GetProfile 获取当前登录用户的个人主页信息
func GetProfile(c *gin.Context) {
	// 1. 从JWT中间件获取用户信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	//将路径参数的用户id提取出来并转化为int64类型，与JWT比较看访问的个人主页接口是否与用户本人匹配
	paramUserID := c.Param("user_id")
	targetID, err := strconv.ParseUint(paramUserID, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDFormatError)
		return
	}
	userID := userClaims.UserID
	if userID != targetID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 2. 获取用户基本信息、收藏列表、历史记录
	var userInfo models.User
	var collectionList []models.Document
	var historyList []models.Document

	userInfo, err = dao.GetUserByID(userID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
		return
	}

	collectionList, err = dao.GetFavoriteDocumentsByUserID(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.GetDataFailed+err.Error())
		return
	}

	historyList, err = dao.GetViewHistoryDocumentsByUserID(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.GetDataFailed+err.Error())
		return
	}

	// 3. 调用response层组装返回数据
	homepageResponse, err := response.BuildHomepageResponse(userInfo, collectionList, historyList)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.ConstructDataFailed)
		return
	}

	// 4. 返回响应
	response.SuccessWithData(c, homepageResponse, constant.GetUserProfileSuccess)
}

// ModifyInfo 修改当前登录用户的个人资料
func ModifyInfo(c *gin.Context) {
	var req dto.ModifyInfoDTO

	// 1. 绑定并验证参数
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError+err.Error())
		return
	}

	// 2. 从JWT中间件获取用户信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	// 3. 从数据库获取当前用户
	user, err := dao.GetUserByID(userClaims.UserID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
		return
	}

	// 4. 检查是否有更新，并应用更新
	// 使用指针的好处：如果前端没传某个字段，这里的指针就是nil，我们就不更新它
	updated := false
	if req.UserName != nil && *req.UserName != user.Username {
		// 检查新用户名是否已被其他用户占用
		existingUser, _ := dao.GetUserByUsername(*req.UserName)
		if existingUser.ID != 0 && existingUser.ID != user.ID {
			response.Fail(c, http.StatusUnprocessableEntity, nil, constant.UserNameAlreadyExist)
			return
		}
		user.Username = *req.UserName
		updated = true
	}
	if req.Email != nil && *req.Email != user.Email {
		// 检查新邮箱是否已被其他用户占用
		existingUser, _ := dao.GetUserByEmail(*req.Email)
		if existingUser.ID != 0 && existingUser.ID != user.ID {
			response.Fail(c, http.StatusUnprocessableEntity, nil, constant.EmailHasBeenUsed)
			return
		}
		user.Email = *req.Email
		updated = true
	}
	if req.UserAvatar != nil && req.UserAvatar.Size != 0 {
		// 检查用户是否上传了新头像，有的话删除原头像
		err := utils.DeleteFile(user.Avatar)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.AvatarDeleteFailed)
			return
		}
		// 然后上传新头像到腾讯云
		avatarURL, err := utils.UploadAvatar(req.UserAvatar)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, err.Error())
			return
		}
		//最后更新头像链接到数据模型
		user.Avatar = avatarURL
		updated = true
	}

	// 如果没有任何更新，可以直接返回成功
	if !updated {
		response.Success(c, nil, constant.NoChangeHappen)
		return
	}

	// 5. 保存到数据库
	if err := dao.UpdateUser(user); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.UpdateUserInfoFailed)
		return
	}

	// 6. 调用response层的结构体组装返回数据
	userBrief := response.UserBriefResponse{
		UserID:     user.ID,
		Username:   user.Username,
		UserAvatar: utils.GetFileURL(user.Avatar), // 使用工具函数处理头像URL
		Status:     user.Status,
		CreateTime: user.CreatedAt.Format("2006-01-02 15:04:05"),
		Email:      user.Email,
		Role:       user.Role,
	}

	// 7. 按照要求，返回更新后的UserBrief信息
	response.SuccessWithData(c, userBrief, constant.UpdateUserInfoSuccess)
}
