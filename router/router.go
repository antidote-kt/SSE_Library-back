package router

import (
	"github.com/antidote-kt/SSE_Library-back/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	userApi := api.Group("/users")
	adminApi := api.Group("/admin")

	api.POST("/upload", controllers.UploadFile)
	api.PUT("/modify", controllers.ModifyDocument)

	userApi.DELETE("/withdrawUpload", controllers.WithdrawUpload)

	adminApi.PUT("/file", controllers.AdminModifyDocument)
	adminApi.POST("/fileStatus", controllers.AdminModifyDocumentStatus)
	return router
}
