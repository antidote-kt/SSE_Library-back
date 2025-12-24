package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func formatTimeForResponse(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func buildCommentResponse(comment models.Comment) dto.CommentResponseDTO {
	commenter := dto.UserBriefDTO{
		UserID: comment.UserID,
	}
	if comment.User != nil {
		commenter.Username = comment.User.Username
		commenter.UserAvatar = utils.GetFileURL(comment.User.Avatar)
		commenter.Status = comment.User.Status
		commenter.CreateTime = formatTimeForResponse(comment.User.CreatedAt)
		commenter.Email = comment.User.Email
		commenter.Role = comment.User.Role
	}

	var sourceData *dto.SourceDataDTO = nil
	if comment.SourceType == "document" && comment.Document != nil {
		sourceData = &dto.SourceDataDTO{
			SourceID:   comment.SourceID,
			Name:       comment.Document.Name,
			SourceType: "document",
		}
	} else if comment.SourceType == "post" && comment.Post != nil {
		sourceData = &dto.SourceDataDTO{
			SourceID:   comment.SourceID,
			Name:       comment.Post.Title,
			SourceType: "post",
		}
	}

	return dto.CommentResponseDTO{
		CommentID:  comment.ID,
		ParentID:   comment.ParentID,
		SourceData: sourceData,
		Commenter:  commenter,
		CreatedAt:  formatTimeForResponse(comment.CreatedAt),
		Content:    comment.Content,
	}
}

func buildCommentResponseList(comments []models.Comment) []dto.CommentResponseDTO {
	commentList := make([]dto.CommentResponseDTO, 0, len(comments))
	for _, comment := range comments {
		commentList = append(commentList, buildCommentResponse(comment))
	}
	return commentList
}

func getCommentIDFromQuery(c *gin.Context) (string, error) {
	commentIDStr := c.Query("commentId")
	if commentIDStr == "" {
		commentIDStr = c.Query("comment_id")
	}
	if commentIDStr == "" {
		commentIDStr = c.Query("comment_Id")
	}
	if commentIDStr == "" {
		return "", errors.New(constant.MsgCommentIDEmpty)
	}
	return commentIDStr, nil
}

// POST /user/comments
func PostComment(c *gin.Context) {
	var request dto.PostCommentDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgParameterError)
		return
	}

	// 验证 sourceData
	if request.SourceData == nil {
		response.Fail(c, http.StatusBadRequest, nil, "sourceData 不能为空")
		return
	}

	// 验证评论内容
	if request.Content == "" {
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgContentEmpty)
		return
	}

	// 验证 commenter
	if request.Commenter.UserID == 0 {
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgUserIDEmpty)
		return
	}

	user, err := dao.GetUserByID(request.Commenter.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusUnauthorized, nil, constant.MsgUnauthorized)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	if request.Commenter.Username != user.Username {
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgUserInfoMismatch)
		return
	}

	// 验证 sourceData 对应的文档或帖子是否存在
	sourceID := request.SourceData.SourceID
	sourceType := request.SourceData.SourceType

	switch sourceType {
	case "document":
		_, err = dao.GetDocumentByID(sourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.MsgRecordNotFound)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
	case "post":
		_, err = dao.GetPostByID(sourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.MsgRecordNotFound)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
	default:
		response.Fail(c, http.StatusBadRequest, nil, "sourceType 必须是 document 或 post")
		return
	}

	// 如果有父评论ID，验证父评论是否存在
	if request.ParentID != nil && *request.ParentID != 0 {
		parentComment, err := dao.GetCommentByID(*request.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusBadRequest, nil, constant.MsgParentCommentNotFound)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}

		// 验证父评论是否在同一个 source 下
		if parentComment.SourceID != sourceID || parentComment.SourceType != sourceType {
			response.Fail(c, http.StatusBadRequest, nil, constant.MsgParentCommentNotInDocument)
			return
		}
	}

	// 创建评论
	comment := &models.Comment{
		UserID:     request.Commenter.UserID,
		Content:    request.Content,
		SourceID:   sourceID,
		SourceType: sourceType,
		ParentID:   request.ParentID,
	}

	err = dao.CreateComment(comment)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.MsgCommentCreateFailed)
		return
	}

	// 获取更新后的评论列表
	comments, err := dao.GetCommentWithUserAndDocument(sourceID, sourceType)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.MsgGetCommentListFailed)
		return
	}

	commentList := buildCommentResponseList(comments)

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": constant.MsgCommentPostSuccess,
		"data":    commentList,
	})
}

// getCommentsBySource 处理评论获取逻辑
func getCommentsBySource(c *gin.Context, sourceID uint64, sourceType string) {
	switch sourceType {
	case "document":
		_, err := dao.GetDocumentByID(sourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, constant.StatusNotFound, nil, constant.MsgRecordNotFound)
				return
			}
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
	case "post":
		_, err := dao.GetPostByID(sourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, constant.StatusNotFound, nil, constant.MsgRecordNotFound)
				return
			}
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
	default:
		response.Fail(c, constant.StatusBadRequest, nil, "sourceType 必须是 document 或 post")
		return
	}

	// 获取评论列表
	comments, err := dao.GetCommentWithUserAndDocument(sourceID, sourceType)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgGetCommentListFailed)
		return
	}

	commentList := buildCommentResponseList(comments)
	response.SuccessWithData(c, commentList, constant.MsgGetCommentListSuccess)
}

