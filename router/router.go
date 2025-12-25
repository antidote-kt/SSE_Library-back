package router

import (
	"github.com/antidote-kt/SSE_Library-back/controllers"
	"github.com/antidote-kt/SSE_Library-back/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.CORSMiddleware())
	api := router.Group("/api")

	api.POST("/login", controllers.Login)
	api.POST("/register", controllers.RegisterUser)
	api.POST("/VCode", controllers.SendVerificationCode) // 请求验证码（临时生成，采用POST）
	api.PUT("/Password", controllers.ChangePassword)     //修改密码
	api.GET("/ws", controllers.ConnectWS)                // WebSocket连接

	// --- 需要认证才能访问的路由 ---
	authed := api.Group("/")
	authed.Use(middlewares.AuthMiddleware())
	{
		// 通用接口
		authed.GET("/comment/:commentId", controllers.GetSingleComment)  // 获取单条评论
		authed.GET("/user/:user_id", controllers.GetProfile)             //查看个人主页
		authed.PUT("/user/:user_id", controllers.ModifyInfo)             //修改个人资料
		authed.GET("/document/:id", controllers.GetDocumentByID)         // 获取文档详情
		authed.GET("/searchdoc", controllers.SearchDocument)             //搜索文档
		authed.GET("/documents", controllers.GetDocumentList)            // 获取文档列表
		authed.PUT("/document", controllers.ModifyDocument)              // 文件信息修改（上传该文件的用户才能修改）
		authed.GET("/chat/messages", controllers.GetChatMessages)        //获取聊天记录
		authed.GET("/chat/search", controllers.SearchChatMessages)       //搜索聊天记录
		authed.GET("/getReminder", controllers.GetNotification)          //获取提醒
		authed.POST("/markReminderRead", controllers.MarkNotification)   //标记提醒为已读
		authed.GET("/category/:categoryId", controllers.GetCategoryDetail) // 获取特定的分类或课程详情（必须在 /category 之前）
		authed.GET("/category", controllers.GetCategoriesAndCourses)     // 获取分类和课程
		authed.GET("/searchcat", controllers.SearchCategoriesAndCourses) // 搜索分类和课程
		authed.PUT("/category", controllers.ModifyCategory)              // 修改分类或课程
		authed.DELETE("/category", controllers.DeleteCategory)           // 删除分类或课程
		authed.POST("/category", controllers.AddCategory)                // 添加分类
		authed.POST("/chat/message", controllers.SendMessage)            // 发送消息
		authed.GET("/chat/sessions", controllers.GetSessionList)         // 获取当前用户的所有会话列表
		authed.POST("/post", controllers.CreatePost)                     // 发帖
		authed.GET("/getPosts", controllers.GetPostList)                 // 获取帖子列表
		authed.GET("/post/:postId", controllers.GetPostDetail)           // 获取帖子详情
		// 评论相关路由：必须在 /post/:post_id 之前，使用更具体的路径避免路由冲突
		authed.GET("/comments/post/:sourceId", controllers.GetPostComments)         // 获取帖子的评论
		authed.GET("/comments/document/:sourceId", controllers.GetDocumentComments) // 获取文档的评论

		// 用户相关操作
		userApi := authed.Group("/user")
		{
			userApi.POST("/document", controllers.UploadDocument)          // 文档上传（需要解析用户id，逻辑绑定到文档表）
			userApi.DELETE("/withdrawUpload", controllers.WithdrawUpload)  // 文件撤回（谁上传谁能撤回）
			userApi.POST("/collect", controllers.CollectDocumentOrPost)    // 收藏资料
			userApi.DELETE("/collect", controllers.WithdrawCollection)     // 取消收藏
			userApi.GET("/document", controllers.GetUserUploadDocument)    //用户查看上传文件列表
			userApi.POST("/comments", controllers.PostComment)             // 发表评论
			userApi.GET("/:user_id/comments", controllers.GetUserComments) // 用户查看自己的评论
			userApi.DELETE("/comment", controllers.DeleteUserComment)      // 用户删除自己的评论（需要认证）
			userApi.GET("/hotCategories", controllers.GetHotCategories)    // 获取热门分类
			userApi.GET("/checkFavorite", controllers.CheckFavorite)       // 获取收藏列表
		}

		// 管理员相关操作
		adminApi := authed.Group("/admin")
		adminApi.Use(middlewares.AdminCheckMiddleware())
		{
			adminApi.PUT("/user", controllers.UpdateUserStatus)                     //更新用户状态
			adminApi.GET("/user", controllers.GetUsers)                             //搜索单个用户
			adminApi.GET("/usersList", controllers.GetUsers)                        // GetUsers同时支持获取列表和搜索
			adminApi.PUT("/document/status", controllers.AdminModifyDocumentStatus) //管理员修改文档状态
			adminApi.GET("/comments", controllers.GetAllComments)                   // 管理员获取所有评论（需要认证）
			adminApi.DELETE("/comment", controllers.DeleteComment)                  // 管理员删除评论（需要认证）

		}
	}

	return router
}
