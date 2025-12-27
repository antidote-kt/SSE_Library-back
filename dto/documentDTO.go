package dto

import "mime/multipart"

type UploadDTO struct {
	// 要上传的文件
	File *multipart.FileHeader `form:"file,omitempty"`
	// 封面图片
	Cover *multipart.FileHeader `form:"cover,omitempty"`
	// 分类名称
	CategoryID uint64 `form:"categoryId" binding:"required"`
	// 上传的资料类型
	Type         string  `form:"type" binding:"required"`
	Name         string  `form:"name" binding:"required"`
	ISBN         *string `form:"ISBN,omitempty"`
	Introduction *string `form:"introduction,omitempty"`
	// 关键词
	Tags     []string `form:"tags,omitempty"`
	VideoURL *string  `form:"videoURL,omitempty"`
	// 作者
	Author *string `form:"author,omitempty"`
	// 创作年份
	CreateYear *string `form:"createYear,omitempty"`
	// 上传者
	UploaderID uint64 `form:"uploaderId" binding:"required"`
}
type WithdrawUploadDTO struct {
	DocumentID uint64 `form:"documentId" binding:"required"`
	UserID     uint64 `form:"userId" binding:"required"`
}

type ModifyDocumentDTO struct {
	Author     *string               `form:"author,omitempty"`
	CategoryID *uint64               `form:"categoryId,omitempty"`
	File       *multipart.FileHeader `form:"file,omitempty"`
	Cover      *multipart.FileHeader `form:"cover,omitempty"`
	VideoURL   *string               `form:"videoURL,omitempty"`
	CreateYear *string               `form:"createYear,omitempty"`
	// 资料id
	DocumentID   uint64   `form:"documentId" binding:"required"`
	ISBN         *string  `form:"ISBN"`
	Name         *string  `form:"name,omitempty"`
	Tags         []string `form:"tags,omitempty"`
	Type         *string  `form:"type,omitempty"`
	Introduction *string  `form:"introduction,omitempty"`
}
type SearchDocumentDTO struct {
	// 筛选科目
	CategoryID *uint64 `form:"categoryId,omitempty"`
	// 搜索关键词
	Key *string `form:"key,omitempty"`
	// 关键词的类型
	TypeOfKey *string `form:"typeOfKey,omitempty"`
	// 筛选文件类型
	Type *string `form:"type,omitempty"`
	// 筛选创作时间
	Year *string `form:"year,omitempty"`
}
type AdminModifyDocumentStatusRequest struct {
	DocumentID uint64  `json:"documentId"`
	Status     *string `json:"status"`
}
type DocumentBriefDTO struct {
	Name        string `json:"name"`
	DocumentID  uint64 `json:"documentId"`
	Type        string `json:"type"`
	UploadTime  string `json:"uploadTime"`
	Status      string `json:"status"`
	Category    string `json:"category"`
	Course      string `json:"course"`
	Collections int    `json:"collections"`
	ReadCounts  int    `json:"readCounts"`
	URL         string `json:"URL"`
	Content     string `json:"content"`
	CreateTime  string `json:"createTime"`
}

// GetDocumentListDTO 获取文档列表请求参数
type GetDocumentListDTO struct {
	IsSuggest  *bool   `form:"isSuggest"`  // 是否为推荐模式
	CategoryID *uint64 `form:"categoryId"` // 分类ID (可选，空则返回全部)
}
