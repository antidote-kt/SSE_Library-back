package controllers

import (
	"errors"
	"strconv"
	"time"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func buildCommentResponse(commentData dao.CommentWithDetails) dto.CommentResponseDTO {
	return dto.CommentResponseDTO{
		CommentID: uint64(commentData.CommentID),
		ParentID:  commentData.ParentID,
		Commenter: dto.UserBriefDTO{
			UserID:     uint64(commentData.UserID),
			Username:   commentData.Username,
			UserAvatar: commentData.UserAvatar,
			Status:     commentData.Status,
			CreateTime: commentData.UserCreateTime.Format("2006-01-02 15:04:05"),
			Email:      commentData.Email,
			Role:       commentData.Role,
		},
		Document: dto.DocumentBriefDTO{
			Name:        commentData.DocumentName,
			DocumentID:  uint64(commentData.DocumentID),
			Type:        commentData.Type,
			UploadTime:  commentData.DocumentUploadTime.Format("2006-01-02 15:04:05"),
			Status:      commentData.DocumentStatus,
			Category:    commentData.Category,
			Course:      commentData.Course,
			Collections: int(commentData.Collections),
			ReadCounts:  int(commentData.ReadCounts),
			URL:         commentData.URL,
			Content:     commentData.DocumentContent,
			CreateTime:  commentData.DocumentCreateTime.Format("2006-01-02 15:04:05"),
		},
		CreatedAt: commentData.CreatedAt.Format("2006-01-02 15:04:05"),
		Content:   commentData.CommentContent,
	}
}

func buildCommentResponseList(comments []dao.CommentWithDetails) []dto.CommentResponseDTO {
	var commentList []dto.CommentResponseDTO
	for _, commentData := range comments {
		commentList = append(commentList, buildCommentResponse(commentData))
	}
	return commentList
}

func getCommentIDFromQuery(c *gin.Context) (string, error) {
	commentIDStr := c.Query("comment_id")
	if commentIDStr == "" {
		commentIDStr = c.Query("comment_Id")
	}
	if commentIDStr == "" {
		return "", errors.New(constant.MsgCommentIDEmpty)
	}
	return commentIDStr, nil
}

// POST /api/books/{document_id}/comments
func PostComment(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgDocumentIDFormatError)
		return
	}

	// 验证文档是否存在
	_, err = dao.GetDocumentByID(documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusNotFound, nil, constant.MsgRecordNotFound)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}
	var request dto.PostCommentDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, constant.StatusBadRequest, gin.H{"error": err.Error()}, constant.MsgParameterError)
		return
	}

	// 验证评论内容
	if request.Content == "" {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgContentEmpty)
		return
	}

	if request.Author.UserID == 0 {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgUserIDEmpty)
		return
	}

	user, err := dao.GetUserByID(uint64(request.Author.UserID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusUnauthorized, nil, constant.MsgUnauthorized)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	if request.Author.UserName != user.Username {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgUserInfoMismatch)
		return
	}

	var parsedTime time.Time
	parsedTime, err = time.Parse(time.RFC3339, request.CreateTime)
	if err != nil {
		parsedTime, err = time.ParseInLocation("2006-01-02 15:04:05", request.CreateTime, time.Local)
		if err != nil {
			parsedTime, err = time.ParseInLocation("2006-01-02 15:04", request.CreateTime, time.Local)
			if err != nil {
				response.Fail(c, constant.StatusBadRequest, nil, constant.MsgCreateTimeFormatError)
				return
			}
		}
	}

	// 如果有父评论ID，验证父评论是否存在
	if request.ParentID != nil && *request.ParentID != 0 {
		parentComment, err := dao.GetCommentByID(*request.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, constant.StatusNotFound, nil, constant.MsgParentCommentNotFound)
				return
			}
			response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}

		if parentComment.DocumentID != documentID {
			response.Fail(c, constant.StatusBadRequest, nil, constant.MsgParentCommentNotInDocument)
			return
		}
	}

	// 创建评论
	comment := &models.Comment{
		UserID:     request.Author.UserID,
		Content:    request.Content,
		DocumentID: documentID,
		ParentID:   request.ParentID,
		CreatedAt:  parsedTime,
	}

	err = dao.CreateComment(comment)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCommentCreateFailed)
		return
	}

	// 获取更新后的评论列表
	comments, err := dao.GetCommentWithUserAndDocument(documentID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgGetCommentListFailed)
		return
	}

	responseData := gin.H{
		"code":    constant.CodeSuccess,
		"message": constant.MsgCommentPostSuccess,
		"data":    buildCommentResponseList(comments),
	}

	c.JSON(constant.StatusCreated, responseData)
}

// GET /api/books/{document_id}/comments
func GetComments(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgDocumentIDFormatError)
		return
	}

	_, err = dao.GetDocumentByID(documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusNotFound, nil, constant.MsgRecordNotFound)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	comments, err := dao.GetCommentWithUserAndDocument(documentID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgGetCommentListFailed)
		return
	}

	responseData := gin.H{
		"code":    constant.CodeSuccess,
		"message": constant.MsgGetCommentListSuccess,
		"data":    buildCommentResponseList(comments),
	}

	c.JSON(constant.StatusOK, responseData)
}

// GET /api/admin/comments
func GetAllComments(c *gin.Context) {
	comments, err := dao.GetAllCommentsWithUserAndDocument()
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgGetCommentListFailed)
		return
	}

	responseData := gin.H{
		"code":    constant.CodeSuccess,
		"message": constant.MsgGetAllCommentsSuccess,
		"data":    buildCommentResponseList(comments),
	}

	c.JSON(constant.StatusOK, responseData)
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

	responseData := gin.H{
		"code":    constant.CodeSuccess,
		"message": constant.MsgCommentDeleteSuccess,
	}

	c.JSON(constant.StatusOK, responseData)
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

	responseData := gin.H{
		"code":    constant.CodeSuccess,
		"message": constant.MsgGetUserCommentsSuccess,
		"data":    buildCommentResponseList(comments),
	}

	c.JSON(constant.StatusOK, responseData)
}

// DELETE /user/deleteComment
func DeleteUserComment(c *gin.Context) {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgUserIDEmpty)
		return
	}

	commentIDStr, err := getCommentIDFromQuery(c)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, err.Error())
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgUserIDFormatError)
		return
	}

	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		response.Fail(c, constant.StatusBadRequest, nil, constant.MsgCommentIDFormatError)
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

	deletedCommentInfo, err := dao.GetCommentDetailByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusNotFound, nil, constant.MsgCommentNotFoundOrNoAccess)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	if uint64(deletedCommentInfo.UserID) != userID {
		response.Fail(c, constant.StatusNotFound, nil, constant.MsgCommentNotFoundOrNoAccess)
		return
	}

	err = dao.DeleteUserComment(userID, commentID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCommentDeleteFailed)
		return
	}

	responseData := gin.H{
		"code":    constant.CodeSuccess,
		"message": constant.MsgCommentDeleteSuccess,
		"data":    []dto.CommentResponseDTO{buildCommentResponse(*deletedCommentInfo)},
	}

	c.JSON(constant.StatusOK, responseData)
}
