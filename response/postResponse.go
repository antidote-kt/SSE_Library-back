package response

import (
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/utils"
)

// PostDetailResponse 帖子详情响应结构
type PostDetailResponse struct {
	PostID       uint64              `json:"postId"`
	SenderID     uint64              `json:"senderId"`
	SenderName   string              `json:"senderName"`
	SenderAvatar string              `json:"senderAvatar"`
	Title        string              `json:"title"`
	Content      string              `json:"content"`
	CommentCount uint32              `json:"commentCount"`
	LikeCount    uint32              `json:"likeCount"`
	SendTime     string              `json:"sendTime"`
	Documents    []InfoBriefResponse `json:"documents"`
	CollectCount uint32              `json:"collectCount"`
}

// postBrief
type PostBriefResponse struct {
	CommentCount uint32 `json:"commentCount"`
	Content      string `json:"content"`
	LikeCount    uint32 `json:"likeCount"`
	PostID       uint64 `json:"postId"`
	CollectCount uint32 `json:"collectCount"`
	SenderAvatar string `json:"senderAvatar"`
	SenderID     uint64 `json:"senderId"`
	SenderName   string `json:"senderName"`
	SendTime     string `json:"sendTime"`
	Title        string `json:"title"`
}

// BuildPostBriefResponse 构建帖子简要响应对象
func BuildPostBriefResponse(post models.Post) PostBriefResponse {
	// 获取发帖用户的信息
	user, err := dao.GetUserByID(post.SenderID)
	senderName := ""
	senderAvatar := ""
	if err != nil {
		// 如果获取用户信息失败，使用默认值
		senderName = "未知用户"
		senderAvatar = ""
	} else {
		senderName = user.Username
		senderAvatar = utils.GetFileURL(user.Avatar)
	}

	return PostBriefResponse{
		PostID:       post.ID,
		SenderID:     post.SenderID,
		SenderName:   senderName,
		SenderAvatar: senderAvatar,
		Title:        post.Title,
		Content:      post.Content,
		CommentCount: post.CommentCount,
		CollectCount: post.CollectCount,
		LikeCount:    post.LikeCount,
		SendTime:     post.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// BuildPostBriefResponseList 构建帖子简要响应对象列表
func BuildPostBriefResponseList(posts []models.Post) []PostBriefResponse {
	postBriefResponses := make([]PostBriefResponse, len(posts))
	for i, post := range posts {
		postBriefResponses[i] = BuildPostBriefResponse(post)
	}
	return postBriefResponses
}

// BuildPostDetailResponse 构建单个帖子详情响应
func BuildPostDetailResponse(post models.Post, documents []models.Document) PostDetailResponse {
	// 根据post.SenderID到user表里查询用户信息（SenderName和SenderAvatar）
	Sender, _ := dao.GetUserByID(post.SenderID)
	SenderName := Sender.Username
	SenderAvatar := Sender.Avatar

	// 构建文档摘要
	var InfoBriefs []InfoBriefResponse
	for _, doc := range documents {
		InfoBriefs = append(InfoBriefs, InfoBriefResponse{
			Name:        doc.Name,
			DocumentID:  doc.ID,
			Type:        doc.Type,
			UploadTime:  doc.CreatedAt.Format("2006-01-02 15:04:05"),
			Status:      doc.Status,
			Category:    strconv.FormatUint(doc.CategoryID, 10),
			Collections: doc.Collections,
			ReadCounts:  doc.ReadCounts,
			Cover:       utils.GetFileURL(doc.Cover), // 处理文档链接
		})
	}

	return PostDetailResponse{
		PostID:       post.ID,
		SenderID:     post.SenderID,
		SenderName:   SenderName,
		SenderAvatar: utils.GetFileURL(SenderAvatar), // 处理头像URL
		Title:        post.Title,
		Content:      post.Content,
		CommentCount: post.CommentCount,
		CollectCount: post.CollectCount,
		LikeCount:    post.LikeCount,
		SendTime:     post.CreatedAt.Format("2006-01-02 15:04:05"),
		Documents:    InfoBriefs,
	}
}

// PostListResponse 帖子列表响应结构
type PostListResponse struct {
	PostID       uint64 `json:"postId"`
	SenderID     uint64 `json:"senderId"`
	SenderName   string `json:"senderName"`
	SenderAvatar string `json:"senderAvatar"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	CommentCount uint32 `json:"commentCount"`
	CollectCount uint32 `json:"collectCount"`
	ReadCount    uint32 `json:"readCount"`
	LikeCount    uint32 `json:"likeCount"`
	SendTime     string `json:"sendTime"`
}

// BuildPostListResponse 构建帖子列表响应
func BuildPostListResponse(posts []models.Post) []PostListResponse {
	var responses []PostListResponse
	for _, post := range posts {
		// 获取发帖人信息
		sender, _ := dao.GetUserByID(post.SenderID)
		senderName := sender.Username
		senderAvatar := sender.Avatar

		responses = append(responses, PostListResponse{
			PostID:       post.ID,
			SenderID:     post.SenderID,
			SenderName:   senderName,
			SenderAvatar: utils.GetFileURL(senderAvatar),
			Title:        post.Title,
			Content:      post.Content,
			CommentCount: post.CommentCount,
			CollectCount: post.CollectCount,
			LikeCount:    post.LikeCount,
			SendTime:     post.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return responses
}
