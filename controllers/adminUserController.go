package controllers

import (
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/gin-gonic/gin"
)

// UpdateUserStatus 管理员修改指定用户的状态
func UpdateUserStatus(c *gin.Context) {
	var req dto.UpdateUserStatusDTO

	// 1. 绑定Query参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 验证status字段是否合法 (例如：只能是 "active" 或 "disabled")
	if req.Status != "active" && req.Status != "disabled" {
		response.Fail(c, http.StatusBadRequest, nil, constant.IllegalStatus)
		return
	}

	// 3. 获取要修改的用户
	user, err := dao.GetUserByID(req.UserID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
		return
	}

	// 4. 更新状态并保存到数据库
	user.Status = req.Status
	if err := dao.UpdateUser(user); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.UpdateUserStatusFailed)
		return
	}

	// 5.调用response层的结构体组装返回数据
	userBrief := response.UserBriefResponse{
		UserID:     user.ID,
		Username:   user.Username,
		UserAvatar: user.Avatar,
		Status:     user.Status,
		CreateTime: user.CreatedAt.Format("2006-01-02 15:04:05"),
		Email:      user.Email,
		Role:       user.Role,
	}

	// 6. 构造并返回更新后的用户信息
	response.SuccessWithData(c, userBrief, constant.UpdateUserStatusSuccess)
}
