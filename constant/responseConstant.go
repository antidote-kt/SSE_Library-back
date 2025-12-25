package constant

// 公共响应常量
const (
	ParamParseError     = "参数解析失败"
	DatabaseError       = "数据库错误"
	UploaderNotExist    = "上传者不存在"
	TimeFormatError     = "时间格式错误"
	GetDataFailed       = "获取数据失败"
	ConstructDataFailed = "构建数据失败"
)

// 收藏相关常量
const (
	FavoriteSuccessMsg        = "收藏成功"
	UnfavoriteSuccessMsg      = "取消收藏成功"
	FavoriteAlreadyExistsMsg  = "文档已收藏"
	FavoriteNotExistMsg       = "文档未被收藏"
	FavoriteListSuccessMsg    = "获取收藏列表成功"
	FavoriteGetFailed         = "查询收藏记录失败"
	FavoriteStatusCheckFailed = "检查收藏状态失败"
	FavoriteCreateFailed      = "创建收藏记录失败"
	FavoriteDeleteFailed      = "删除收藏记录失败"
	FavoriteGetSuccess        = "获取收藏记录成功"
	FavoriteTypeNotAllow      = "不支持的收藏类型"
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
	DocumentDeletedFailed       = "文档删除失败"
	DocumentUpdateFail          = "文档更新失败"
	OldCoverDeleteFailed        = "旧封面删除失败"
	OpenDocumentCoverFailed     = "打开文档封面失败"
	UploadCoverImageFailed      = "封面上传失败"
	OldDocumentDeleteFailed     = "旧文档删除失败"
	DocumentOpenFailed          = "打开文档失败"
	DocumentUploadFailed        = "文档上传失败"
	CollectionUpdateFailed      = "更新文档收藏数失败"
	GetFavoriteDocumentFailed   = "获取收藏文档列表失败"
	DefaultAuthor               = "佚名"
	NotAllowWithdrawOthers      = "不允许撤回其他人的文档"
	DocumentNotAllow            = "文档不允许查看"
	DocumentObtain              = "文档获取成功"
	DocumentsObtain             = "文档列表获取成功"
	DocumentTagCreateFailed     = "创建文档标签关联失败"
	DocumentTagGetFailed        = "文档标签关联查询失败"
	DocumentIDLack              = "文档ID不能为空"
	DocumentNotOpen             = "文档状态不是公开的，无法收藏"
)

// Tag相关常量
const (
	TagCreateFailed    = "标签创建失败"
	TagGetFailed       = "标签查询失败"
	OldTagDeleteFailed = "标签删除失败"
)

// 分类相关常量
const (
	CategoryNotExist            = "分类不存在"
	ParentCategoryNotExist      = "父分类不存在"
	CategoryNameAlreadyExist    = "分类名称已存在"
	MsgGetCategoriesSuccess     = "获取分类和课程成功"
	MsgGetCategoriesListFailed  = "获取分类列表失败"
	MsgCategoryCountFailed      = "统计分类文档失败"
	MsgCategoryReadCountFailed  = "统计分类浏览量失败"
	MsgGetHotCategoriesSuccess  = "获取热门分类成功"
	MsgGetHotCategoriesFailed   = "获取热门分类失败"
	MsgCategoryCreateFailed     = "添加分类失败"
	MsgCategoryCreateSuccess    = "添加分类成功"
	MsgCategoryDeleteSuccess    = "删除分类成功"
	MsgCategoryDeleteFailed     = "删除分类失败"
	MsgCategoryNameRequired     = "分类名称不能为空"
	MsgCategoryUpdateSuccess    = "修改分类成功"
	MsgCategoryUpdateFailed     = "修改分类失败"
	MsgCategoryIDOrNameRequired = "分类ID或名称至少需要提供一个"
	MsgGetCategoryDetailSuccess = "获取分类详情成功"
)

