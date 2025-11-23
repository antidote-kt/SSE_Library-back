package dto

// ModifyCategoryDTO 定义了修改分类或课程时需要绑定的数据
type ModifyCategoryDTO struct {
	ID          *uint64 `form:"id,omitempty"`          // 分类ID（必需，用于定位要修改的分类）
	Name        *string `form:"name,omitempty"`        // 分类名称（可选）
	Description *string `form:"description,omitempty"` // 分类描述（可选）
	IsCourse    *string `form:"isCourse,omitempty"`    // 是否是课程（可选，string类型，需要转换为bool）
	ParentID    *uint64 `form:"parentId,omitempty"`    // 父分类ID（可选）
}
