package utils

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"
)

// ErrSummaryNotPDF 文档 URL 不是 PDF。
var ErrSummaryNotPDF = errors.New("not a pdf url")

// ErrSummaryEmptyText 未能从 PDF 解析出有效文本。
var ErrSummaryEmptyText = errors.New("empty pdf text")

// DocumentURLPathLooksLikePDF 根据 URL 路径判断是否按 PDF 处理（支持带 query）。
func DocumentURLPathLooksLikePDF(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Path == "" {
		return strings.HasSuffix(strings.ToLower(strings.TrimSpace(raw)), ".pdf")
	}
	return strings.HasSuffix(strings.ToLower(u.Path), ".pdf")
}

// TruncateRunesForSummary 按 rune 截断正文并附加说明，避免超出模型上下文。
func TruncateRunesForSummary(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "\n\n【说明：正文过长，已截断后续部分；摘要仅基于以上片段。】"
}

// ExtractDocumentPDFPlainText 下载 PDF、抽取并清洗正文（与 RAG 学习流程一致），返回截断后的纯文本。
func ExtractDocumentPDFPlainText(documentURL string, maxRunes int) (string, error) {
	if !DocumentURLPathLooksLikePDF(documentURL) {
		return "", ErrSummaryNotPDF
	}
	tmpPath, err := DownloadFromCOSToTemp(documentURL)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpPath)

	rawText, err := ExtractTextFromPDF(tmpPath)
	if err != nil {
		return "", err
	}
	cleaned := CleanText(rawText)
	if strings.TrimSpace(cleaned) == "" {
		return "", ErrSummaryEmptyText
	}
	return TruncateRunesForSummary(cleaned, maxRunes), nil
}
