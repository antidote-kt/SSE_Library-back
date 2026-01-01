package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePost 发布帖子接口
// POST /api/post
func CreatePost(c *gin.Context) {
	var req dto.CreatePostDTO

	// 1. 绑定参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 获取当前登录用户身份 (JWT)
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	// 3. 校验发帖人ID是否与当前登录用户一致 (防止替人发帖)
	if req.SenderID != userClaims.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 5. 构建 Post 模型
	post := models.Post{
		SenderID: req.SenderID,
		Title:    req.Title,
		Content:  req.Content,
		// SendTime 由 GORM 的 autoCreateTime 自动处理
	}

	// 6. 调用 DAO 保存数据
	if err := dao.CreatePostWithTx(&post, req.DocumentIDs); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "发帖失败: "+err.Error())
		return
	}

	// 7. 返回成功响应
	// 构造返回数据
	responseData := gin.H{
		"postId": post.ID,
	}
	response.Success(c, responseData, constant.CreatePostSuccess)
}

// GetPostList 获取帖子列表接口
// GET /api/posts
func GetPostList(c *gin.Context) {
	var req dto.GetPostListDTO

	// 1. 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 调用 DAO 获取帖子列表
	posts, err := dao.GetPostList(req.Key, req.Order)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 3. 构建响应数据
	postList := response.BuildPostListResponse(posts)

	// 4. 返回成功响应
	response.SuccessWithData(c, postList, constant.PostsObtain)
}

// GetPostDetail 获取帖子详情接口
func GetPostDetail(c *gin.Context) {
	// 1. 获取帖子ID
	var postIDStr = c.Param("postId")
	postID, _ := strconv.ParseUint(postIDStr, 10, 64)

	// 2. 调用 DAO 获取帖子详情
	post, err := dao.GetPostByID(postID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 3. 记录浏览历史 (异步)
	if claims, exists := c.Get(constant.UserClaims); exists {
		userClaims := claims.(*utils.MyClaims)
		go func(uid uint64, pid uint64) {
			// 传入 "post" 类型
			_ = dao.AddViewHistory(uid, pid, "post")
		}(userClaims.UserID, postID) // 回调函数实现异步
	}

	// 4. 查询帖子相关文档
	postdocs, err := dao.GetDocumentsByPostID(post.ID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 5. 调用response层构建响应数据
	postDetail := response.BuildPostDetailResponse(post, postdocs)

	// 6. 返回成功响应
	response.SuccessWithData(c, postDetail, constant.GetPostDetailSuccess)

}

// DoLikePost 点赞帖子
func DoLikePost(c *gin.Context) {
	// 1. 获取 PostID (Body参数)
	var req dto.LikePostDTO
	// 绑定查询参数
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 从JWT中间件获取用户信息(拿UserID)
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	// 3. 调用 DAO 执行点赞
	err := dao.LikePost(userClaims.UserID, req.PostID)
	if err != nil {
		if err.Error() == "禁止重复点赞" {
			response.Fail(c, http.StatusBadRequest, nil, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, "点赞失败: "+err.Error())
		return
	}

	// 4. 为点赞结果创建通知，并插入通知表，发送websocket通知
	// 4.1 获取点赞的帖子对象以及发帖人
	post, err := dao.GetPostByID(req.PostID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}
	user, err := dao.GetUserByID(post.SenderID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}
	// 仅当点赞人与发帖人本人不同时才创建通知
	if userClaims.UserID != post.SenderID {
		// 4.2 创建通知
		notification := models.Notification{
			ReceiverID: post.SenderID,
			Type:       "like",
			Content:    fmt.Sprintf("你发表的帖子《%s》被用户\"  %s \" 点赞", post.Title, user.Username),
			IsRead:     false,
			SourceID:   post.ID,
			SourceType: constant.PostType,
		}
		err = dao.CreateNotification(&notification)
		if err != nil {
			log.Println("创建通知失败:", err) // 通知创建失败不影响点赞操作本身，因此这里只打印日志而不返回空切片和错误
		}
		// 3.发送websocket通知
		// 构建提醒数据格式
		wsData := gin.H{
			"reminderId":   notification.ID,
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

	// 5. 返回成功响应
	response.Success(c, nil, "点赞成功")
}

// DoUnlikePost 取消点赞
func DoUnlikePost(c *gin.Context) {
	// 1. 获取 PostID (Body参数)
	var req dto.LikePostDTO
	// 绑定查询参数
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 从JWT中间件获取用户信息(拿UserID)
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	// 2. 调用 DAO
	err := dao.UnlikePost(userClaims.UserID, req.PostID)
	if err != nil {
		if err.Error() == "not liked yet" {
			response.Fail(c, http.StatusBadRequest, nil, "您尚未点赞，无法取消")
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, "取消点赞失败")
		return
	}

	response.Success(c, nil, "取消点赞成功")
}

// GetPostLikeStatus 查询当前用户是否点赞
func GetPostLikeStatus(c *gin.Context) {
	// 1. 获取 PostID (Body参数)
	var req dto.LikePostDTO
	// 绑定查询参数
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 从JWT中间件获取用户信息(拿UserID)
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	// 调用 DAO
	isLiked, err := dao.IsPostLikedByUser(userClaims.UserID, req.PostID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	response.Success(c, gin.H{
		"isLiked": isLiked,
		"postId":  req.PostID,
	}, "获取点赞状态成功")
}

// GetUserPostList 获取用户的帖子列表（包括收藏的帖子和自己发布的帖子）
// GET /api/user/postList/:userId
func GetUserPostList(c *gin.Context) {
	// 1. 获取路径参数中的userId
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDFormatError)
		return
	}

	// 2. 获取当前登录用户身份 (JWT)
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	// 3. 验证请求的用户ID是否与JWT中的用户ID一致（防止越权操作）
	if userClaims.UserID != userID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 4. 验证用户是否存在
	_, err = dao.GetUserByID(userID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
		return
	}

	// 5. 获取用户收藏的帖子列表
	collectPosts, err := dao.GetFavoritePostsByUserID(userID)

	// 6. 获取用户发布的帖子列表
	myPosts, err := dao.GetPostsByUserID(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 7. 构建响应数据
	userPostListResponse := response.BuildUserPostListResponse(collectPosts, myPosts)

	// 8. 返回成功响应
	response.SuccessWithData(c, userPostListResponse, constant.PostsObtain)
}

// DeletePost 删除帖子接口
// DELETE /api/post
func DeletePost(c *gin.Context) {
	// 1. 绑定请求参数
	var req dto.DeletePostDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 验证 PostID 是否有效
	if req.PostID == 0 {
		response.Fail(c, http.StatusBadRequest, nil, constant.PostIDNotNull)
		return
	}
	postID := req.PostID

	// 3. 获取当前登录用户身份 (JWT)
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	// 4. 验证帖子是否存在，并检查是否为发帖人
	post, err := dao.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.PostNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 5. 验证是否为发帖人（只有发帖人才能删除）
	if post.SenderID != userClaims.UserID {
		response.Fail(c, http.StatusForbidden, nil, constant.PostDeleteNotAllowed)
		return
	}

	// 6. 调用 DAO 删除帖子（包括评论和文档关联）
	err = dao.DeletePostWithTx(postID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DeletePostFailed)
		return
	}

	// 7. 返回成功响应
	response.Success(c, nil, constant.DeletePostSuccess)
}
