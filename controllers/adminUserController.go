package controllers

import (
	"net/http"

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
		response.Fail(c, http.StatusBadRequest, nil, "参数无效: "+err.Error())
		return
	}

	// 2. 验证status字段是否合法 (例如：只能是 "active" 或 "disabled")
	if req.Status != "active" && req.Status != "disabled" {
		response.Fail(c, http.StatusBadRequest, nil, "无效的状态值")
		return
	}

	// 3. 获取要修改的用户
	user, err := dao.GetUserByID(req.UserID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, nil, "用户不存在")
		return
	}

	// 4. 更新状态并保存到数据库
	user.Status = req.Status
	if err := dao.UpdateUser(user); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "更新用户状态失败")
		return
	}

	// 5. 构造并返回更新后的用户信息
	response.Success(c, gin.H{
		"data": buildUserBriefDTO(user), // 复用user_controller中的辅助函数
	}, "用户状态更新成功")
}

// GetUsers 获取或搜索用户列表
func GetUsers(c *gin.Context) {
	var req dto.SearchUsersDTO

	// 1. 绑定可选的Query参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数无效: "+err.Error())
		return
	}

	// 2. 调用DAO层进行查询
	users, err := dao.GetUsers(req.Username, req.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "查询用户列表失败")
		return
	}

	// 3. 将用户列表转换为DTO列表
	var userDTOs []dto.UserBriefDTO
	for _, user := range users {
		userDTOs = append(userDTOs, buildUserBriefDTO(user))
	}

	// 4. 返回成功的响应
	response.Success(c, gin.H{"data": userDTOs}, "获取用户列表成功")
}