// GET /post/{sourceId}/comments
func GetPostComments(c *gin.Context) {
	sourceIDStr := c.Param("sourceId")

	sourceType := "post"

	if sourceIDStr == "" {
		response.Fail(c, constant.StatusBadRequest, nil, "sourceId 参数缺失")
		return
	}

	sourceID, err := strconv.ParseUint(sourceIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, "sourceId 格式错误")
		return
	}

	getCommentsBySource(c, sourceID, sourceType)
}

// GET /document/{sourceId}/comments
func GetDocumentComments(c *gin.Context) {
	sourceIDStr := c.Param("sourceId")

	sourceType := "document"

	if sourceIDStr == "" {
		response.Fail(c, constant.StatusBadRequest, nil, "sourceId 参数缺失")
		return
	}

	sourceID, err := strconv.ParseUint(sourceIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, "sourceId 格式错误")
		return
	}

	getCommentsBySource(c, sourceID, sourceType)
}

// GET /api/admin/comments
func GetAllComments(c *gin.Context) {
	comments, err := dao.GetAllCommentsWithUserAndDocument()
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgGetCommentListFailed)
		return
	}

	commentList := buildCommentResponseList(comments)

	response.SuccessWithData(c, commentList, constant.MsgGetAllCommentsSuccess)
}

// DELETE /api/admin/comment
func DeleteComment(c *gin.Context) {
	commentIDStr, err := getCommentIDFromQuery(c)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, err.Error())
		return
	}

	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgCommentIDFormatError)
		return
	}

	_, err = dao.GetCommentByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusNotFound, nil, constant.MsgCommentNotFound)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	err = dao.DeleteComment(commentID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCommentDeleteFailed)
		return
	}

	response.Response(c, constant.StatusOK, constant.CodeSuccess, constant.MsgCommentDeleteSuccess, gin.H{})
}

// GET /users/{user_id}/comments
func GetUserComments(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgUserIDFormatError)
		return
	}

	_, err = dao.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusNotFound, nil, constant.MsgUserNotFound)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	comments, err := dao.GetUserCommentsWithUserAndDocument(userID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgGetCommentListFailed)
		return
	}

	commentList := buildCommentResponseList(comments)

	response.SuccessWithData(c, commentList, constant.MsgGetUserCommentsSuccess)
}

// DELETE /user/comment
func DeleteUserComment(c *gin.Context) {
	// 查询参数：userId 和 commentId
	userIDStr := c.Query("userId")
	commentIDStr := c.Query("commentId")

	if commentIDStr == "" {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgCommentIDEmpty)
		return
	}

	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgCommentIDFormatError)
		return
	}

	// 获取评论信息
	comment, err := dao.GetCommentDetailByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusNotFound, nil, constant.MsgCommentNotFound)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	// 如果提供了 userId，验证评论是否属于该用户
	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			response.Fail(c, constant.StatusBadRequest, nil, constant.MsgUserIDFormatError)
			return
		}

		// 验证用户是否存在
		_, err = dao.GetUserByID(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, constant.StatusNotFound, nil, constant.MsgUserNotFound)
				return
			}
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}

		// 验证评论是否属于该用户
		if comment.UserID != userID {
			response.Fail(c, constant.StatusNotFound, nil, constant.MsgCommentNotFoundOrNoAccess)
			return
		}
	}

	// 执行删除
	err = dao.DeleteUserComment(comment.UserID, commentID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCommentDeleteFailed)
		return
	}

	// 成功后不返回评论列表，只返回成功消息
	response.Response(c, constant.StatusOK, constant.CodeSuccess, constant.MsgCommentDeleteSuccess, gin.H{})
}

// GetSingleComment 获取单条评论
// GET /comment/{commentId}
func GetSingleComment(c *gin.Context) {
	commentIDStr := c.Param("commentId")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgCommentIDFormatError)
		return
	}

	// 获取评论详情
	comment, err := dao.GetCommentDetailByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusNotFound, nil, constant.MsgCommentNotFound)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	// 使用 buildCommentResponse 构建响应
	commentResponse := buildCommentResponse(*comment)

	// 构建符合 API 文档的响应结构（评论数据包装在 data 字段中）
	commentData := gin.H{
		"commentId":  commentResponse.CommentID,
		"parentId":   commentResponse.ParentID,
		"sourceData": commentResponse.SourceData,
		"commenter":  commentResponse.Commenter,
		"createdAt":  commentResponse.CreatedAt,
		"content":    commentResponse.Content,
	}

	response.SuccessWithData(c, commentData, constant.MsgGetSingleCommentSuccess)
}
