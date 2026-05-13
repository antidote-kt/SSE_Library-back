package response

// AISummaryData 查看/生成 AI 摘要接口返回的 data 结构（与前端约定字段名一致）。
type AISummaryData struct {
	FromCache   bool   `json:"fromcache"`
	ContentType string `json:"contentType"`
	ContentID   uint64 `json:"contentId"`
	SummaryID   uint64 `json:"summaryId"`
	Summary     string `json:"summary"`
}
