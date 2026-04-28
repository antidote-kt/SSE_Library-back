package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

// 下载 COS 文件到本地临时文件
func DownloadFromCOSToTemp(cosUrl string) (string, error) {
	// 1. 发起 HTTP GET 请求
	resp, err := http.Get(cosUrl)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 2. 检查 HTTP 状态码，防止下载到了 403/404 的错误页面代码
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("文件下载失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 3. 创建临时文件
	tmpFile, err := os.CreateTemp("", "kb-*.pdf")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer tmpFile.Close()

	// 4. 将网络流写入文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("写入本地文件失败: %v", err)
	}

	return tmpFile.Name(), nil
}

// 提取 PDF 文本
func ExtractTextFromPDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}

// CleanText 对原始文本进行预清洗，提高 Embedding 精准度
var (
	// 匹配两个及以上的换行符
	reMultiNewline = regexp.MustCompile(`\n{2,}`)
	// 匹配两个及以上的空格/空白符（不包括换行）
	reMultiSpace = regexp.MustCompile(`[ \t]{2,}`)
)

func CleanText(text string) string {
	// 1. 去除首尾空格
	text = strings.TrimSpace(text)

	// 2. 将多个连续换行替换为单个换行
	// PDF 解析经常会在段落间产生大量 \n，压缩它们有助于保持语义连贯
	text = reMultiNewline.ReplaceAllString(text, "\n")

	// 3. 将多个连续的空格或制表符替换为单个空格
	text = reMultiSpace.ReplaceAllString(text, " ")

	// 4. (可选) 去除掉一些 PDF 常见的杂质，如控制字符
	// text = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(text, "")

	return text
}

// 文本切片 (滑动窗口算法，保留上下文)
// chunkSize: 每个分块的最大字符数 (如 500)
// overlap: 相邻分块的重叠字符数 (如 50)
func ChunkText(text string, chunkSize, overlap int) []string {
	// rune 类型是 int32 的别名，表示一个 Unicode 码点（可以正确处理多字节字符）
	// 将以字节为单位的文本字符串（字节流）转化为以字符为单位的文本字符串（字符流），保证按字符切片。
	runes := []rune(text)
	var chunks []string
	length := len(runes)

	for i := 0; i < length; i += chunkSize - overlap {
		end := i + chunkSize
		if end > length {
			end = length
		}
		chunks = append(chunks, string(runes[i:end]))
		if end == length {
			break
		}
	}
	return chunks
}
