package dto

// UpdateUserStatusDTO 定义了管理员修改用户状态时绑定的Query参数
type UpdateUserStatusDTO struct {
	UserID uint64 `form:"user_id" binding:"required"`
	Status string `form:"status" binding:"required"`
}

// SearchUsersDTO 定义了管理员搜索用户时绑定的Query参数
type SearchUsersDTO struct {
	UserID   *uint64 `form:"user_id"` // 使用指针以区分 "0" 和 "未提供"
	Username *string `form:"username"`
}
