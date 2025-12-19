package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"

	"github.com/spf13/viper"
)

// SendVerificationEmail 发送验证码邮件（原生smtp，不用email库）
func SendVerificationEmail(toEmail, code string) error {
	// 1. 获取配置
	host := "smtp.qq.com"
	port := 465
	username := viper.GetString("email.username")
	password := viper.GetString("email.password")
	from := viper.GetString("email.from_address")

	// 2. 组装邮件内容 (包括头部)
	// 注意：原生 smtp 需要手动写 Header，否则邮件没有标题
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("SSE Library <%s>", from)
	headers["To"] = toEmail
	headers["Subject"] = "SSE Library 注册验证码"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + fmt.Sprintf(`
        <div style="padding: 20px;">
            <p>您的验证码是：<b style="font-size: 24px; color: #007bff;">%s</b></p>
            <p>5分钟内有效。如非本人操作，请忽略。</p>
        </div>
    `, code)

	// 3. 建立 SSL 连接 (强制使用 465)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), tlsConfig)
	if err != nil {
		log.Printf("SSL 连接失败: %v", err)
		return err
	}
	defer func(conn *tls.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// 4. 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Printf("创建 SMTP 客户端失败: %v", err)
		return err
	}

	// 5. 认证
	auth := smtp.PlainAuth("", username, password, host)
	if err = client.Auth(auth); err != nil {
		log.Printf("SMTP 认证失败: %v", err)
		return err
	}

	// 6. 设置发件人和收件人
	if err = client.Mail(from); err != nil {
		return err
	}
	if err = client.Rcpt(toEmail); err != nil {
		return err
	}

	// 7. 写入数据
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	// 关键点：这里关闭写入流，告诉服务器“我说完了”
	// 如果这里没报错，说明邮件内容已经成功传给了服务器
	err = w.Close()
	if err != nil {
		log.Printf("邮件数据传输异常: %v", err)
		// 这里如果报错，则邮件内容没能成功发送，返回错误
		return err
	}

	// 8. 退出 (Quit)
	// QQ 邮箱经常在这里已经断开连接了，导致报错 short response
	// 但既然上面 w.Close() 成功了，邮件就已经发送了。
	// 所以这里我们执行 Quit，但不检查它的错误，或者只打印日志但不返回错误。
	err = client.Quit()

	log.Printf("验证码邮件发送成功 : %s", toEmail)
	return nil
}
