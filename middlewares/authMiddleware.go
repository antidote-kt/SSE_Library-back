package middlewares

import (
	"strconv"
	"strings"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(constant.StatusUnauthorized, gin.H{
				"code":    constant.CodeError,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// 检查Bearer token格式（支持有空格和无空格两种格式）
		var token string
		if strings.HasPrefix(authHeader, "Bearer ") {
			// 标准格式：Bearer token
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 {
				c.JSON(constant.StatusUnauthorized, gin.H{
					"code":    constant.CodeError,
					"message": "Unauthorized",
				})
				c.Abort()
				return
			}
			token = tokenParts[1]
		} else if strings.HasPrefix(authHeader, "Bearer") {
			// 无空格格式：Bearertoken
			token = strings.TrimPrefix(authHeader, "Bearer")
		} else {
			c.JSON(constant.StatusUnauthorized, gin.H{
				"code":    constant.CodeError,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// TODO: 这里应该验证JWT token的有效性
		// 目前为了测试，我们简单验证token不为空
		if token == "" {
			c.JSON(constant.StatusUnauthorized, gin.H{
				"code":    constant.CodeError,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// TODO: 从JWT token中解析用户信息
		// 目前为了测试，我们从token中提取用户ID（如果token格式为 "token_123"）
		userID := extractUserIDFromToken(token)
		if userID == 0 {
			c.JSON(constant.StatusUnauthorized, gin.H{
				"code":    constant.CodeError,
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("username", "user_"+strconv.FormatUint(uint64(userID), 10))
		c.Set("user_role", "user")

		c.Next()
	}
}

// extractUserIDFromToken 从token中提取用户ID
// 支持格式: "token_123" -> 123
func extractUserIDFromToken(token string) uint {
	// 如果token包含下划线，尝试提取数字部分
	if strings.Contains(token, "_") {
		parts := strings.Split(token, "_")
		if len(parts) >= 2 {
			if userID, err := strconv.ParseUint(parts[len(parts)-1], 10, 32); err == nil {
				return uint(userID)
			}
		}
	}

	// 如果token是纯数字，直接解析
	if userID, err := strconv.ParseUint(token, 10, 32); err == nil {
		return uint(userID)
	}

	// 默认返回1（为了向后兼容）
	return 1
}

// OptionalAuthMiddleware 可选认证中间件（不强制要求认证）
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// 检查Bearer token格式
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				token := tokenParts[1]
				if token != "" {
					// TODO: 验证JWT token并设置用户信息
					userID := extractUserIDFromToken(token)
					c.Set("user_id", userID)
					c.Set("username", "user_"+strconv.FormatUint(uint64(userID), 10))
					c.Set("user_role", "user")
				}
			}
		}

		c.Next()
	}
}
