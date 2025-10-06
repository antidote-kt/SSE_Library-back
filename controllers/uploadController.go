package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

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

// 生成安全的文件名
func generateSecureFilename(originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d%s", timestamp, ext)
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

func UploadFile(c *gin.Context) {
	// 1. 获取上传的文件
	file, header, err := c.Request.FormFile("file") // "file" 是前端表单字段名
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "未上传文件",
		})
		return
	}
	defer file.Close()

	// 2. 验证文件类型
	if !validateFileType(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的文件类型",
		})
		return
	}

	// 3. 验证文件大小
	if !validateFileSize(header.Size) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件大小超出限制或文件为空",
		})
		return
	}

	// 4. 生成安全的文件名
	secureFilename := generateSecureFilename(header.Filename)

	// 5. 上传到 COS
	fileURL, err := utils.UploadFile(secureFilename, file)
	if err != nil {
		// 不暴露内部错误详情
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "文件上传失败，请稍后重试",
		})
		return
	}

	// 6. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "上传成功",
		"data": gin.H{
			"key":           secureFilename,
			"url":           fileURL,
			"original_name": header.Filename,
		},
	})
}
