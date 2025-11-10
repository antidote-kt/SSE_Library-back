package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Response(ctx *gin.Context, httpStatus int, code int, msg string, data gin.H) {
	ctx.JSON(httpStatus, gin.H{"code": code, "message": msg, "data": data})
}

func Success(ctx *gin.Context, data gin.H, msg string) {
	Response(ctx, http.StatusOK, http.StatusOK, msg, data)
}

func SuccessWithData(ctx *gin.Context, data interface{}, msg string) {
	// 直接返回包含数组的完整响应格式
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": msg,
		"data":    data,
	})
}

func Fail(ctx *gin.Context, httpStatus int, data gin.H, msg string) {
	Response(ctx, httpStatus, httpStatus, msg, data)
}
