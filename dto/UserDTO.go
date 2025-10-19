package dto

// RegisterDTO 定义了用户注册时需要绑定的数据
type RegisterDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=20"`
	Avatar   string `json:"UserAvatar" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// LoginDTO 定义了用户登录时需要绑定的数据
type LoginDTO struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ModifyInfoDTO 定义了用户修改个人资料时需要绑定的数据（PS：查看个人主页请求参数只有路径参数，无需专门结构体解析）
type ModifyInfoDTO struct {
	UserName   *string `json:"userName" binding:"required"`
	UserAvatar *string `json:"userAvatar" binding:"required"`
	Email      *string `json:"email" binding:"required"`
	Password   *string `json:"password" binding:"required"`
}

// HomepageDTO 是用户主页接口返回的完整数据结构
type HomepageDTO struct {
	UserBrief      UserBriefDTO        `json:"userBrief"`
	CollectionList []DocumentDetailDTO `json:"collectionList"`
	HistoryList    []DocumentDetailDTO `json:"historyList"`
}

// UserBriefDTO 定义了用户基本信息
type UserBriefDTO struct {
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	UserAvatar string `json:"userAvatar"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

// DocumentDetailDTO 定义了文档的详细信息，用于列表展示
type DocumentDetailDTO struct {
	InfoBrief    InfoBriefDTO `json:"infoBrief"`
	BookISBN     string       `json:"bookISBN"`
	Author       string       `json:"author"`
	Uploader     UserBriefDTO `json:"uploader"`
	Cover        string       `json:"Cover"`
	Tags         []string     `json:"tags"`
	Introduction string       `json:"introduction"`
	CreateYear   string       `json:"createYear"`
}

// InfoBriefDTO 定义了文档的摘要信息
type InfoBriefDTO struct {
	Name        string `json:"name"`
	DocumentID  uint64 `json:"document_id"`
	Type        string `json:"type"`
	UploadTime  string `json:"uploadTime"`
	Status      string `json:"status"`
	Category    string `json:"category"`
	Course      string `json:"course"`
	Collections int    `json:"collections"`
	ReadCounts  int    `json:"readCounts"`
	URL         string `json:"URL"`
}
