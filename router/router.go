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

	api.POST("/document", controllers.UploadFile)
	api.PUT("/document", controllers.ModifyDocument)

	userApi.DELETE("/withdrawUpload", controllers.WithdrawUpload)

	adminApi.PUT("/document", controllers.AdminModifyDocument)
	adminApi.PUT("/document/status", controllers.AdminModifyDocumentStatus)
	return router
}
