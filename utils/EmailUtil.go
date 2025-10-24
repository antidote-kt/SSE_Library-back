package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/jordan-wright/email"
	"github.com/spf13/viper"
)

// SendVerificationEmail 发送验证码邮件
func SendVerificationEmail(toEmail, code string) error {
	pool := config.GetEmailPool()
	from := viper.GetString("email.from_address")

	e := email.NewEmail()
	e.From = fmt.Sprintf("SSE Library <%s>", from)
	e.To = []string{toEmail}
	e.Subject = "您的SSE Library验证码"
	e.HTML = []byte(fmt.Sprintf(`
        <p>您好！</p>
        <p>您的验证码是：<b>%s</b></p>
        <p>此验证码将在5分钟内有效，请尽快使用。</p>
        <p>如果您没有请求此验证码，请忽略此邮件。</p>
        <br>
        <p>SSE Library 团队</p>
    `, code)) // 邮件HTML模板

	// 使用连接池发送邮件
	err := pool.Send(e, 10*time.Second) // 设置10秒超时
	if err != nil {
		log.Printf("发送邮件到 %s 失败: %v", toEmail, err)
		return fmt.Errorf("发送邮件失败")
	}

	log.Printf("验证码邮件已成功发送到 %s", toEmail)
	return nil
}
