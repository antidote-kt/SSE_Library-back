package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CategoryResponse
type CategoryResponse struct {
	ID          uint64              `json:"id"`
	Name        string              `json:"name"`
	IsCourse    bool                `json:"isCourse"`
	FileCounts  int64               `json:"fileCounts"`
	ReadCounts  int64               `json:"readCounts"`
	Description string              `json:"description,omitempty"`
	ParentID    *uint64             `json:"parentId,omitempty"`
	Children    []*CategoryResponse `json:"children,omitempty"`
}

// buildCategoryTree 构建分类树
func buildCategoryTree(categories []models.Category, fileCounts map[uint64]int64, readCounts map[uint64]int64) []*CategoryResponse {
	// 分类ID对应分类映射
	categoryMap := make(map[uint64]*CategoryResponse)
	var rootCategories []*CategoryResponse

	for _, category := range categories {
		categoryResp := &CategoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			IsCourse:    category.IsCourse,
			FileCounts:  fileCounts[category.ID],
			ReadCounts:  readCounts[category.ID],
			Description: category.Description,
			ParentID:    category.ParentID,
			Children:    make([]*CategoryResponse, 0),
		}
		categoryMap[category.ID] = categoryResp
	}

	for _, category := range categories {
		categoryResp := categoryMap[category.ID]
		if category.ParentID == nil {
			// 顶级分类
			rootCategories = append(rootCategories, categoryResp)
		} else {
			// 将子分类添加到父分类的children中
			if parent, ok := categoryMap[*category.ParentID]; ok {
				parent.Children = append(parent.Children, categoryResp)
			}
		}
	}

	return rootCategories
}

// GetCategoriesAndCourses 获取所有分类和课程
// GET /api/category
func GetCategoriesAndCourses(c *gin.Context) {
	// 获取所有分类和课程
	categories, err := dao.GetAllCategories()
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	// 统计每类文档的数量和浏览量
	fileCounts := make(map[uint64]int64)
	readCounts := make(map[uint64]int64)

	for _, category := range categories {
		// 统计文档数量
		fileCount, err := dao.CountDocumentsByCategory(category.ID)
		if err != nil {
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
		fileCounts[category.ID] = fileCount

		// 统计浏览量
		readCount, err := dao.GetDocumentReadCountsByCategory(category.ID)
		if err != nil {
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
		readCounts[category.ID] = readCount
	}

	// 构建分类树
	categoryTree := buildCategoryTree(categories, fileCounts, readCounts)

	response.SuccessWithData(c, categoryTree, constant.MsgGetCategoriesSuccess)
}

// SearchCategoriesAndCourses 搜索分类和课程
// GET /api/searchcat?name=xxx
func SearchCategoriesAndCourses(c *gin.Context) {
	// 获取分类和课程
	name := c.Query("name")
	if name == "" {
		// 如果没有搜索关键词，返回所有分类和课程
		GetCategoriesAndCourses(c)
		return
	}

	// 搜索分类和课程
	categories, err := dao.SearchCategoriesByName(name)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	// 统计每类文档的数量和浏览量
	fileCounts := make(map[uint64]int64)
	readCounts := make(map[uint64]int64)

	for _, category := range categories {
		// 统计文档数量
		fileCount, err := dao.CountDocumentsByCategory(category.ID)
		if err != nil {
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCategoryCountFailed)
			return
		}
		fileCounts[category.ID] = fileCount

		// 统计浏览量
		readCount, err := dao.GetDocumentReadCountsByCategory(category.ID)
		if err != nil {
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCategoryReadCountFailed)
			return
		}
		readCounts[category.ID] = readCount
	}

	// 构建分类树
	categoryTree := buildCategoryTree(categories, fileCounts, readCounts)

	response.SuccessWithData(c, categoryTree, constant.MsgGetCategoriesSuccess)
}

// AddCategory 添加分类/课程接口
func AddCategory(c *gin.Context) {
	var req dto.AddCategoryDTO

	// 1. 绑定参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 业务校验
	// 2.1 如果指定了父分类，检查父分类是否存在
	if req.ParentCatID != nil && *req.ParentCatID != 0 {
		_, err := dao.GetCategoryByID(*req.ParentCatID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusBadRequest, nil, constant.ParentCategoryNotExist)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
		//可选：检查父分类是否允许有子分类（例如，如果父分类本身是课程，不允许再有子分类）
		//if parentCategory.IsCourse {
		// response.Fail(c, http.StatusBadRequest, nil, "课程不能作为父分类")
		// return
		//}
	}

	// 2.2 检查同名分类 (防止重复)
	existingCats, _ := dao.GetCategoryByName(req.Name)
	if len(existingCats) > 0 {
		// 简单的重名检查，更严格的检查应该看是否在同一个父分类下重名
		response.Fail(c, http.StatusUnprocessableEntity, nil, constant.CategoryNameAlreadyExist)
		return
	}

	// 3. 构建模型
	category := models.Category{
		Name:        req.Name,
		IsCourse:    *req.IsCourse,
		Description: req.Description,
		ParentID:    req.ParentCatID,
		// 其他三个时间相关字段在dao.CreateCategory(&category)这一步会由gorm自动处理
	}

	// 4. 保存到数据库
	if err := dao.CreateCategory(&category); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.MsgCategoryCreateFailed)
		return
	}

	// 5. 返回成功响应
	response.Success(c, gin.H{"success": true}, constant.MsgCategoryCreateSuccess)
}

