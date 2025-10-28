package response

import (
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/utils"
)

type InfoBriefResponse struct {
	Name        string `json:"name"`
	DocumentID  uint64 `json:"documentId"`
	Type        string `json:"type"`
	UploadTime  string `json:"uploadTime"`
	Status      string `json:"status"`
	Category    string `json:"category"`
	Collections int    `json:"collections"`
	ReadCounts  int    `json:"readCounts"`
	URL         string `json:"URL"`
}

type UploaderResponse struct {
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	UserAvatar string `json:"userAvatar"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

type DocumentDetailResponse struct {
	InfoBrief    InfoBriefResponse `json:"infoBrief"`
	BookISBN     string            `json:"bookISBN"`
	Author       string            `json:"author"`
	Uploader     UploaderResponse  `json:"uploader"`
	Cover        string            `json:"Cover"`
	Tags         []string          `json:"tags"`
	Introduction string            `json:"introduction"`
	CreateYear   string            `json:"createYear"`
}

// buildDocumentDetailResponse 构建文档详情响应对象
func BuildDocumentDetailResponse(document models.Document) (DocumentDetailResponse, error) {
	// 获取上传者信息
	uploader, err := dao.GetUserByID(document.UploaderID)
	if err != nil {
		return DocumentDetailResponse{}, err
	}

	// 获取文档标签
	tags, err := dao.GetDocumentTagByDocumentID(document.ID)
	if err != nil {
		return DocumentDetailResponse{}, err
	}

	// 转换标签为字符串数组
	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.TagName
	}

	// 获取分类信息
	category, err := dao.GetCategoryByID(document.CategoryID)
	if err != nil {
		return DocumentDetailResponse{}, err
	}

	// 构建 InfoBriefResponse
	infoBrief := InfoBriefResponse{
		Name:        document.Name,
		DocumentID:  document.ID,
		Type:        document.Type,
		UploadTime:  document.CreatedAt.Format("2006-01-02 15:04:05"),
		Status:      document.Status,
		Category:    category.Name,
		Collections: document.Collections,
		ReadCounts:  document.ReadCounts,
		URL:         utils.GetResponseFileURL(document),
	}

	// 构建 UploaderResponse
	uploaderResponse := UploaderResponse{
		UserID:     uploader.ID,
		Username:   uploader.Username,
		UserAvatar: uploader.Avatar,
		Status:     uploader.Status,
		CreateTime: uploader.CreatedAt.Format("2006-01-02 15:04:05"),
		Email:      uploader.Email,
		Role:       uploader.Role,
	}

	// 构建 DocumentDetailResponse
	docDetailResponse := DocumentDetailResponse{
		InfoBrief:    infoBrief,
		BookISBN:     document.BookISBN,
		Author:       document.Author,
		Uploader:     uploaderResponse,
		Cover:        utils.GetFileURL(document.Cover),
		Tags:         tagNames,
		Introduction: document.Introduction,
		CreateYear:   document.CreateYear,
	}

	return docDetailResponse, nil
}
func BuildInfoBriefResponse(document models.Document) (InfoBriefResponse, error) {
	category, err := dao.GetCategoryByID(document.CategoryID)
	if err != nil {
		return InfoBriefResponse{}, err
	}
	infoBriefResponse := InfoBriefResponse{
		Name:        document.Name,
		DocumentID:  document.ID,
		Type:        document.Type,
		UploadTime:  document.CreatedAt.Format("2006-01-02 15:04:05"),
		Status:      document.Status,
		Category:    category.Name,
		Collections: document.Collections,
		ReadCounts:  document.ReadCounts,
		URL:         utils.GetResponseFileURL(document),
	}

	return infoBriefResponse, nil
}
