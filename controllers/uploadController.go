package controllers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

type UploadRequest struct {
	// 要上传的文件
	File *multipart.FileHeader `form:"file" binding:"required"`
	// 封面图片
	Cover *multipart.FileHeader `form:"cover,omitempty"`
	// 书籍名称
	Category string `form:"category" binding:"required"`
	// 上传的资料类型
	Type string  `form:"type" binding:"required"`
	Name string  `form:"name" binding:"required"`
	ISBN *string `form:"ISBN,omitempty"`
	// 关键词
	Tags []string `form:"tags,omitempty"`
	// 作者
	AuthorName *string `form:"authorName,omitempty"`
	// 创作年份
	CreateYear *string `form:"createYear,omitempty"`
	// 上传者
	UploaderName   string `form:"uploaderName" binding:"required"`
	UploaderID     string `form:"uploaderId" binding:"required"`
	UploaderAvatar string `form:"uploaderAvatar" binding:"required"`
	// 上传时间
	UploadTime *string `form:"uploadTime,omitempty"`
}

// 允许的文件类型
var allowedFileTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".txt":  true,
}

// 最大文件大小 (10MB)
const maxFileSize = 10 << 20

// 文档状态常量
const (
	DocumentStatusAudit    = "audit"    // 审核中
	DocumentStatusOpen     = "open"     // 已通过
	DocumentStatusRejected = "rejected" // 已拒绝
	DocumentStatusDeleted  = "deleted"  // 已删除
)

// UploadFile 处理文件上传的主函数
func UploadFile(c *gin.Context) {
	var req UploadRequest

	// 1. 绑定并验证请求参数
	if err := bindAndValidateRequest(c, &req); err != nil {
		return
	}

	// 2. 上传主文件
	fileURL, err := uploadMainFile(req.File, req.Category)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, gin.H{}, err.Error())
		return
	}

	// 3. 上传封面图片（如果有）
	coverURL, err := uploadCoverImage(req.Cover, req.Category)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, gin.H{}, err.Error())
		return
	}

	// 4. 保存文档信息到数据库
	db := config.GetDB()

	// 查找分类
	var category models.Category
	if err := db.Where("name = ?", req.Category).First(&category).Error; err != nil {
		response.Fail(c, http.StatusBadRequest, gin.H{}, "分类不存在: "+req.Category)
		return
	}

	// 转换上传者ID
	uploaderID, err := strconv.ParseUint(req.UploaderID, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, gin.H{}, "上传者ID格式错误")
		return
	}

	// 查询上传者信息
	var uploader models.User
	if err := db.Where("id = ?", uploaderID).First(&uploader).Error; err != nil {
		response.Fail(c, http.StatusBadRequest, gin.H{}, "上传者不存在")
		return
	}

	// 创建文档记录
	document := models.Document{
		Type:     req.Type,
		Name:     req.Name,
		BookISBN: getStringValue(req.ISBN),
		Author: func() string {
			author := getStringValue(req.AuthorName)
			if author == "" {
				return "佚名"
			}
			return author
		}(),
		UploaderID:   uploaderID,
		CourseID:     category.ID, // 使用分类ID作为CourseID
		Cover:        coverURL,
		Introduction: "", // 可以后续添加
		CreateYear:   getStringValue(req.CreateYear),
		Status:       DocumentStatusAudit, // 默认状态为审核中
		ReadCounts:   0,
		Collections:  0,
		URL:          fileURL,
	}

	if err := db.Create(&document).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, gin.H{}, "保存文档信息失败")
		return
	}

	// 保存标签
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			// 查找或创建标签
			var tag models.DocumentTag
			if err := db.Where("tag_name = ?", tagName).First(&tag).Error; err != nil {
				// 标签不存在，创建新标签
				tag = models.DocumentTag{
					TagName: tagName,
				}
				if err := db.Create(&tag).Error; err != nil {
					continue // 跳过创建失败的标签
				}
			}

			// 创建文档标签关联
			tagMap := models.DocumentTagMap{
				DocumentID: document.ID,
				TagID:      tag.ID,
			}
			if err := db.Create(&tagMap).Error; err != nil {
				continue // 跳过创建失败的关联
			}
		}
	}

	// 5. 返回成功响应
	responseData := gin.H{
		"infoBrief": gin.H{
			"name":        document.Name,
			"document_id": document.ID,
			"type":        document.Type,
			"uploadTime":  document.CreatedAt.Format("2006-01-02 15:04:05"),
			"status":      document.Status,
			"category":    category.Name,
			"collections": document.Collections,
			"readCounts":  document.ReadCounts,
			"URL":         document.URL,
		},
		"bookISBN": document.BookISBN,
		"author":   document.Author,
		"uploader": gin.H{
			"userId":     uploader.ID,
			"username":   uploader.Username,
			"userAvatar": uploader.Avatar,
			"status":     uploader.Status,
			"createTime": uploader.CreatedAt.Format("2006-01-02 15:04:05"),
			"email":      uploader.Email,
			"role":       uploader.Role,
		},
		"Cover":        document.Cover,
		"tags":         req.Tags,
		"introduction": document.Introduction,
		"createYear":   document.CreateYear,
	}
	response.Success(c, responseData, "上传成功")
}