// DeleteCategory 删除分类或课程
// DELETE /api/category?name=xxx
func DeleteCategory(c *gin.Context) {
	// 获取查询参数 name
	name := c.Query("name")
	if name == "" {
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgCategoryNameRequired)
		return
	}

	// 检查分类是否存在
	categories, err := dao.GetCategoryByName(name)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	if len(categories) == 0 {
		response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
		return
	}

	// 删除分类（软删除）
	err = dao.DeleteCategoryByName(name)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.MsgCategoryDeleteFailed)
		return
	}

	// 返回成功响应（根据图片要求，返回格式为 {"code": 0, "message": "string"}）
	response.Success(c, nil, constant.MsgCategoryDeleteSuccess)
}

// ModifyCategoryResponse 修改分类响应结构体
type ModifyCategoryResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsCourse    string `json:"isCourse"`
	ParentID    uint64 `json:"parentId"`
}

// ModifyCategory 修改分类或课程
// PUT /api/category
func ModifyCategory(c *gin.Context) {
	var request dto.ModifyCategoryDTO

	// 解析请求参数
	if err := c.ShouldBind(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 至少需要提供 id 或 name 来定位分类
	if request.ID == nil && (request.Name == nil || *request.Name == "") {
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgCategoryIDOrNameRequired)
		return
	}

	var category models.Category
	var err error

	// 根据 id 或 name 查找分类
	if request.ID != nil {
		category, err = dao.GetCategoryByID(*request.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
	} else if request.Name != nil && *request.Name != "" {
		categories, err := dao.GetCategoryByName(*request.Name)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
		if len(categories) == 0 {
			response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
			return
		}
		// 如果找到多个同名分类，取第一个
		category = categories[0]
	}

	// 动态更新分类字段（仅更新客户端提供的字段）
	if request.Name != nil && *request.Name != "" {
		category.Name = *request.Name
	}
	if request.Description != nil {
		category.Description = *request.Description
	}
	if request.IsCourse != nil {
		// 将 string 转换为 bool
		isCourse, err := strconv.ParseBool(*request.IsCourse)
		if err != nil {
			// 如果转换失败，尝试其他可能的格式
			if *request.IsCourse == "1" || *request.IsCourse == "true" || *request.IsCourse == "True" {
				isCourse = true
			} else {
				isCourse = false
			}
		}
		category.IsCourse = isCourse
	}
	if request.ParentID != nil {
		// 如果 parentId 为 0，表示设置为顶级分类
		if *request.ParentID == 0 {
			category.ParentID = nil
		} else {
			// 验证父分类是否存在
			parentCategory, err := dao.GetCategoryByID(*request.ParentID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					response.Fail(c, http.StatusNotFound, nil, "父分类不存在")
					return
				}
				response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
				return
			}
			category.ParentID = &parentCategory.ID
		}
	}

	// 更新分类到数据库
	err = dao.UpdateCategory(&category)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.MsgCategoryUpdateFailed)
		return
	}

	// 构建响应数据
	parentID := uint64(0)
	if category.ParentID != nil {
		parentID = *category.ParentID
	}
	isCourseStr := "false"
	if category.IsCourse {
		isCourseStr = "true"
	}

	responseData := ModifyCategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsCourse:    isCourseStr,
		ParentID:    parentID,
	}

	// 返回成功响应（根据图片要求，返回格式包含 data 对象）
	response.SuccessWithData(c, responseData, constant.MsgCategoryUpdateSuccess)
}

// buildCategoryResponseWithChildren 递归构建分类响应，包括子分类
func buildCategoryResponseWithChildren(category models.Category, fileCounts map[uint64]int64, readCounts map[uint64]int64) *CategoryResponse {
	categoryResp := &CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		IsCourse:    category.IsCourse,
		FileCounts:  fileCounts[category.ID],
		ReadCounts:  readCounts[category.ID],
		Description: category.Description,
		ParentID:    category.ParentID,
		Children:    make([]*CategoryResponse, 0),
	}

	// 递归获取子分类
	children, err := dao.GetCategoriesByParentID(category.ID)
	if err == nil && len(children) > 0 {
		// 为子分类统计文件数量和浏览量
		childrenFileCounts := make(map[uint64]int64)
		childrenReadCounts := make(map[uint64]int64)

		for _, child := range children {
			fileCount, _ := dao.CountDocumentsByCategory(child.ID)
			readCount, _ := dao.GetDocumentReadCountsByCategory(child.ID)
			childrenFileCounts[child.ID] = fileCount
			childrenReadCounts[child.ID] = readCount
		}

		// 递归构建子分类响应
		for _, child := range children {
			childResp := buildCategoryResponseWithChildren(child, childrenFileCounts, childrenReadCounts)
			categoryResp.Children = append(categoryResp.Children, childResp)
		}
	}

	return categoryResp
}

// GetCategoryDetail 获取特定的分类或课程详情
// GET /category/:categoryId
func GetCategoryDetail(c *gin.Context) {
	// 获取路径参数
	categoryIDStr := c.Param("categoryId")
	if categoryIDStr == "" {
		response.Fail(c, constant.StatusBadRequest, nil, "categoryId 参数不能为空")
		return
	}

	// 解析 categoryId
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, "categoryId 格式错误")
		return
	}

	// 获取分类信息
	category, err := dao.GetCategoryByID(categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusNotFound, nil, constant.CategoryNotExist)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	// 统计文件数量和浏览量
	fileCount, err := dao.CountDocumentsByCategory(category.ID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCategoryCountFailed)
		return
	}

	readCount, err := dao.GetDocumentReadCountsByCategory(category.ID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCategoryReadCountFailed)
		return
	}

	// 构建文件数量和浏览量映射
	fileCounts := map[uint64]int64{category.ID: fileCount}
	readCounts := map[uint64]int64{category.ID: readCount}

	// 递归构建分类响应（包括子分类）
	categoryResp := buildCategoryResponseWithChildren(category, fileCounts, readCounts)

	// 返回成功响应
	response.SuccessWithData(c, categoryResp, constant.MsgGetCategoryDetailSuccess)
}
