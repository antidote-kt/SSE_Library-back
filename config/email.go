package config

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/jordan-wright/email"
	"github.com/spf13/viper"
)

var EmailPool *email.Pool

func InitEmail() {
	host := viper.GetString("email.smtp_host")
	port := viper.GetInt("email.smtp_port")
	username := viper.GetString("email.username")
	password := viper.GetString("email.password")

	var err error
	// 创建一个连接池，提高邮件发送效率
	// 参数：服务器地址, 并发数, 认证信息
	EmailPool, err = email.NewPool(
		fmt.Sprintf("%s:%d", host, port),
		3, // 后续可以再调整
		smtp.PlainAuth("", username, password, host),
	)

	if err != nil {
		log.Fatalf("初始化邮件服务失败: %v", err)
	}
	log.Println("邮件服务初始化成功!")
}

func GetEmailPool() *email.Pool {
	return EmailPool
}