// 生成安全的文件名
func generateSecureFilename(originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d%s", timestamp, ext)
}

// 生成文件存储路径
func generateFilePath(category, originalName string) string {
	// 按分类和日期组织文件
	now := time.Now()
	datePath := now.Format("2006/01/02") // 年/月/日
	secureFilename := generateSecureFilename(originalName)

	// 路径格式: files/{category}/{year/month/day}/{filename}
	path := fmt.Sprintf("files/%s/%s/%s", category, datePath, secureFilename)
	return path
}

// 生成封面存储路径
func generateCoverPath(category, originalName string) string {
	now := time.Now()
	datePath := now.Format("2006/01/02")
	secureFilename := generateSecureFilename(originalName)

	// 路径格式: covers/{category}/{year/month/day}/{filename}
	path := fmt.Sprintf("covers/%s/%s/%s", category, datePath, secureFilename)
	return path
}

// 验证文件类型
func validateFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return allowedFileTypes[ext]
}

// 验证文件大小
func validateFileSize(size int64) bool {
	return size > 0 && size <= maxFileSize
}

// bindAndValidateRequest 绑定并验证请求参数
func bindAndValidateRequest(c *gin.Context, req *UploadRequest) error {
	if err := c.ShouldBind(req); err != nil {
		response.Fail(c, http.StatusBadRequest, gin.H{}, "参数绑定失败: "+err.Error())
		return err
	}

	// 验证主文件
	if !validateFileType(req.File.Filename) {
		response.Fail(c, http.StatusBadRequest, gin.H{}, "不支持的文件类型")
		return fmt.Errorf("不支持的文件类型")
	}

	if !validateFileSize(req.File.Size) {
		response.Fail(c, http.StatusBadRequest, gin.H{}, "文件大小超出限制或文件为空")
		return fmt.Errorf("文件大小超出限制或文件为空")
	}

	return nil
}

// uploadMainFile 上传主文件
func uploadMainFile(file *multipart.FileHeader, category string) (string, error) {
	fileReader, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败")
	}
	defer fileReader.Close()

	// 使用新的路径生成
	filePath := generateFilePath(category, file.Filename)
	fileURL, err := utils.UploadFile(filePath, fileReader)
	if err != nil {
		return "", fmt.Errorf("文件上传失败，请稍后重试")
	}

	return fileURL, nil
}

// uploadCoverImage 上传封面图片
func uploadCoverImage(cover *multipart.FileHeader, category string) (string, error) {
	if cover == nil {
		return "", nil // 没有封面图片，返回空字符串
	}

	if !validateFileType(cover.Filename) {
		return "", fmt.Errorf("封面图片格式不支持")
	}

	coverFile, err := cover.Open()
	if err != nil {
		return "", fmt.Errorf("打开封面文件失败")
	}
	defer coverFile.Close()

	// 使用新的路径生成
	coverPath := generateCoverPath(category, cover.Filename)
	coverURL, err := utils.UploadFile(coverPath, coverFile)
	if err != nil {
		return "", fmt.Errorf("封面图片上传失败")
	}

	return coverURL, nil
}

// 辅助函数：安全地获取字符串指针的值
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
