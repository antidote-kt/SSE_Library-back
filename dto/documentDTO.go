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
	// 上传时间
	UploadTime *string `form:"uploadTime,omitempty"`
}
type WithdrawUploadDTO struct {
	DocumentID uint64 `form:"document_id" binding:"required"`
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
	UploadTime   *string  `form:"uploadTime,omitempty"`
}
type AdminModifyDocumentDTO struct {
	Author     *string               `form:"author,omitempty"`
	CategoryID *uint64               `form:"categoryId,omitempty"`
	File       *multipart.FileHeader `form:"file,omitempty"`
	Cover      *multipart.FileHeader `form:"cover,omitempty"`
	VideoURL   *string               `form:"videoURL,omitempty"`
	CreateYear *string               `form:"createYear,omitempty"`
	// 资料id
	DocumentID   uint64   `form:"document_id" binding:"required"`
	ISBN         *string  `form:"ISBN"`
	Name         *string  `form:"name,omitempty"`
	Tags         []string `form:"tags,omitempty"`
	Type         *string  `form:"type,omitempty"`
	Introduction *string  `form:"introduction,omitempty"`
	UploadTime   *string  `form:"uploadTime,omitempty"`
}
type AdminModifyDocumentStatusRequest struct {
	DocumentID uint64  `form:"document_id"`
	Name       *string `form:"name,omitempty"`
	Status     *string `form:"status"`
	Type       *string `form:"type"`
}
