package middlewares

import (
	"net/http"
	"strings"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 创建一个JWT认证的中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization header
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			response.Fail(c, http.StatusUnauthorized, nil, constant.RequestWithoutToken)
			c.Abort() // 阻止请求继续
			return
		}

		// 2. 按空格分割，验证Token格式是否为 "Bearer [token]"
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Fail(c, http.StatusUnauthorized, nil, constant.TokenFormatError)
			c.Abort()
			return
		}

		// 3. parts[1] 是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, nil, err.Error())
			c.Abort()
			return
		}

		// 4. 中间件验证完将当前请求的用户信息保存到请求的上下文c上，后续的处理函数就可以通过c.Get("user_claims")来获取当前请求的用户信息
		// claims := c.Get("user_claims") 之后还需要 userClaims := claims.(*utils.MyClaims)来将接口类型转换为具体的声明结构体
		// 此时就可对userClaims进行点下标访问用户信息
		c.Set("user_claims", claims)

		c.Next() // 放行，处理后续的请求
	}
}
