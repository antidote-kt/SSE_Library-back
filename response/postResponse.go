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
	ReadCount    uint32              `json:"readCount"`
	LikeCount    uint32              `json:"likeCount"`
	SendTime     string              `json:"sendTime"`
	Documents    []InfoBriefResponse `json:"documents"`
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
			URL:         utils.GetResponseFileURL(doc), // 处理文档链接
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
		ReadCount:    post.ReadCount,
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
			ReadCount:    post.ReadCount,
			LikeCount:    post.LikeCount,
			SendTime:     post.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return responses
}