// 用户相关常量
const (
	UserNotExist            = "用户不存在"
	UserNameAlreadyExist    = "用户名已存在"
	UserRegisterFailed      = "用户注册失败"
	UserRegisterSuccess     = "用户注册成功"
	UserLoginSuccess        = "用户登录成功"
	UserIDLack              = "缺少userId参数"
	UserIDFormatError       = "userId参数格式错误"
	UserBeenSuspended       = "用户已被停用"
	GetUserInfoFailed       = "无法获取用户信息，请重新登录"
	GetUserSuccess          = "搜索用户信息成功"
	GetUserProfileSuccess   = "获取用户个人主页成功"
	NonSelf                 = "非用户本人，不允许访问"
	IllegalStatus           = "无效用户状态"
	UpdateUserStatusFailed  = "更新用户状态失败"
	UpdateUserStatusSuccess = "更新用户状态成功"
	AvatarDeleteFailed      = "头像删除失败"
	NonUserAvatar           = "用户没有上传头像"
	OpenAvatarFailed        = "打开头像文件失败"
	UploadAvatarFailed      = "头像上传失败"
	NoChangeHappen          = "没有需要更新的信息"
	UpdateUserInfoFailed    = "更新用户信息失败"
	UpdateUserInfoSuccess   = "个人资料修改成功"
)

// 密码、邮箱和鉴权相关常量
const (
	PasswordEncryptFailed = "密码加密失败"
	PasswordUpdateFailed  = "密码更新失败"
	PasswordUpdateSuccess = "密码更新成功"
	UnauthorizedEmail     = "用户邮箱错误"
	EmailHasBeenUsed      = "邮箱已被注册"
	PasswordFalse         = "密码错误"
	TokenGenerateFailed   = "Token生成失败"
	TokenFormatError      = "Token格式错误"
	RequestWithoutToken   = "请求未携带token，无权限访问"
	TokenParseFailed      = "Token解析失败"
	UserNonLogin          = "用户未登录"
	NoPermission          = "无管理员权限，禁止访问"
)

// 验证码相关常量
const (
	VerificationCodeSendFailed = "验证码发送失败"
	VerificationCodeCheckError = "验证码校验失败"
	VerificationCodeStoreError = "验证码存储失败"
	VerificationCodeExpired    = "验证码错误或已过期"
	InvalidTransaction         = "无效的验证码业务"
	VCodeSendSuccess           = "验证码已发送至您的邮箱，请注意查收"
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
	MsgCommentPostSuccess      = "评论发表成功"
	MsgCommentDeleteSuccess    = "删除评论成功"
	MsgGetCommentListSuccess   = "获取评论列表成功"
	MsgGetAllCommentsSuccess   = "获取所有评论成功"
	MsgGetUserCommentsSuccess  = "获取用户评论成功"
	MsgGetSingleCommentSuccess = "获取单条评论成功"
)

// 聊天相关常量
const (
	GetSenderFailed           = "获取发送者信息失败"
	SessionIDLack             = "sessionId参数不能为空"
	SessionIDFormatError      = "sessionId参数格式错误"
	GetChatMessageSuccess     = "获取聊天记录成功"
	SearchKeyLack             = "searchKey参数不能为空"
	UserNotInSession          = "用户不是会话的参与者"
	ChatMsgContentEmpty       = "消息内容不能为空"
	NotSelfMsg                = "不能给自己发送消息"
	CreateNewSessionFailed    = "创建新会话失败"
	LackSessionIDOrReceiverID = "缺少会话ID或接收者ID"
	SendMsgFailed             = "消息发送失败"
	SendRealTimeMsgFailed     = "实时消息发送失败"
	SendMsgSuccess            = "消息发送成功"
	GetSessionListSuccess     = "获取会话列表成功"
)

// websocket相关常量
const (
	WSConnectFailed = "WebSocket连接失败"
)

const (
	GetNotificationSuccess = "成功获取通知"
	NotificationIDInvalid  = "通知ID无效"
	NotificationNotExist   = "通知不存在"
	MarkReadFailed         = "标记已读失败"
	MarkReadSuccess        = "标记已读成功"
)

// 帖子相关常量
const (
	CreatePostFailed         = "发帖失败"
	CreatePostSuccess        = "发帖成功"
	CreatePostDocumentFailed = "创建帖子文档关联失败"
	GetPostDetailSuccess     = "获取帖子详情成功"
	PostsObtain              = "帖子列表获取成功"
	PostNotExist             = "帖子不存在"
	GetFavoritePostFailed    = "获取收藏帖子列表失败"
)
