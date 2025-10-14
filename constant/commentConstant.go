package constant

const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusNotFound            = 404
	StatusInternalServerError = 500
)

const (
	CodeSuccess = 0
	CodeError   = 1
)

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

const (
	MsgCommentPostSuccess     = "评论发表成功"
	MsgCommentDeleteSuccess   = "删除评论成功"
	MsgGetCommentListSuccess  = "获取评论列表成功"
	MsgGetAllCommentsSuccess  = "获取所有评论成功"
	MsgGetUserCommentsSuccess = "获取用户评论成功"
)
