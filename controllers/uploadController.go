package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 最大文件大小 (10MB)
const maxFileSize = 10 << 20

// UploadFile 处理文件上传的主函数
func UploadFile(c *gin.Context) {
	db := config.GetDB()
	var req dto.UploadDTO
	// 1. 绑定并验证请求参数
	if err := bindAndValidateRequest(c, &req); err != nil {
		return
	}

	// 2. 上传主文件
	fileURL, err := utils.UploadMainFile(req.File, req.Category)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "上传文件失败"+err.Error())
		return
	}

	// 3. 上传封面图片（如果有）
	coverURL, err := utils.UploadCoverImage(req.Cover, req.Category)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "上传封面失败"+err.Error())
		return
	}

	// 4.保存文档信息到数据库（使用事务）
	// 查询分类是否存在
	categories, err := dao.GetCategoryByName(req.Category)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误"+err.Error())
	}
	if len(categories) == 0 {
		response.Fail(c, http.StatusNotFound, nil, "分类不存在")
	}
	category := categories[0]
	// 查询上传者是否存在
	uploader, err := dao.GetUserByID(req.UploaderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "上传用户不存在")
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误"+err.Error())
	}
	// 查询课程是否存在
	course, err := dao.GetCourseByID(req.CourseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "课程不存在")
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误"+err.Error())
	}
	document := models.Document{
		Type:        req.Type,
		Name:        req.Name,
		UploaderID:  uploader.ID,
		CategoryID:  category.ID,
		CourseID:    course.ID,
		Status:      constant.DocumentStatusAudit,
		URL:         fileURL,
		ReadCounts:  0,
		Collections: 0,
	}
	// 处理可选字段
	if req.ISBN != nil {
		document.BookISBN = *req.ISBN
	}
	if coverURL != "" {
		document.Cover = coverURL
	}

	if req.AuthorName != nil {
		document.Author = *req.AuthorName
	} else {
		document.Author = "佚名"
	}

	if req.Introduction != nil {
		document.Introduction = *req.Introduction
	}

	if req.CreateYear != nil {
		document.CreateYear = *req.CreateYear
	}
	if req.UploadTime != nil {
		document.CreatedAt, err = time.Parse("2006-01-02 15:04:05", *req.UploadTime)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, nil, "时间格式错误")
		}
	}
	// 事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建文档记录
		if err := dao.CreateDocumentWithTx(tx, document, req.Tags); err != nil {
			return err
		}
		// 如果没有返回错误，事务将自动提交
		return nil
	})
	// 检查事务执行结果
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
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
			"URL":         utils.GetFileURL(document.URL),
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
		"Cover":        utils.GetFileURL(document.Cover),
		"tags":         req.Tags,
		"introduction": document.Introduction,
		"createYear":   document.CreateYear,
	}
	response.Success(c, responseData, "上传成功")
}

// 验证文件大小
func validateFileSize(size int64) bool {
	return size > 0 && size <= maxFileSize
}

// 绑定并验证请求参数
func bindAndValidateRequest(c *gin.Context, req *dto.UploadDTO) error {
	if err := c.ShouldBind(req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数绑定失败: "+err.Error())
		return err
	}

	if !validateFileSize(req.File.Size) {
		response.Fail(c, http.StatusBadRequest, nil, "文件大小超出限制或文件为空")
		return fmt.Errorf("文件大小超出限制或文件为空")
	}

	return nil
}
