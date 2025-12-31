package controllers

import (
	"encoding/json"
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

// ModifyDocument 文档修改接口
func ModifyDocument(c *gin.Context) {
	var request dto.ModifyDocumentDTO
	// 声明分类模型用于后续获取分类信息
	var category models.Category
	// 记录旧状态，用于判断是否需要删除原始文件
	var oldType string
	// 获取数据库连接实例
	db := config.GetDB()

	// 解析并验证请求参数
	if err := c.ShouldBind(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 解析 tags JSON 字符串为 []string
	// 如果前端传了 tags 字段（即使是空数组 []），也需要处理标签更新
	var tags []string
	hasTagsField := false
	if request.Tags != "" {
		hasTagsField = true
		if err := json.Unmarshal([]byte(request.Tags), &tags); err != nil {
			response.Fail(c, http.StatusBadRequest, nil, "标签格式错误: "+err.Error())
			return
		}
	}

	// 根据文档ID查找要修改的文档
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		// 如果文档不存在，返回错误信息
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}
	// 记录修改前的文档类型，用于后续文件处理判断
	oldType = document.Type

	// 动态更新文档字段（仅更新客户端提供的字段）
	if request.Author != nil {
		document.Author = *request.Author
	}
	if request.CategoryID != nil {
		// 验证分类ID是否存在
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

	// 处理封面图片更新
	if request.Cover != nil {
		// 先删除旧封面，如果document.cover为空串则不做任何操作
		err := utils.DeleteFile(document.Cover)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.OldCoverDeleteFailed)
			return
		}
		// 上传新封面图片到指定分类目录
		document.Cover, err = utils.UploadCoverImage(request.Cover, category.Name)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, err.Error())
			return
		}
	}

	// 处理文档文件或视频链接更新
	if request.File != nil || request.VideoURL != nil {
		// 只有原来的类型不是video才删除，如果原来是video，cos没有相应的资源
		if oldType != constant.VideoType {
			// 删除旧文件
			err := utils.DeleteFile(document.URL)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, nil, constant.OldDocumentDeleteFailed)
				return
			}
		}

		// 更新文档URL：如果是视频URL则直接赋值，否则上传新文件
		if request.VideoURL != nil {
			document.URL = *request.VideoURL
		} else if request.File != nil {
			document.URL, err = utils.UploadMainFile(request.File, category.Name)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, nil, err.Error())
				return
			}
		}
		// 重新上传文件后需要重新审核
		document.Status = constant.DocumentStatusPending
	}

	// 使用事务确保数据一致性：更新文档信息和标签映射
	err = db.Transaction(func(tx *gorm.DB) error {
		// 在事务中更新文档信息
		if err := dao.UpdateDocumentWithTx(tx, document); err != nil {
			return err
		}

		// 如果请求中包含标签字段，则更新标签映射关系
		if hasTagsField {
			// 删除原有的标签映射关系
			if err := dao.DeleteDocumentTagByDocumentIDWithTx(tx, document.ID); err != nil {
				return errors.New(constant.OldTagDeleteFailed)
			}
			// 如果解析后的标签数组不为空，创建新的标签映射关系
			if len(tags) > 0 {
				if err := dao.CreateDocumentTagWithTx(tx, document.ID, tags); err != nil {
					return err
				}
			}
			// 如果 tags 为空数组，只删除不创建，实现清空标签的效果
		}
		return nil
	})

	// 检查事务执行结果
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 验证上传者是否存在
	_, err = dao.GetUserByID(document.UploaderID)
	if err != nil {
		// 如果上传者不存在，返回错误信息
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UploaderNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 返回成功响应
	response.Success(c, nil, constant.DocumentUpdateSuccess)
}

