package utils

import (
	"errors"
	"time"
)

import (
	"github.com/golang-jwt/jwt/v5"
)

// MyClaims 自定义声明类型 并内嵌jwt.RegisteredClaims
// jwt.RegisteredClaims包含了一些标准声明 (iss, sub, aud, exp, nbf, iat, jti)
type MyClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// mySecret 用于签名的密钥，请务必保密并从配置文件中读取，这里为了演示写成常量
var mySecret = []byte("这是一个非常安全的密钥")

const TokenExpireDuration = time.Hour * 24 // Token有效期，例如24小时

// GenerateToken 生成JWT
func GenerateToken(userID uint64, username, role string) (string, error) {
	// 创建我们自己的声明
	claims := MyClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                          // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                          // 生效时间
			Issuer:    "sse-library",                                           // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(mySecret)
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*MyClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return mySecret, nil
	})
	if err != nil {
		return nil, err
	}
	// 对token对象中的Claim进行类型断言
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("无效的Token")
}
