package dto

// UpdateUserStatusDTO 定义了管理员修改用户状态时绑定的Query参数
type UpdateUserStatusDTO struct {
	UserID uint64 `form:"userId" binding:"required"`
	Status string `form:"status" binding:"required"`
}

// SearchUsersDTO 定义了管理员搜索用户时绑定的Query参数
type SearchUsersDTO struct {
	UserID   *uint64 `form:"userId, omitempty"` // 使用指针以区分 "0" 和 "未提供"
	Username *string `form:"username, omitempty"`
}

//当使用 form 标签时：
//支持同时处理 GET 的查询参数和 POST 的 x-www-form-urlencoded 数据
//omitempty 可与 form 配合使用，表示该参数为可选参数
//指针类型能更好地区分 "未传值" 和 "零值" 的情况
