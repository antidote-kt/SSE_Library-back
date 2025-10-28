package constant

const (
	MaxFileSize = 10 << 20 // 最大文件大小 (10MB)
)
const (
	DocumentStatusAudit    = "audit"    // 审核中
	DocumentStatusOpen     = "open"     // 开放
	DocumentStatusClose    = "close"    // 关闭
	DocumentStatusWithdraw = "withdraw" // 已撤回
)

const (
	VideoType = "video"
	FileType  = "file"
	BookType  = "book"
)

const (
	TypeOfKeyName         = "name"
	TypeOfKeyAuthor       = "author"
	TypeOfKeyBookISBN     = "bookISBN"
	TypeOfKeyIntroduction = "introduction"
	TypeOfKeyTag          = "tag"
)
