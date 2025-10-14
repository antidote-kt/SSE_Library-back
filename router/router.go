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
	userApi := api.Group("/users")
	adminApi := api.Group("/admin")
	booksApi := api.Group("/books")
	userDeleteApi := router.Group("/user")
	apiUserDeleteApi := api.Group("/user") // 支持 /api/user/deleteComment 路径

	api.POST("/document", controllers.UploadFile)
	api.PUT("/document", controllers.ModifyDocument)

	userApi.DELETE("/withdrawUpload", controllers.WithdrawUpload)

	adminApi.PUT("/document", controllers.AdminModifyDocument)
	adminApi.PUT("/document/status", controllers.AdminModifyDocumentStatus)
	adminApi.GET("/comments", middlewares.AuthMiddleware(), controllers.GetAllComments)  // 管理员获取所有评论（需要认证）
	adminApi.DELETE("/comment", middlewares.AuthMiddleware(), controllers.DeleteComment) // 管理员删除评论（需要认证）

	// 评论相关路由
	booksApi.POST("/:document_id/comments", middlewares.AuthMiddleware(), controllers.PostComment)         // 发表评论（需要认证）
	booksApi.GET("/:document_id/comments", controllers.GetComments)                                        // 获取评论列表（无需认证）
	userApi.GET("/:user_id/comments", middlewares.AuthMiddleware(), controllers.GetUserComments)           // 用户查看自己的评论（需要认证）
	userDeleteApi.DELETE("/deleteComment", middlewares.AuthMiddleware(), controllers.DeleteUserComment)    // 用户删除自己的评论（需要认证）
	apiUserDeleteApi.DELETE("/deleteComment", middlewares.AuthMiddleware(), controllers.DeleteUserComment) // 用户删除自己的评论（支持/api路径）
	return router
}
