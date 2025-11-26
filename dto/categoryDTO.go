package dto

// AddCategoryDTO 添加分类接口的请求参数
type AddCategoryDTO struct {
	IsCourse *bool `json:"isCourse" binding:"required"`
	// 当前端没有传值时，gin框架会为bool类型赋默认值false，与传了false值会混淆，故使用指针区分（若没传值，*IsCourse还是false，但IsCourse为nil）
	ParentCatID *uint64 `json:"parentCatId"`
	// 此处同理，没传值时*ParentCatID默认为0，但ParentCatID为nil
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description"`
}

// ModifyCategoryDTO 定义了修改分类或课程时需要绑定的数据
type ModifyCategoryDTO struct {
	ID          *uint64 `form:"id,omitempty"`          // 分类ID（必需，用于定位要修改的分类）
	Name        *string `form:"name,omitempty"`        // 分类名称（可选）
	Description *string `form:"description,omitempty"` // 分类描述（可选）
	IsCourse    *string `form:"isCourse,omitempty"`    // 是否是课程（可选，string类型，需要转换为bool）
	ParentID    *uint64 `form:"parentId,omitempty"`    // 父分类ID（可选）

}
