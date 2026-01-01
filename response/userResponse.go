package response

import (
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/utils"
)

// UserBriefResponse 定义了用户基本信息
type UserBriefResponse struct {
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	UserAvatar string `json:"userAvatar"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

// HomepageResponse 是用户主页接口返回的完整数据结构
type HomepageResponse struct {
	UserBrief      UserBriefResponse   `json:"userBrief"`
	Password       string              `json:"password"`
	CollectionList []InfoBriefResponse `json:"collectionList"`
	HistoryList    []InfoBriefResponse `json:"historyList"`
}

func BuildHomepageResponse(user models.User, collectionList, historyList []models.Document) (HomepageResponse, error) {
	// 1. 构建用户简要信息 (UserBrief)
	userBrief := UserBriefResponse{
		UserID:     user.ID,
		Username:   user.Username,
		UserAvatar: utils.GetFileURL(user.Avatar), // 使用工具函数处理头像URL
		Status:     user.Status,
		CreateTime: user.CreatedAt.Format("2006-01-02 15:04:05"),
		Email:      user.Email,
		Role:       user.Role,
	}

	// 2. 构建收藏列表 (CollectionList)
	var collectionResponses []InfoBriefResponse
	for _, doc := range collectionList {
		// 将单个文档信息转换为 InfoBriefResponse 格式
		category, _ := dao.GetCategoryByID(doc.CategoryID)
		docBrief := InfoBriefResponse{
			Name:        doc.Name,
			DocumentID:  doc.ID,
			Type:        doc.Type,
			UploadTime:  doc.CreatedAt.Format("2006-01-02 15:04:05"),
			Status:      doc.Status,
			Category:    category.Name,
			Collections: doc.Collections,
			ReadCounts:  doc.ReadCounts,
			Cover:       utils.GetFileURL(doc.Cover),
		}
		collectionResponses = append(collectionResponses, docBrief)
	}
	// 确保切片不为 nil，返回空数组 [] 而不是 null
	if collectionResponses == nil {
		collectionResponses = make([]InfoBriefResponse, 0)
	}

	// 3. 构建历史记录列表 (HistoryList)
	var historyResponses []InfoBriefResponse
	for _, doc := range historyList {
		// 将单个文档信息转换为 InfoBriefResponse 格式
		category, _ := dao.GetCategoryByID(doc.CategoryID)
		docBrief := InfoBriefResponse{
			Name:        doc.Name,
			DocumentID:  doc.ID,
			Type:        doc.Type,
			UploadTime:  doc.CreatedAt.Format("2006-01-02 15:04:05"),
			Status:      doc.Status,
			Category:    category.Name,
			Collections: doc.Collections,
			ReadCounts:  doc.ReadCounts,
			Cover:       utils.GetFileURL(doc.Cover),
		}
		historyResponses = append(historyResponses, docBrief)
	}
	// 确保切片不为 nil
	if historyResponses == nil {
		historyResponses = make([]InfoBriefResponse, 0)
	}

	// 4. 组装最终的 HomepageResponse
	response := HomepageResponse{
		UserBrief:      userBrief,
		Password:       user.Password, // 注意安全风险
		CollectionList: collectionResponses,
		HistoryList:    historyResponses,
	}

	return response, nil
}
