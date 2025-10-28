package controllers

import (
	"net/http"
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

// GetProfile 获取当前登录用户的个人主页信息
func GetProfile(c *gin.Context) {
	// 1. 从JWT中间件获取用户信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, "无法获取用户信息，请重新登录")
		return
	}
	userClaims := claims.(*utils.MyClaims)
	//将路径参数的用户id提取出来并转化为int64类型，与JWT比较看访问的个人主页接口是否与用户本人匹配
	paramUserID := c.Param("user_id")
	targetID, err := strconv.ParseUint(paramUserID, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "无效的用户ID格式")
		return
	}
	userID := userClaims.UserID
	if userID != targetID {
		response.Fail(c, http.StatusUnauthorized, nil, "访问的主页不是您本人的主页")
		return
	}

	// 2. 并行获取用户基本信息、收藏列表、历史记录
	userChan := make(chan models.User)
	collectionChan := make(chan []models.Document)
	historyChan := make(chan []models.Document)
	errChan := make(chan error, 3)

	go func() {
		user, err := dao.GetUserByID(userID)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- user
	}()

	go func() {
		collections, err := dao.GetFavoriteDocumentsByUserID(userID)
		if err != nil {
			errChan <- err
			return
		}
		collectionChan <- collections
	}()

	go func() {
		histories, err := dao.GetViewHistoryDocumentsByUserID(userID)
		if err != nil {
			errChan <- err
			return
		}
		historyChan <- histories
	}()

	// 等待所有goroutine完成
	var user models.User
	var collectionList, historyList []models.Document
	for i := 0; i < 3; i++ {
		select {
		case u := <-userChan:
			user = u
		case cl := <-collectionChan:
			collectionList = cl
		case hl := <-historyChan:
			historyList = hl
		case err := <-errChan:
			response.Fail(c, http.StatusInternalServerError, nil, "获取数据失败: "+err.Error())
			return
		}
	}

	// 3. 组装成 DTO
	homepageDTO := dto.HomepageDTO{
		UserBrief:      buildUserBriefDTO(user),
		Password:       user.Password,
		CollectionList: buildDocumentDetailDTOs(collectionList),
		HistoryList:    buildDocumentDetailDTOs(historyList),
	}

	response.Success(c, gin.H{"profile": homepageDTO}, "获取个人主页信息成功")
}

// ModifyInfo 修改当前登录用户的个人资料
func ModifyInfo(c *gin.Context) {
	var req dto.ModifyInfoDTO

	// 1. 绑定并验证参数
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数格式错误: "+err.Error())
		return
	}

	// 2. 从JWT中间件获取用户信息
	claims, exists := c.Get("user_claims")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	// 3. 从数据库获取当前用户
	user, err := dao.GetUserByID(userClaims.UserID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, nil, "用户不存在")
		return
	}

	// 4. 检查是否有更新，并应用更新
	// 使用指针的好处：如果前端没传某个字段，这里的指针就是nil，我们就不更新它
	updated := false
	if req.UserName != nil && *req.UserName != user.Username {
		// 检查新用户名是否已被其他用户占用
		existingUser, _ := dao.GetUserByUsername(*req.UserName)
		if existingUser.ID != 0 && existingUser.ID != user.ID {
			response.Fail(c, http.StatusUnprocessableEntity, nil, "用户名已存在")
			return
		}
		user.Username = *req.UserName
		updated = true
	}
	if req.Email != nil && *req.Email != user.Email {
		// 检查新邮箱是否已被其他用户占用
		existingUser, _ := dao.GetUserByEmail(*req.Email)
		if existingUser.ID != 0 && existingUser.ID != user.ID {
			response.Fail(c, http.StatusUnprocessableEntity, nil, "邮箱已被注册")
			return
		}
		user.Email = *req.Email
		updated = true
	}
	if req.UserAvatar != nil && req.UserAvatar.Size != 0 {
		// 检查用户是否上传了新头像，有的话删除原头像
		err := utils.DeleteFile(user.Avatar)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, "头像删除失败")
			return
		}
		// 然后上传新头像到腾讯云
		avatarURL, err := utils.UploadAvatar(req.UserAvatar)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, "头像上传失败")
			return
		}
		//最后更新头像链接到数据模型
		user.Avatar = avatarURL
		updated = true
	}

	// 如果没有任何更新，可以直接返回成功
	if !updated {
		response.Success(c, nil, "没有需要更新的信息")
		return
	}

	// 5. 保存到数据库
	if err := dao.UpdateUser(user); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "更新用户信息失败")
		return
	}

	// 6. 按照要求，返回更新后的UserBrief信息
	response.Success(c, gin.H{
		"userBrief": buildUserBriefDTO(user),
	}, "个人资料修改成功")
}

