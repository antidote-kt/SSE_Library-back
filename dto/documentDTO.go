package dto

import "mime/multipart"

type UploadDTO struct {
	// 要上传的文件
	File *multipart.FileHeader `form:"file" binding:"required"`
	// 封面图片
	Cover *multipart.FileHeader `form:"cover,omitempty"`
	// 分类名称
	Category string `form:"category" binding:"required"`
	CourseID uint64 `form:"courseId" binding:"required"`
	// 上传的资料类型
	Type         string  `form:"type" binding:"required"`
	Name         string  `form:"name" binding:"required"`
	ISBN         *string `form:"ISBN,omitempty"`
	Introduction *string `form:"introduction,omitempty"`
	// 关键词
	Tags []string `form:"tags,omitempty"`
	// 作者
	AuthorName *string `form:"authorName,omitempty"`
	// 创作年份
	CreateYear *string `form:"createYear,omitempty"`
	// 上传者
	UploaderName   string `form:"uploaderName" binding:"required"`
	UploaderID     uint64 `form:"uploaderId" binding:"required"`
	UploaderAvatar string `form:"uploaderAvatar" binding:"required"`
	// 上传时间
	UploadTime *string `form:"uploadTime,omitempty"`
}
type WithdrawUploadDTO struct {
	DocumentID uint64  `form:"document_id" binding:"required"`
	UserID     *uint64 `form:"userId"`
}

type ModifyDocumentDTO struct {
	Author     *string               `form:"author,omitempty"`
	Category   *string               `form:"category,omitempty"`
	File       *multipart.FileHeader `form:"file,omitempty"`
	Cover      *multipart.FileHeader `form:"cover,omitempty"`
	CreateYear *string               `form:"createYear,omitempty"`
	// 资料id
	DocumentID uint64   `form:"document_id" binding:"required"`
	ISBN       *string  `form:"ISBN"`
	Name       *string  `form:"name,omitempty"`
	Tags       []string `form:"tags,omitempty"`
	Type       *string  `form:"type,omitempty"`
	UploadTime *string  `form:"uploadTime,omitempty"`
}
type AdminModifyDocumentDTO struct {
	Author     *string               `form:"author,omitempty"`
	Category   *string               `form:"category,omitempty"`
	File       *multipart.FileHeader `form:"file,omitempty"`
	Cover      *multipart.FileHeader `form:"cover,omitempty"`
	CreateYear *string               `form:"createYear,omitempty"`
	// 资料id
	DocumentID uint64   `form:"document_id" binding:"required"`
	ISBN       *string  `form:"ISBN"`
	Name       *string  `form:"name,omitempty"`
	Tags       []string `form:"tags,omitempty"`
	Type       *string  `form:"type,omitempty"`
	UploadTime *string  `form:"uploadTime,omitempty"`
}
type AdminModifyDocumentStatusRequest struct {
	DocumentID uint64  `json:"document_id"`
	Name       *string `json:"Name,omitempty"`
	NewStatus  string  `json:"newStatus"`
	Type       string  `json:"type"`
}
