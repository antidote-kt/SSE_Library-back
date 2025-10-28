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

// UploadFile 处理文件上传的主函数
func UploadDocument(c *gin.Context) {
	db := config.GetDB()
	var req dto.UploadDTO
	// 1. 绑定并验证请求参数
	if err := bindAndValidateRequest(c, &req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, err.Error())
		return
	}
	category, err := dao.GetCategoryByID(req.CategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
	}
	var fileURL string
	// 2. 上传主文件
	if req.File != nil {
		fileURL, err = utils.UploadMainFile(req.File, category.Name)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.FileUploadFailed)
			return
		}
	}

	var coverURL string
	// 3. 上传封面图片（如果有）
	if req.Cover != nil {
		coverURL, err = utils.UploadCoverImage(req.Cover, category.Name)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, err.Error())
			return
		}
	}

	// 4.保存文档信息到数据库（使用事务）
	// 查询上传者是否存在
	uploader, err := dao.GetUserByID(req.UploaderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UploaderNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}
	document := models.Document{
		Type:        req.Type,
		Name:        req.Name,
		UploaderID:  uploader.ID,
		CategoryID:  category.ID,
		Status:      constant.DocumentStatusAudit,
		URL:         fileURL,
		ReadCounts:  0,
		Collections: 0,
	}
	if req.File != nil {
		document.URL = fileURL
	} else if req.VideoURL != nil {
		document.URL = *req.VideoURL
	}
	// 处理可选字段
	if req.ISBN != nil {
		document.BookISBN = *req.ISBN
	}
	if req.Cover != nil {
		document.Cover = coverURL
	}

	if req.Author != nil {
		document.Author = *req.Author
	} else {
		document.Author = constant.DefaultAuthor
	}

	if req.Introduction != nil {
		document.Introduction = *req.Introduction
	}

	if req.CreateYear != nil {
		document.CreateYear = *req.CreateYear
	}
	if req.UploadTime != nil {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", *req.UploadTime)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, nil, constant.TimeFormatError)
			return
		}
		document.CreatedAt = parsedTime
	}
	// 事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建文档记录
		document, err = dao.CreateDocumentWithTx(tx, document, req.Tags)
		if err != nil {
			return err
		}
		// 如果没有返回错误，事务将自动提交
		return nil
	})
	// 检查事务执行结果
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DocumentCreateFail)
		return
	}

	// 5. 返回成功响应
	// 由于刚创建的文档可能还没有标签，我们先使用BuildDocumentDetailResponse创建基础结构
	docDetailResponse, err := response.BuildDocumentDetailResponse(document)
	if err != nil {
		// 如果构建响应失败，仍返回基本成功信息
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 更新响应中的标签为上传时提供的标签，因为刚创建的文档可能还没有从数据库获取到标签
	docDetailResponse.Tags = req.Tags

	response.SuccessWithData(c, docDetailResponse, constant.DocumentCreateSuccess)
}

// 验证文件大小
func validateFileSize(size int64) bool {
	return size > 0 && size <= constant.MaxFileSize
}

// 绑定并验证请求参数
func bindAndValidateRequest(c *gin.Context, req *dto.UploadDTO) error {
	if err := c.ShouldBind(req); err != nil {
		return errors.New(constant.ParamParseError)
	}
	if req.File == nil {
		return nil
	}
	if !validateFileSize(req.File.Size) {
		return fmt.Errorf("文件大小不能超过%vMB", constant.MaxFileSize/1024/1024)
	}

	return nil
}
