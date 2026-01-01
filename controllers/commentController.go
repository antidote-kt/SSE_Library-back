package controllers

import (
	"errors"
	"log"
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
		document, err := dao.GetDocumentByID(sourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.MsgRecordNotFound)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}
		// 检查文档状态是否为open，如果不是open状态则不能评论
		if document.Status != constant.DocumentStatusOpen {
			response.Fail(c, http.StatusForbidden, nil, constant.DocumentNotOpen)
			return
		}
	case "post":
		_, err := dao.GetPostByID(sourceID)
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

	// 将评论结果处理成通知格式并插入通知表
	// 1.检查评论的对象资源类型（是document还是post）
	// 2.根据类型查找资源
	// 3.调用资源（数据模型）的属性值构建通知内容
	if sourceType == "document" {
		document, err := dao.GetDocumentByID(sourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.MsgRecordNotFound)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}

		// 仅当评论者不是文档上传者自己时才通知
		if request.Commenter.UserID != document.UploaderID {
			notification := &models.Notification{
				ReceiverID: document.UploaderID,
				Type:       "comment",
				Content:    "您上传的文档《" + document.Name + "》收到了一条新的评论",
				IsRead:     false,
				SourceID:   sourceID,
				SourceType: sourceType,
			}

			err = dao.CreateNotification(notification)
			if err != nil {
				log.Println("创建通知失败:", err) // 通知创建失败不影响评论本身发表，因此这里只打印日志而不返回错误
			}

			// 此时 notification.ID 已经被 GORM 自动填充了
			// CreateNotification函数中，GORM 的 `Create` 方法接收一个接口（通常是指向结构体的指针）。
			// 它会生成 SQL 插入语句，并在执行后，利用数据库驱动（如 `database/sql`）的能力获取 `LastInsertId`，
			// 然后反射（Reflect）将这个 ID 赋值回给结构体的主键字段（通常是 `ID`）。因此不需要让函数显式返回 ID，直接用原对象就能拿到了。
			newID := notification.ID

			// WebSocket 实时推送提醒
			// 构建提醒数据格式
			wsData := gin.H{
				"reminderId":   newID,
				"remindertype": notification.Type,
				"content":      notification.Content,
				"sendTime":     notification.CreatedAt,
				"sourceId":     notification.SourceID,
				"sourceType":   notification.SourceType,
			}

			// 将评论信息推送给接收者 (如果在线) ，实现客户端实时接收
			// 调用 WebSocket 管理器发送
			err = utils.WSManager.SendToUser(document.UploaderID, utils.WSMessage{
				Type:       "reminder",
				ReceiverID: document.UploaderID,
				Data:       wsData,
			})
			if err != nil {
				// 实时推送失败，但消息已持久化，接收者下次上线时可通过 GetNotification 拉取新提醒
				// 因此仅打印日志不返回错误
				log.Printf("WS推送给接收者 %d 失败(可能离线): %v", document.UploaderID, err)
			}
		}
	} else { // sourceType == "post"
		post, err := dao.GetPostByID(sourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, constant.MsgRecordNotFound)
				return
			}
			response.Fail(c, http.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
			return
		}

		// 仅当评论者不是帖子发布者自己时才通知
		if request.Commenter.UserID != post.SenderID {
			notification := &models.Notification{
				ReceiverID: post.SenderID,
				Type:       "comment",
				Content:    "您发布的帖子《" + post.Title + "》收到了一条新的评论",
				IsRead:     false,
				SourceID:   sourceID,
				SourceType: sourceType,
			}

			err = dao.CreateNotification(notification)
			if err != nil {
				log.Println("创建通知失败:", err) // 通知创建失败不影响评论本身发表，因此这里只打印日志而不返回错误
			}

			// 此时 notification.ID 已经被 GORM 自动填充了
			// CreateNotification函数中，GORM 的 `Create` 方法接收一个接口（通常是指向结构体的指针）。
			// 它会生成 SQL 插入语句，并在执行后，利用数据库驱动（如 `database/sql`）的能力获取 `LastInsertId`，
			//然后反射（Reflect）将这个 ID 赋值回给结构体的主键字段（通常是 `ID`）。因此不需要让函数显式返回 ID，直接用原对象就能拿到了。
			newID := notification.ID

			// WebSocket 实时推送提醒
			// 构建提醒数据格式
			wsData := gin.H{
				"reminderId":   newID,
				"remindertype": notification.Type,
				"content":      notification.Content,
				"sendTime":     notification.CreatedAt,
				"sourceId":     notification.SourceID,
				"sourceType":   notification.SourceType,
			}

			// 将评论信息推送给接收者 (如果在线) ，实现客户端实时接收
			// 调用 WebSocket 管理器发送
			err = utils.WSManager.SendToUser(post.SenderID, utils.WSMessage{
				Type:       "reminder",
				ReceiverID: post.SenderID,
				Data:       wsData,
			})
			if err != nil {
				// 实时推送失败，但消息已持久化，接收者下次上线时可通过 GetNotification 拉取新提醒
				// 因此仅打印日志不返回错误
				log.Printf("WS推送给接收者 %d 失败(可能离线): %v", post.SenderID, err)
			}
		}
	}

	// 如果有父评论，且父评论不是自己发表的，则通知父评论的发表者
	if request.ParentID != nil && *request.ParentID != 0 {
		parentComment, err := dao.GetCommentByID(*request.ParentID) // 前面已经有父评论相关验证，这里直接调取评论数据即可
		if request.Commenter.UserID != parentComment.UserID {
			notification := &models.Notification{
				ReceiverID: parentComment.UserID,
				Type:       "comment",
				Content:    "您发表的评论 “" + parentComment.Content + "” 收到了一条新的回复，回复内容：“" + request.Content + "”",
				IsRead:     false,
				SourceID:   sourceID,
				SourceType: sourceType,
			}

			err = dao.CreateNotification(notification)
			if err != nil {
				log.Println("创建通知失败:", err) // 通知创建失败不影响评论本身发表，因此这里只打印日志而不返回错误
			}

			// 此时 notification.ID 已经被 GORM 自动填充了
			// CreateNotification函数中，GORM 的 `Create` 方法接收一个接口（通常是指向结构体的指针）。
			// 它会生成 SQL 插入语句，并在执行后，利用数据库驱动（如 `database/sql`）的能力获取 `LastInsertId`，
			// 然后反射（Reflect）将这个 ID 赋值回给结构体的主键字段（通常是 `ID`）。因此不需要让函数显式返回 ID，直接用原对象就能拿到了。
			newID := notification.ID

			// WebSocket 实时推送提醒
			// 构建提醒数据格式
			wsData := gin.H{
				"reminderId":   newID,
				"remindertype": notification.Type,
				"content":      notification.Content,
				"sendTime":     notification.CreatedAt,
				"sourceId":     notification.SourceID,
				"sourceType":   notification.SourceType,
			}

			// 将评论信息推送给接收者 (如果在线) ，实现客户端实时接收
			// 调用 WebSocket 管理器发送
			err = utils.WSManager.SendToUser(parentComment.UserID, utils.WSMessage{
				Type:       "reminder",
				ReceiverID: parentComment.UserID,
				Data:       wsData,
			})
			if err != nil {
				// 实时推送失败，但消息已持久化，接收者下次上线时可通过 GetNotification 拉取新提醒
				// 因此仅打印日志不返回错误
				log.Printf("WS推送给接收者 %d 失败(可能离线): %v", parentComment.UserID, err)
			}
		}
	}

	// 如果是帖子类型，更新帖子的评论数
	if sourceType == "post" {
		if err := dao.IncrementPostCommentCount(sourceID); err != nil {
			log.Printf("更新帖子评论数失败: %v", err)
			// 这里不返回错误，因为评论已经创建成功，只是统计更新失败
		}
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
	// 验证管理员身份
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	if userClaims.Role != "admin" {
		response.Fail(c, http.StatusForbidden, nil, constant.NoPermission)
		return
	}

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
	// 验证管理员身份
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	if userClaims.Role != "admin" {
		response.Fail(c, http.StatusForbidden, nil, constant.NoPermission)
		return
	}

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

	comment, err := dao.GetCommentByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, constant.StatusNotFound, nil, constant.MsgCommentNotFound)
			return
		}
		response.Fail(c, constant.StatusInternalServerError, nil, constant.MsgDatabaseQueryFailed)
		return
	}

	// 如果是帖子类型，先更新帖子的评论数（在删除前）
	if comment.SourceType == "post" {
		if err := dao.DecrementPostCommentCount(comment.SourceID); err != nil {
			log.Printf("更新帖子评论数失败: %v", err)
			// 这里不返回错误，继续执行删除操作
		}
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

	// 如果是帖子类型，先更新帖子的评论数（在删除前）
	if comment.SourceType == "post" {
		if err := dao.DecrementPostCommentCount(comment.SourceID); err != nil {
			log.Printf("更新帖子评论数失败: %v", err)
			// 这里不返回错误，继续执行删除操作
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
