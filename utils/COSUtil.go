package utils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
func UploadFile(key string, file io.Reader) error {
	// 对象键（Key）是对象在存储桶中的唯一标识。
	client, _ := getCOSClient()
	_, err := client.Object.Put(context.Background(), key, file, nil)
	return err
}

func DeleteFile(filename string) error {
	if filename == "" {
		return nil
	}
	client, _ := getCOSClient()
	_, err := client.Object.Delete(context.Background(), filename)
	return err
}

// filename：文件在 COS 中的路径（如 "images/avatar.jpg"） 返回文件的完整 HTTP/HTTPS 访问地址
func GetFileURL(filename string) string {
	if filename == "" {
		return ""
	}
	client, _ := getCOSClient()
	oUrl := client.Object.GetObjectURL(filename)
	return oUrl.String()
}

// 上传主文件
func UploadMainFile(file *multipart.FileHeader, category string) (string, error) {
	if file == nil || file.Size == 0 {
		return "", nil
	}
	fileReader, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败")
	}
	defer fileReader.Close()

	// 使用新的路径生成
	secureFilename := generateSecureFilename(file.Filename)
	filePath := fmt.Sprintf("files/%s/%s", category, secureFilename)
	err = UploadFile(filePath, fileReader)
	if err != nil {
		return "", fmt.Errorf("文件上传失败，请稍后重试")
	}

	return filePath, nil
}

// 上传封面图片
func UploadCoverImage(cover *multipart.FileHeader, category string) (string, error) {
	if cover == nil || cover.Size == 0 {
		return "", nil // 没有封面图片，返回空字符串
	}

	coverFile, err := cover.Open()
	if err != nil {
		return "", fmt.Errorf("打开封面文件失败")
	}
	defer coverFile.Close()

	// 使用新的路径生成
	secureFilename := generateSecureFilename(cover.Filename)
	coverPath := fmt.Sprintf("covers/%s/%s", category, secureFilename)
	err = UploadFile(coverPath, coverFile)
	if err != nil {
		return "", fmt.Errorf("封面图片上传失败")
	}

	return coverPath, nil
}

// 生成安全的文件名
func generateSecureFilename(originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d%s", timestamp, ext)
}
