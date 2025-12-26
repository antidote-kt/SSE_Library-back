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
	var postIDstr = c.Param("postId")
	postID, _ := strconv.ParseUint(postIDstr, 10, 64)

	// 2. 调用 DAO 获取帖子详情
	post, err := dao.GetPostByID(postID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 3. 查询帖子相关文档
	postdocs, err := dao.GetDocumentsByPostID(post.ID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 4. 调用response层构建响应数据
	postDetail := response.BuildPostDetailResponse(post, postdocs)

	// 5. 返回成功响应
	response.SuccessWithData(c, postDetail, constant.GetPostDetailSuccess)

}
