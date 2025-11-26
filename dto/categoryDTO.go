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
