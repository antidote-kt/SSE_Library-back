package constant

// 公共响应常量
const (
	ParamParseError  = "参数解析失败"
	DatabaseError    = "数据库错误"
	UploaderNotExist = "上传者不存在"
	TimeFormatError  = "时间格式错误"
)

// 收藏相关常量
const (
	FavoriteSuccessMsg        = "收藏成功"
	UnfavoriteSuccessMsg      = "取消收藏成功"
	FavoriteAlreadyExistsMsg  = "文档已收藏"
	FavoriteNotExistMsg       = "文档未被收藏"
	FavoriteListSuccessMsg    = "获取收藏列表成功"
	FavoriteStatusCheckFailed = "检查收藏状态失败"
	FavoriteCreateFailed      = "创建收藏记录失败"
	FavoriteDeleteFailed      = "删除收藏记录失败"
)

// 用户上传文档相关常量
const (
	GetUserUploadSuccessMsg  = "获取用户上传文档成功"
	WithdrawUploadSuccessMsg = "撤回成功"
)

// 文档相关常量
const (
	DocumentNotExist            = "文档不存在"
	DocumentCreateFail          = "文档创建失败"
	DocumentCreateSuccess       = "文档创建成功"
	DocumentStatusUpdateSuccess = "文档状态更新成功"
	DocumentStatusUpdateFailed  = "文档状态更新失败"
	DocumentUpdateSuccess       = "文档更新成功"
	DocumentUpdateFail          = "文档更新失败"
	OldCoverDeleteFailed        = "旧封面删除失败"
	CoverUploadFailed           = "封面上传失败"
	OldFileDeleteFailed         = "旧文件删除失败"
	FileUploadFailed            = "文件上传失败"
	OldTagDeleteFailed          = "标签删除失败"
	CollectionUpdateFailed      = "更新文档收藏数失败"
	GetFavoriteDocumentFailed   = "获取收藏文档列表失败"
	DefaultAuthor               = "佚名"
	NotAllowWithdrawOthers      = "不允许撤回其他人的文档"
	NotAllowWithdraw            = "文档不在审核中，不允许撤回"
	DocumentObtain              = "文档获取成功"
)

// 分类相关常量
const (
	CategoryNotExist = "分类不存在"
)

// 用户相关常量
const (
	UserNotExist      = "用户不存在"
	UserIDLack        = "缺少userId参数"
	GetUserInfoFailed = "无法获取用户信息，请重新登录"
	NonSelf           = "非用户本人，不允许访问"
)

// 评论失败相关常量
const (
	MsgDocumentIDFormatError      = "文档ID格式错误"
	MsgUserIDFormatError          = "用户ID格式错误"
	MsgCommentIDFormatError       = "评论ID格式错误"
	MsgRecordNotFound             = "记录不存在"
	MsgUserNotFound               = "用户不存在"
	MsgCommentNotFound            = "评论不存在"
	MsgCommentNotFoundOrNoAccess  = "评论不存在或无权限删除"
	MsgDatabaseQueryFailed        = "数据库查询失败"
	MsgDatabaseOperationFailed    = "数据库操作失败"
	MsgParameterError             = "参数错误"
	MsgContentEmpty               = "评论内容不能为空"
	MsgUserIDEmpty                = "用户ID不能为空"
	MsgCommentIDEmpty             = "评论ID不能为空"
	MsgUnauthorized               = "用户不存在，未认证"
	MsgCommentCreateFailed        = "评论创建失败"
	MsgCommentDeleteFailed        = "删除评论失败"
	MsgGetCommentListFailed       = "获取评论列表失败"
	MsgUserInfoMismatch           = "用户名与数据库不一致"
	MsgParentCommentNotFound      = "父评论不存在"
	MsgParentCommentNotInDocument = "父评论不属于该文档"
	MsgCreateTimeFormatError      = "创建时间格式错误"
)

// 评论成功相关常量
const (
	MsgCommentPostSuccess     = "评论发表成功"
	MsgCommentDeleteSuccess   = "删除评论成功"
	MsgGetCommentListSuccess  = "获取评论列表成功"
	MsgGetAllCommentsSuccess  = "获取所有评论成功"
	MsgGetUserCommentsSuccess = "获取用户评论成功"
)
