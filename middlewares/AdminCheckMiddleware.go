package middlewares

import (
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

// AdminCheckMiddleware 检查当前登录用户是否为管理员
func AdminCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从AuthMiddleware设置的上下文中获取用户信息
		claims, exists := c.Get("user_claims")
		if !exists {
			// 这种情况理论上不应该发生，因为AuthMiddleware会先拦截
			response.Fail(c, http.StatusUnauthorized, nil, constant.UserNonLogin)
			c.Abort()
			return
		}

		// 2. 类型断言，获取自定义的Claims
		userClaims, ok := claims.(*utils.MyClaims)
		if !ok {
			response.Fail(c, http.StatusUnauthorized, nil, constant.InvalidToken)
			c.Abort()
			return
		}

		// 3. 核心：检查用户角色是否为 "admin"
		if userClaims.Role != "admin" {
			// 如果不是管理员，返回 403 Forbidden 错误
			response.Fail(c, http.StatusForbidden, nil, constant.NoPermission)
			c.Abort()
			return
		}

		// 4. 如果是管理员，则放行
		c.Next()
	}
}
