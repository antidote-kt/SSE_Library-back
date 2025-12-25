package constant

const (
	MaxFileSize = 10 << 24 // 最大文件大小 (160MB)
)
const (
	DocumentStatusPending   = "pending"   // 审核中
	DocumentStatusOpen      = "open"      // 开放
	DocumentStatusClosed    = "closed"    // 关闭
	DocumentStatusWithdrawn = "withdrawn" // 已撤回
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
