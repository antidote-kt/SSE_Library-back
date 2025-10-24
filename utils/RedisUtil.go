package utils

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/go-redis/redis/v8"
)

const VerificationCodeExpireDuration = time.Minute * 5 // 验证码有效期5分钟

// GenerateVerificationCode 生成指定长度的随机数字验证码
func GenerateVerificationCode(length int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) // 随机数生成器

	var sb strings.Builder
	for i := 0; i < length; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rng.Intn(r)])
	}
	return sb.String()
}

// StoreVerificationCode 将验证码存入Redis
func StoreVerificationCode(email, usage, code string) error {
	rdb := config.GetRedisClient()
	key := fmt.Sprintf("verify_code:%s:%s", usage, email) // Key格式: verify_code:业务:邮箱
	log.Printf("存储验证码，生成的Key: %s", key)                   // 添加日志
	return rdb.Set(config.Ctx, key, code, VerificationCodeExpireDuration).Err()
}

// CheckVerificationCode 从Redis校验验证码
func CheckVerificationCode(email, usage, codeToCheck string) (bool, error) {
	rdb := config.GetRedisClient()
	key := fmt.Sprintf("verify_code:%s:%s", usage, email)
	log.Printf("校验验证码，查找的Key: %s", key) // 添加日志
	storedCode, err := rdb.Get(config.Ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		// Key不存在，说明验证码过期或从未发送
		return false, nil // 返回 false, 无错误
	} else if err != nil {
		// 其他Redis错误
		return false, err
	}

	// 校验成功后立即删除验证码，防止重复使用
	if storedCode == codeToCheck {
		rdb.Del(config.Ctx, key) // 忽略删除错误
		return true, nil
	}

	return false, nil // 验证码不匹配
}
