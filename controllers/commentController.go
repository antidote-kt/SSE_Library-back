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

func formatTimeForResponse(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func loadCategoryNames(comments []models.Comment) (map[uint64]string, error) {
	categoryIDSet := make(map[uint64]struct{})
	for _, comment := range comments {
		if comment.Document != nil && comment.Document.CategoryID != 0 {
			categoryIDSet[comment.Document.CategoryID] = struct{}{}
		}
	}

	if len(categoryIDSet) == 0 {
		return map[uint64]string{}, nil
	}

	ids := make([]uint64, 0, len(categoryIDSet))
	for id := range categoryIDSet {
		ids = append(ids, id)
	}

	categories, err := dao.GetCategoriesByIDs(ids)
	if err != nil {
		return nil, err
	}

	categoryNames := make(map[uint64]string, len(categories))
	for _, category := range categories {
		categoryNames[category.ID] = category.Name
	}

	return categoryNames, nil
}

func buildCommentResponse(comment models.Comment, categoryNames map[uint64]string) dto.CommentResponseDTO {
	commenter := dto.UserBriefDTO{
		UserID: comment.UserID,
	}
	if comment.User != nil {
		commenter.Username = comment.User.Username
		commenter.UserAvatar = comment.User.Avatar
		commenter.Status = comment.User.Status
		commenter.CreateTime = formatTimeForResponse(comment.User.CreatedAt)
		commenter.Email = comment.User.Email
		commenter.Role = comment.User.Role
	}

	document := dto.DocumentBriefDTO{
		DocumentID: comment.DocumentID,
	}
	if comment.Document != nil {
		document.Name = comment.Document.Name
		document.Type = comment.Document.Type
		document.UploadTime = formatTimeForResponse(comment.Document.CreatedAt)
		document.Status = comment.Document.Status
		document.Collections = comment.Document.Collections
		document.ReadCounts = comment.Document.ReadCounts
		document.URL = comment.Document.URL
		document.Content = comment.Document.Introduction
		document.CreateTime = formatTimeForResponse(comment.Document.CreatedAt)
		if comment.Document.CategoryID != 0 {
			if categoryName, ok := categoryNames[comment.Document.CategoryID]; ok {
				document.Category = categoryName
				document.Course = categoryName
			}
		}
	}

	return dto.CommentResponseDTO{
		CommentID: comment.ID,
		ParentID:  comment.ParentID,
		Commenter: commenter,
		Document:  document,
		CreatedAt: formatTimeForResponse(comment.CreatedAt),
		Content:   comment.Content,
	}
}

func buildCommentResponseList(comments []models.Comment, categoryNames map[uint64]string) []dto.CommentResponseDTO {
	var commentList []dto.CommentResponseDTO
	for _, commentData := range comments {
		commentList = append(commentList, buildCommentResponse(commentData, categoryNames))
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
		UserID:     uint64(request.Author.UserID),
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

	categoryNames, err := loadCategoryNames(comments)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	response.Response(c, constant.StatusCreated, constant.CodeSuccess, constant.MsgCommentPostSuccess, gin.H{
		"list": buildCommentResponseList(comments, categoryNames),
	})
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

	categoryNames, err := loadCategoryNames(comments)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	response.Response(c, constant.StatusOK, constant.CodeSuccess, constant.MsgGetCommentListSuccess, gin.H{
		"list": buildCommentResponseList(comments, categoryNames),
	})
}

// GET /api/admin/comments
func GetAllComments(c *gin.Context) {
	comments, err := dao.GetAllCommentsWithUserAndDocument()
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgGetCommentListFailed)
		return
	}

	categoryNames, err := loadCategoryNames(comments)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	response.Response(c, constant.StatusOK, constant.CodeSuccess, constant.MsgGetAllCommentsSuccess, gin.H{
		"list": buildCommentResponseList(comments, categoryNames),
	})
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

	categoryNames, err := loadCategoryNames(comments)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	response.Response(c, constant.StatusOK, constant.CodeSuccess, constant.MsgGetUserCommentsSuccess, gin.H{
		"list": buildCommentResponseList(comments, categoryNames),
	})
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

	if deletedCommentInfo.UserID != userID {
		response.Fail(c, constant.StatusNotFound, nil, constant.MsgCommentNotFoundOrNoAccess)
		return
	}

	err = dao.DeleteUserComment(userID, commentID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgCommentDeleteFailed)
		return
	}

	// 获取删除后用户在该文档下的所有剩余评论
	remainingComments, err := dao.GetUserCommentsWithUserAndDocument(userID)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgGetCommentListFailed)
		return
	}

	categoryNames, err := loadCategoryNames(remainingComments)
	if err != nil {
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	response.Response(c, constant.StatusOK, constant.CodeSuccess, constant.MsgCommentDeleteSuccess, gin.H{
		"list": buildCommentResponseList(remainingComments, categoryNames),
	})
}