// --- 以下是辅助函数，用于将Model转换为DTO ---

// buildUserBriefDTO 将 models.User 转换为 dto.UserBriefDTO
func buildUserBriefDTO(user models.User) dto.UserBriefDTO {
	return dto.UserBriefDTO{
		UserID:     user.ID,
		Username:   user.Username,
		UserAvatar: user.Avatar,
		Status:     user.Status,
		CreateTime: user.CreatedAt.Format("2006-01-02 15:04:05"),
		Email:      user.Email,
		Role:       user.Role,
	}
}

// buildDocumentDetailDTOs 将 []models.Document 转换为 []dto.DocumentDetailDTO
func buildDocumentDetailDTOs(documents []models.Document) []dto.DocumentDetailDTO {
	var dtoList []dto.DocumentDetailDTO
	if len(documents) == 0 {
		return dtoList
	}

	// --- 性能优化：批量收集所有需要的ID ---
	uploaderIDs := make(map[uint64]bool)
	categoryIDs := make(map[uint64]bool)
	docIDs := make([]uint64, len(documents))

	for i, doc := range documents {
		docIDs[i] = doc.ID
		uploaderIDs[doc.UploaderID] = true
		categoryIDs[doc.CategoryID] = true
	}

	// --- 批量查询关联数据 ---
	// 1. 批量查询所有相关的上传者
	users, _ := dao.GetUsersByIDs(mapToSlice(uploaderIDs)) // 假设DAO中有此函数
	userMap := make(map[uint64]models.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// 2. 批量查询所有相关的分类，以及它们的父分类
	categories, _ := dao.GetCategoriesByIDs(mapToSlice(categoryIDs))
	categoryMap := make(map[uint64]models.Category)
	parentIDsToFetch := make(map[uint64]bool)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
		if !cat.IsCourse && cat.ParentID != nil && *cat.ParentID != 0 {
			parentIDsToFetch[*cat.ParentID] = true
		}
	}
	parentCategories, _ := dao.GetCategoriesByIDs(mapToSlice(parentIDsToFetch))
	for _, pCat := range parentCategories {
		categoryMap[pCat.ID] = pCat // 将父分类也加入map，方便查找
	}

	//// 3. 批量查询所有相关的标签
	//tags, _ := dao.GetTagsByDocumentIDs(docIDs) // 假设DAO中有此函数，返回 map[uint64][]models.Tag
	//tagMap := tags

	// --- 在内存中高效组装DTO ---
	for _, doc := range documents {
		uploader := userMap[doc.UploaderID]
		category := categoryMap[doc.CategoryID]

		//var tagNames []string
		//if docTags, ok := tagMap[doc.ID]; ok {
		//	for _, tag := range docTags {
		//		tagNames = append(tagNames, tag.TagName)
		//	}
		//}

		courseName := ""
		if category.IsCourse {
			courseName = category.Name
		} else if category.ParentID != nil {
			if parentCat, ok := categoryMap[*category.ParentID]; ok && parentCat.IsCourse {
				courseName = parentCat.Name
			}
		}

		dtoList = append(dtoList, dto.DocumentDetailDTO{
			InfoBrief: dto.InfoBriefDTO{
				Name:        doc.Name,
				DocumentID:  doc.ID,
				Type:        doc.Type,
				UploadTime:  doc.CreatedAt.Format("2006-01-02 15:04:05"),
				Status:      doc.Status,
				Category:    category.Name,
				Course:      courseName,
				Collections: doc.Collections,
				ReadCounts:  doc.ReadCounts,
				URL:         utils.GetFileURL(doc.URL),
			},
			BookISBN: doc.BookISBN,
			Author:   doc.Author,
			Uploader: buildUserBriefDTO(uploader),
			Cover:    utils.GetFileURL(doc.Cover),
			//Tags:         tagNames,
			Introduction: doc.Introduction,
			CreateYear:   doc.CreateYear,
		})
	}
	return dtoList
}

// mapToSlice 是一个辅助函数，将map的key转换为slice
func mapToSlice(m map[uint64]bool) []uint64 {
	s := make([]uint64, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}
