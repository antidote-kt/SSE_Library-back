package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/spf13/viper"
	"github.com/tencentyun/cos-go-sdk-v5"
)

var (
	cosClient *cos.Client
	bucketURL string
	once      sync.Once
)

// 初始化COS客户端（单例模式）
func initCOSClient() {
	bucketName := viper.GetString("bucket.bucketName")
	appID := viper.GetString("bucket.appID")
	region := viper.GetString("bucket.region")
	domain := viper.GetString("bucket.domain")
	secretID := viper.GetString("bucket.secretID")
	secretKey := viper.GetString("bucket.secretKey")

	bucketURL = fmt.Sprintf("http://%s-%s.cos.%s.%s", bucketName, appID, region, domain)
	u, _ := url.Parse(bucketURL)
	b := &cos.BaseURL{BucketURL: u}
	cosClient = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
}

// 获取COS客户端实例
func getCOSClient() (*cos.Client, string) {
	once.Do(initCOSClient)
	return cosClient, bucketURL
}
func UploadFile(key string, file io.Reader) (string, error) {
	// 对象键（Key）是对象在存储桶中的唯一标识。
	client, bucketURL := getCOSClient()
	_, err := client.Object.Put(context.Background(), key, file, nil)
	return bucketURL + "/" + key, err
}

func DeleteFile(filename string) error {
	client, _ := getCOSClient()
	_, err := client.Object.Delete(context.Background(), filename)
	return err
}

// GetFileURL filename：文件在 COS 中的路径（如 "images/avatar.jpg"） 返回文件的完整 HTTP/HTTPS 访问地址
func GetFileURL(filename string) string {
	client, _ := getCOSClient()
	oUrl := client.Object.GetObjectURL(filename)
	return oUrl.String()
}
