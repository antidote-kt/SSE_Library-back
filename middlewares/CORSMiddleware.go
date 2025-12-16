package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 动态获取请求的 Origin 头部
		origin := ctx.Request.Header.Get("Origin")
		if origin != "" {
			// 将 Allow-Origin 设置为请求方的域名，而不是 "*"
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// 如果没有 Origin 头（比如非浏览器请求），可以设为 * 或者保持原样
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		ctx.Writer.Header().Set("Access-Control-Max-Age", "86400")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		ctx.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(200)
		} else {
			ctx.Next()
		}
	}
}
