package controllers

import (
	"errors"
	"fmt"
	"net/http"

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

// UploadDocument 文档上传控制器 - 处理文档上传的主函数
// 该函数执行以下步骤：
// 1. 解析并验证上传请求参数
// 2. 上传主文件和封面图片
// 3. 将文档信息保存到数据库（使用事务）
// 4. 返回上传结果
func UploadDocument(c *gin.Context) {
	// 获取数据库连接
	db := config.GetDB()

	// 初始化上传请求结构体
	var req dto.UploadDTO

	// 1. 绑定并验证请求参数
	if err := bindAndValidateRequest(c, &req); err != nil {
		// 参数验证失败，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, err.Error())
		return
	}

	// 查询文档分类信息
	category, err := dao.GetCategoryByID(req.CategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 分类不存在，返回错误响应
			response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
		}
		// 数据库操作错误，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 用于存储文件URL的变量
	var fileURL string

	// 2. 上传主文件（如果有）
	if req.File != nil {
		// 使用工具函数上传主文件
		fileURL, err = utils.UploadMainFile(req.File, category.Name)
		if err != nil {
			// 文件上传失败，返回错误响应
			response.Fail(c, http.StatusInternalServerError, nil, err.Error())
			return
		}
	} else if req.VideoURL != nil {
		fileURL = *req.VideoURL
	}

	// 用于存储封面图片URL的变量
	var coverURL string

	// 3. 上传封面图片（如果有）
	if req.Cover != nil {
		// 使用工具函数上传封面图片
		coverURL, err = utils.UploadCoverImage(req.Cover, category.Name)
		if err != nil {
			// 封面上传失败，返回错误响应
			response.Fail(c, http.StatusInternalServerError, nil, err.Error())
			return
		}
	}

	// 4. 保存文档信息到数据库（使用事务）
	// 查询上传者是否存在
	uploader, err := dao.GetUserByID(req.UploaderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 上传者不存在，返回错误响应
			response.Fail(c, http.StatusNotFound, nil, constant.UploaderNotExist)
			return
		}
		// 数据库操作错误，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构建文档对象
	document := models.Document{
		Type:        req.Type,                       // 文档类型
		Name:        req.Name,                       // 文档名称
		UploaderID:  uploader.ID,                    // 上传者ID
		CategoryID:  category.ID,                    // 分类ID
		Status:      constant.DocumentStatusPending, // 文档状态（默认为审核中）
		URL:         fileURL,                        // 文件URL
		ReadCounts:  0,                              // 阅读次数（初始为0）
		Collections: 0,                              // 收藏次数（初始为0）
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
		// 如果没有提供作者，默认使用常量中的默认作者
		document.Author = constant.DefaultAuthor
	}

	if req.Introduction != nil {
		document.Introduction = *req.Introduction
	}

	if req.CreateYear != nil {
		document.CreateYear = *req.CreateYear
	}
	
	// 使用数据库事务创建文档
	err = db.Transaction(func(tx *gorm.DB) error {
		// 使用事务创建文档记录
		document, err = dao.CreateDocumentWithTx(tx, document, req.Tags)
		if err != nil {
			return err
		}
		// 如果没有返回错误，事务将自动提交
		return nil
	})

	// 检查事务执行结果
	if err != nil {
		// 文档创建失败，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DocumentCreateFail)
		return
	}

	// 5. 返回成功响应
	docDetailResponse, err := response.BuildDocumentDetailResponse(document)
	if err != nil {
		// 如果构建响应失败，仍返回基本成功信息
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 返回上传成功的响应
	response.SuccessWithData(c, docDetailResponse, constant.DocumentCreateSuccess)
}

// validateFileSize 验证文件大小是否符合要求
// 参数: size: 文件大小（字节）
// 返回值: bool: 是否符合要求
func validateFileSize(size int64) bool {
	// 检查文件大小是否在允许范围内（大于0且不超过最大限制）
	return size > 0 && size <= constant.MaxFileSize
}

// bindAndValidateRequest 绑定并验证请求参数
func bindAndValidateRequest(c *gin.Context, req *dto.UploadDTO) error {
	// 使用ShouldBind解析请求参数
	if err := c.ShouldBind(req); err != nil {
		// 参数解析错误
		return errors.New(constant.ParamParseError)
	}

	// 如果没有文件，直接返回
	if req.File == nil {
		return nil
	}

	// 验证文件大小
	if !validateFileSize(req.File.Size) {
		// 文件太大，返回错误信息
		return fmt.Errorf("文件大小不能超过%vMB", constant.MaxFileSize/1024/1024)
	}

	return nil
}
