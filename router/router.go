package router

import (
	"github.com/antidote-kt/SSE_Library-back/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	api.POST("/upload", controllers.UploadFile)
	return router
}
