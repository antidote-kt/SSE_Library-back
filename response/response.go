package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Response(ctx *gin.Context, httpStatus int, code int, data gin.H, msg string) {
	ctx.JSON(httpStatus, gin.H{"code": code, "data": data, "message": msg})
}

func Success(ctx *gin.Context, data gin.H, msg string) {
	Response(ctx, http.StatusOK, http.StatusOK, data, msg)
}
func Fail(ctx *gin.Context, httpStatus int, data gin.H, msg string) {
	Response(ctx, httpStatus, httpStatus, data, msg)
}
