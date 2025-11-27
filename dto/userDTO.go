package dto

import "mime/multipart"

// RegisterDTO 定义了用户注册时需要绑定的数据
type RegisterDTO struct {
	Email            string                `form:"email" binding:"required,email"`
	Username         string                `form:"username" binding:"required,min=3,max=20"`
	Avatar           *multipart.FileHeader `form:"userAvatar,omitempty"`
	Password         string                `form:"password" binding:"required,min=6,max=20"`
	VerificationCode string                `form:"Code" binding:"required,min=6,max=6"`
}

// LoginDTO 定义了用户登录时需要绑定的数据
type LoginDTO struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordDTO 定义了用户修改密码时需要绑定的数据
type ChangePasswordDTO struct {
	Email            string `json:"email" binding:"required,email"`
	NewPassword      string `json:"newPassword" binding:"required,min=6,max=20"`
	VerificationCode string `json:"Code" binding:"required,min=6,max=6"`
}

// SendCodeDTO 定义了请求发送验证码时需要绑定的数据
type SendCodeDTO struct {
	Email string `json:"email" binding:"required,email"`
	Usage string `json:"usage" binding:"required"` // 例如: "reset-password", "register"
}

// ModifyInfoDTO 定义了用户修改个人资料时需要绑定的数据（PS：查看个人主页请求参数只有路径参数，无需专门结构体解析）
type ModifyInfoDTO struct {
	UserName   *string               `form:"userName,omitempty"`
	UserAvatar *multipart.FileHeader `form:"userAvatar,omitempty"`
	Email      *string               `form:"email,omitempty"`
	//这里omitempty的限制表示如果前端传来的form参数为空，则不进行参数绑定，也就不会自动创建空值变量而影响controller的指针判空(该法适用于可选参数)
}

// UserBriefDTO 定义了用户基本信息
type UserBriefDTO struct {
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	UserAvatar string `json:"userAvatar"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}
