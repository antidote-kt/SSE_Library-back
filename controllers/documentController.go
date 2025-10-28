package controllers

import (
	"errors"
	"net/http"
	"strconv"

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

// 修改文档
func ModifyDocument(c *gin.Context) {
	var request dto.ModifyDocumentDTO
	var category models.Category
	//记录旧状态
	var oldType string
	db := config.GetDB()
	if err := c.ShouldBind(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 查找要修改的文档
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}
	oldType = document.Type

	// 更新文档字段（如果提供了相应字段）
	if request.Author != nil {
		document.Author = *request.Author
	}
	if request.CategoryID != nil {
		category, err = dao.GetCategoryByID(*request.CategoryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
		document.CategoryID = category.ID
	}
	if request.CreateYear != nil {
		document.CreateYear = *request.CreateYear
	}
	if request.ISBN != nil {
		document.BookISBN = *request.ISBN
	}
	if request.Name != nil {
		document.Name = *request.Name
	}
	if request.Type != nil {
		document.Type = *request.Type
	}
	if request.Introduction != nil {
		document.Introduction = *request.Introduction
	}
	if request.Cover != nil {
		// 先删除旧封面，如果document.cover为空串则不做任何操作
		err := utils.DeleteFile(document.Cover)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.OldCoverDeleteFailed)
			return
		}
		//上传新封面
		document.Cover, err = utils.UploadCoverImage(request.Cover, category.Name)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.CoverUploadFailed)
			return
		}
	}
	if request.File != nil || request.VideoURL != nil {
		// 只有原来的类型不是video才删除，如果原来是video，cos没有相应的资源
		if oldType != constant.VideoType {
			err := utils.DeleteFile(document.URL)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, nil, constant.OldFileDeleteFailed)
				return
			}
		}
		if request.VideoURL != nil {
			document.URL = *request.VideoURL
		} else if request.File != nil {
			document.URL, err = utils.UploadMainFile(request.File, category.Name)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, nil, constant.FileUploadFailed)
				return
			}
		}
		// 重新上传文件需要重新审核
		document.Status = constant.DocumentStatusAudit
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := dao.UpdateDocumentWithTx(tx, document); err != nil {
			return err
		}
		if len(request.Tags) > 0 {
			// 删除原有的标签映射
			if err := dao.DeleteDocumentTagByDocumentIDWithTx(tx, document.ID); err != nil {
				return errors.New(constant.OldTagDeleteFailed)
			}
			// 创建新标签映射
			if err := dao.CreateDocumentTagWithTx(tx, document.ID, request.Tags); err != nil {
				return err
			}

		}
		return nil
	})
	// 检查事务执行结果
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}
	_, err = dao.GetUserByID(document.UploaderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UploaderNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	response.Success(c, nil, constant.DocumentUpdateSuccess)
}

// 获取文档详情
func GetDocumentByID(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.ParamParseError)
		return
	}
	document, err := dao.GetDocumentByID(documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构建文档详情响应
	docDetailResponse, err := response.BuildDocumentDetailResponse(document)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	response.SuccessWithData(c, docDetailResponse, constant.DocumentObtain)
}

func SearchDocument(c *gin.Context) {
	var request dto.SearchDocumentDTO
	if err := c.ShouldBind(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}
	if request.CategoryID != nil {
		_, err := dao.GetCategoryByID(*request.CategoryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
	}

	// 使用DAO层进行搜索
	documents, err := dao.SearchDocumentsByParams(request)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构建响应数据
	var results []response.DocumentDetailResponse
	for _, document := range documents {
		docDetailResponse, err := response.BuildDocumentDetailResponse(document)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
		results = append(results, docDetailResponse)
	}

	response.SuccessWithData(c, results, constant.DocumentObtain)
}
