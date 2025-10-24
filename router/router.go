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

	// --- 需要认证才能访问的路由 ---
	authed := api.Group("/")
	authed.Use(middlewares.AuthMiddleware())
	{
		// 通用接口
		authed.GET("/:document_id/comments", controllers.GetComments) // 获取对某书的评论列表
		authed.GET("/user/:user_id", controllers.GetProfile)          //查看个人主页
		authed.PUT("/user/:user_id", controllers.ModifyInfo)          //修改个人资料

		// 用户相关操作
		userApi := authed.Group("/user")
		{
			userApi.POST("/document", controllers.UploadFile)               // 文档上传（需要解析用户id，逻辑绑定到文档表）
			userApi.DELETE("/withdrawUpload", controllers.WithdrawUpload)   // 文件撤回（谁上传谁能撤回）
			userApi.PUT("/document", controllers.ModifyDocument)            // 文件信息修改（上传该文件的用户才能修改）
			userApi.POST("/collect", controllers.CollectDocument)           // 收藏资料
			userApi.DELETE("/collect", controllers.WithdrawCollection)      // 取消收藏
			userApi.GET("/document", controllers.GetUserUploadDocument)     //用户查看上传文件列表
			userApi.POST("/:document_id/comments", controllers.PostComment) // 发表评论
			userApi.GET("/:user_id/comments", controllers.GetUserComments)  // 用户查看自己的评论
			userApi.DELETE("/comment", controllers.DeleteUserComment)       // 用户删除自己的评论（需要认证）
		}

		// 管理员相关操作
		adminApi := authed.Group("/admin")
		adminApi.Use(middlewares.AdminCheckMiddleware())
		{
			adminApi.PUT("/user", controllers.UpdateUserStatus) //更新用户状态
			adminApi.GET("/user", controllers.GetUsers)
			adminApi.GET("/usersList", controllers.GetUsers) // GetUsers同时支持获取列表和搜索
			adminApi.PUT("/file", controllers.AdminModifyDocumentStatus)
			adminApi.POST("/fileStatus", controllers.AdminModifyDocumentStatus)
			adminApi.GET("/comments", controllers.GetAllComments)  // 管理员获取所有评论（需要认证）
			adminApi.DELETE("/comment", controllers.DeleteComment) // 管理员删除评论（需要认证）

		}
	}

	return router
}