// GetDocumentByID 文档详情获取接口
func GetDocumentByID(c *gin.Context) {
	// 从URL参数中获取文档ID字符串
	documentIDStr := c.Param("id")
	// 将字符串格式的文档ID转换为uint64类型
	documentID, err := strconv.ParseUint(documentIDStr, 10, 64)
	if err != nil {
		// 如果转换失败，返回参数解析错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.ParamParseError)
		return
	}

	// 通过DAO层根据文档ID查询文档信息
	document, err := dao.GetDocumentByID(documentID)
	if err != nil {
		// 如果文档不存在，返回404错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		// 其他数据库错误，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}
	// 更新阅读量
	document.ReadCounts++
	if err := dao.UpdateDocument(document); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DocumentUpdateFail)
	}

	// 增加浏览次数 (异步执行)
	go func() {
		_ = dao.IncrementDocumentViewCount(documentID)
	}()

	// 记录浏览历史 (异步执行)
	// 从JWT解析用户信息
	if claims, exists := c.Get(constant.UserClaims); exists {
		userClaims := claims.(*utils.MyClaims)
		go func(uid uint64, sourceID uint64) {
			// 传入 "document" 类型
			_ = dao.AddViewHistory(uid, sourceID, "document")
		}(userClaims.UserID, documentID) // 回调函数实现异步
	}

	// 构建文档详情响应数据结构
	docDetailResponse, err := response.BuildDocumentDetailResponse(document)
	if err != nil {
		// 如果构建响应数据失败，返回数据库错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 返回成功响应，携带文档详情数据
	response.SuccessWithData(c, docDetailResponse, constant.DocumentObtain)
}

// SearchDocument 文档搜索接口
// 根据多种参数（如关键词、分类ID、作者等）搜索文档
// 支持多种搜索条件的组合查询
func SearchDocument(c *gin.Context) {
	// 声明搜索请求参数结构体
	var request dto.SearchDocumentDTO

	// 解析并验证请求参数
	if err := c.ShouldBind(&request); err != nil {
		// 如果参数解析失败，返回400错误
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 如果请求中包含分类ID参数，验证分类是否存在
	if request.CategoryID != nil {
		_, err := dao.GetCategoryByID(*request.CategoryID)
		if err != nil {
			// 如果分类不存在，返回404错误
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
				return
			}
			// 其他数据库错误，返回500错误
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
	}

	// 通过DAO层根据请求参数进行文档搜索
	documents, err := dao.SearchDocumentsByParams(request)
	if err != nil {
		// 如果搜索过程中发生错误，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构建搜索结果响应数据列表
	var results []response.DocumentDetailResponse
	// 遍历搜索到的文档列表，为每个文档构建详细响应信息
	for _, document := range documents {
		docDetailResponse, err := response.BuildDocumentDetailResponse(document)
		if err != nil {
			// 如果构建响应数据失败，返回500错误
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
		// 将构建好的文档详情响应添加到结果列表中
		results = append(results, docDetailResponse)
	}

	// 返回成功响应，携带搜索结果列表
	response.SuccessWithData(c, results, constant.DocumentObtain)
}

// GetDocumentList 获取文档列表接口 (支持推荐和分类筛选)
// GET /documents
func GetDocumentList(c *gin.Context) {
	var req dto.GetDocumentListDTO

	// 1. 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 处理布尔值指针 (默认为 false)
	isSuggest := false
	if req.IsSuggest != nil {
		isSuggest = *req.IsSuggest
	}

	// 3. 调用DAO获取文档列表
	documents, err := dao.GetDocumentList(isSuggest, req.CategoryID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 5. 构建响应列表
	var results []response.DocumentDetailResponse
	for _, doc := range documents {
		// 复用已有的详情构建函数，它已经包含了PostList的构建逻辑
		resp, err := response.BuildDocumentDetailResponse(doc)
		if err != nil {
			// 如果某条数据构建失败（如关联数据缺失），记录日志并跳过，确保列表整体可用
			continue
		}
		results = append(results, resp)
	}

	// 确保返回空数组而不是null
	if results == nil {
		results = make([]response.DocumentDetailResponse, 0)
	}

	// 6. 返回成功响应
	response.SuccessWithData(c, results, constant.DocumentsObtain)
}
