package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

// GetChatMessages 获取聊天记录接口
func GetChatMessages(c *gin.Context) {
	// 从查询参数获取sessionId
	sessionIdStr := c.Query("sessionId")
	if sessionIdStr == "" {
		// 会话ID缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.SessionIDLack)
		return
	}

	// 从查询参数获取userId
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		// 用户ID缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDLack)
		return
	}

	// 将字符串转换为uint64类型的会话ID
	sessionID, err := strconv.ParseUint(sessionIdStr, 10, 64)
	if err != nil {
		// 会话ID格式错误，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.SessionIDFormatError)
		return
	}

	// 将字符串转换为uint64类型的用户ID
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		// 用户ID格式错误，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgUserIDFormatError)
		return
	}

	// 从上下文中获取用户声明信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		// 获取用户信息失败，返回错误响应
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}

	// 类型转换用户声明信息
	userClaims := claims.(*utils.MyClaims)

	// 验证请求的用户ID是否与当前登录用户ID一致
	if userClaims.UserID != userID {
		// 不是本人操作，返回错误响应
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 验证用户是否是会话的参与者
	session, err := dao.GetSessionByID(sessionID)
	if err != nil {
		// 如果会话不存在，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 验证当前用户是会话的参与者（User1ID或User2ID）
	if session.User1ID != userID && session.User2ID != userID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.UserNotInSession)
		return
	}

	// 查询指定会话中的所有聊天记录
	messages, err := dao.GetAllMessagesBySession(sessionID)
	if err != nil {
		// 数据库操作错误，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构造返回数据数组，存储聊天记录信息
	responseData, err := response.BuildChatRecordResponses(messages)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 返回用户的聊天记录列表
	response.SuccessWithData(c, responseData, constant.GetChatMessageSuccess)
}

// SearchChatMessages 全局搜索聊天记录（返回包含搜索关键词的聊天信息所在的会话列表）
// GET /api/chat/globalSearch
// 处理逻辑：先查目标聊天信息messages，再对其中每一条message循环处理：
// 1. 根据message.sessionId查找对应的会话session
// 2. 建立message.sessionId到接口响应结构体SearchChatResponse的键值对映射，
// 3. 若映射已存在（前面已经处理过相同会话的信息），则只需更新所属会话的响应结构体中的消息数量字段即可
// 4. 每次循环都append到[]*response.SearchChatResponse数组里
// PS：这个接口的数据查询业务逻辑比较复杂，也比较特殊，没有复用性需求，故无需额外定义build response函数
func SearchChatMessages(c *gin.Context) {
	// 1. 获取参数
	keyword := c.Query("searchKey")
	if keyword == "" {
		response.Fail(c, http.StatusBadRequest, nil, constant.SearchKeyLack)
		return
	}

	// 2. 获取当前用户
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	currentUserID := userClaims.UserID
	currentUser, err := dao.GetUserByID(userClaims.UserID)

	// 3. 查出所有符合条件的消息 (已按时间倒序)
	messages, err := dao.SearchMessagesByUserAndKeyword(currentUserID, keyword)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 4. 在内存中进行分组和聚合
	// key: SessionID, value: *response.SearchChatResult
	resultMap := make(map[uint64]*response.SearchChatResult)
	var resultList []*response.SearchChatResult
	// 用切片来保持顺序（需要按最近匹配的会话排序）
	// 这里使用指针切片 []*response.SearchChatResult 而不是结构体切片 []response.SearchChatResult，主要有以下几个原因：
	// - 方便在 Map 中修改值：
	// 在 Go 语言中，从 map 获取的值是不可寻址的（unaddressable）。也就是说，如果定义 map[uint64]SearchChatResult（非指针），当你写 entry := resultMap[id] 时，你得到的是该结构体的一个副本。
	// 如果随后修改 entry.MatchCount++，你修改的只是这个副本的计数，原 map 中的结构体不会发生任何变化。
	// 而使用指针 *SearchChatResult，map 中存储的是指向内存中同一个对象的地址。当你 entry := resultMap[id] 时，你拿到的是同一个指针副本，但它指向的还是原来的对象。
	// 此时修改 entry.MatchCount++ 能够真正影响到内存中的对象，也就能同步更新到 resultList 引用的对象中。
	//
	// - 避免大数据拷贝，提高性能：
	// 如果 SearchChatResult 结构体比较大（虽然这里不算太大，但包含字符串等字段），每次从 map 取值、或者 append 到切片时，如果不使用指针，都会发生结构体的值拷贝。
	// 使用指针只拷贝 8 字节（64位系统）的内存地址，效率更高。
	//
	// - 保持 Map 和 List 的数据一致性：
	// 在这个逻辑中，我们同时维护了一个 resultMap（用于快速查找去重）和一个 resultList（用于保持顺序）。
	// 我们希望修改 Map 中的元素（比如增加计数）时，List 中的对应元素也能自动体现这个修改。
	// 通过让它们都指向同一个内存地址（指针），我们就不用在修改 Map 后再去遍历 List 查找更新了。
	for _, msg := range messages {
		if entry, exists := resultMap[msg.SessionID]; exists {
			// 如果该会话已存在结果中
			entry.MatchCount++
			// 因为步骤3中的 DAO 查出来是按时间倒序的，所以第一次存入的就是最新的，后续遍历到的都是旧的，不用更新 LatestMsg
		} else {
			// 如果该会话第一次出现
			// 4.1 获取会话信息以确定“对方”是谁
			session, err := dao.GetSessionByID(msg.SessionID)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
				return
			}

			// 确定对方ID
			var targetUserID uint64
			if session.User1ID == currentUserID {
				targetUserID = session.User2ID
			} else {
				targetUserID = session.User1ID
			}

			// 4.2 获取对方用户信息
			targetUser, err := dao.GetUserByID(targetUserID)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
				return
			}

			// 4.3 构造结果对象
			newEntry := &response.SearchChatResult{
				SessionID:   msg.SessionID,
				UserID1:     targetUserID,
				UserName1:   targetUser.Username,
				UserAvatar1: utils.GetFileURL(targetUser.Avatar),
				UserID2:     currentUser.ID,
				UserName2:   currentUser.Username,
				UserAvatar2: utils.GetFileURL(currentUser.Avatar),
				MatchCount:  1,
				LatestMsg:   msg.Content,
			}

			// 将该会话插入map表
			resultMap[msg.SessionID] = newEntry
			resultList = append(resultList, newEntry)
		}
	}

	// 5. 返回结果
	// 如果没有搜索结果，返回空数组而不是 null
	if resultList == nil {
		resultList = []*response.SearchChatResult{}
	}

	response.SuccessWithData(c, resultList, constant.SearchChatSuccess)
}

// CreateChatSession 创建聊天会话接口
// POST /api/createChat
func CreateChatSession(c *gin.Context) {
	var req dto.CreateChatSessionDTO

	// 1. 绑定参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 验证参数（req.SenderID必须和用户本人一致）
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	if req.SenderID != nil && *req.SenderID != userClaims.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 3. 创建会话
	session := models.Session{
		User1ID: *req.SenderID,
		User2ID: *req.ReceiverID,
	}
	if err := dao.CreateSession(&session); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.CreateNewSessionFailed)
		return
	}

	// 4. 构建返回数据（CreateChatSession和GetSessionList复用同一个封装的响应构建函数）
	// 注意由于会话相关接口要返回未读消息数，而查询未读消息需要明确当前用户本人ID，在BuildSessionResponse中光靠session不够
	// 因此这里再传入从JWT解析的用户本人ID，用于统计未读消息数
	responseData, err := response.BuildSessionResponse(session, userClaims.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	response.SuccessWithData(c, responseData, constant.CreateNewSessionSuccess)
}

// SendMessage 发送信息接口
// POST /api/chat/message
func SendMessage(c *gin.Context) {
	var req dto.SendMessageDTO

	// 1. 绑定参数 (multipart/form-data)
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 验证内容不能为空
	if req.Content == "" {
		response.Fail(c, http.StatusBadRequest, nil, constant.ChatMsgContentEmpty)
		return
	}

	// 3. 获取当前登录用户ID
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	currentUserID := userClaims.UserID

	var targetSessionID uint64

	// 4. 判断 SessionID
	if req.SessionID != nil && *req.SessionID != 0 {
		// 4.1 如果传了 SessionID，直接验证并使用
		targetSessionID = *req.SessionID
		session, err := dao.GetSessionByID(targetSessionID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
		// 验证当前用户是否在会话中
		if session.User1ID != currentUserID && session.User2ID != currentUserID {
			response.Fail(c, http.StatusForbidden, nil, constant.UserNotInSession)
			return
		}
	} else if req.ReceiverID != nil && *req.ReceiverID != 0 {
		// 4.2 如果没传 SessionID 但传了 ReceiverID
		receiverID := *req.ReceiverID

		// 不能给自己发消息
		if receiverID == currentUserID {
			response.Fail(c, http.StatusBadRequest, nil, constant.NotSelfMsg)
			return
		}

		// 检查是否已有会话
		session, err := dao.GetSessionByUsers(currentUserID, receiverID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}

		if session != nil {
			// 已有会话，使用现有ID
			targetSessionID = session.ID
		} else {
			// 无会话，创建新会话
			newSession := models.Session{
				User1ID: currentUserID,
				User2ID: receiverID,
			}
			if err := dao.CreateSession(&newSession); err != nil {
				response.Fail(c, http.StatusInternalServerError, nil, constant.CreateNewSessionFailed)
				return
			}
			targetSessionID = newSession.ID
		}
	} else {
		// 两个ID都没传
		response.Fail(c, http.StatusBadRequest, nil, constant.LackSessionIDOrReceiverID)
		return
	}

	// 5. 创建消息
	message := models.Message{
		SessionID: targetSessionID,
		SenderID:  currentUserID,
		Content:   req.Content,
		Status:    "unread", // 默认为未读
	}

	if err := dao.CreateMessage(&message); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.SendMsgFailed)
		return
	}

	// 6. 更新会话的最后活动时间
	// 这样该会话在列表中就会排到最前面
	if err := dao.UpdateSessionTime(targetSessionID); err != nil {
		// 记录日志即可，更新时间失败不应影响消息发送的成功响应
		// log.Printf("更新会话时间失败: %v", err)
	}

	// 7. WebSocket 实时推送（若接收者不在线，存入数据库就算调用成功；若在线而通过websocket推送失败，则调用失败）
	// 获取发送者信息，用于前端展示
	sender, _ := dao.GetUserByID(currentUserID)

	wsData := gin.H{
		"sessionId":    targetSessionID,
		"messageId":    strconv.FormatUint(message.ID, 10),
		"senderId":     currentUserID,
		"senderName":   sender.Username,
		"senderAvatar": utils.GetFileURL(sender.Avatar),
		"content":      message.Content,
		"sendTime":     message.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 7.1 推送给接收者 (如果在线) 以及本人，实现客户端实时接收
	// 注意：这里 receiverID 需要根据逻辑获取 (如果聊天信息是通过SessionID发送的，需要查Session找对方ID；否则直接以ReceiverID为准)
	var realReceiverID uint64
	if req.ReceiverID != nil && *req.ReceiverID != 0 {
		realReceiverID = *req.ReceiverID
	} else {
		// 根据SessionID反查接收者
		session, _ := dao.GetSessionByID(targetSessionID)
		if session.User1ID == currentUserID {
			realReceiverID = session.User2ID
		} else {
			realReceiverID = session.User1ID
		}
	}
	// 调用 WebSocket 管理器发送
	err := utils.WSManager.SendToUser(realReceiverID, utils.WSMessage{
		Type:       "chat_message",
		ReceiverID: realReceiverID,
		Data:       wsData,
	})
	if err != nil {
		// 实时推送失败，但消息已持久化，接收者下次上线时可通过 GetChatMessages 拉取历史消息
		// 因此仅打印日志不返回错误
		log.Printf("WS推送给接收者 %d 失败(可能离线): %v", realReceiverID, err)
	}

	// 7.2 推送给发送者
	// 可实现“多端同步”：如果一台电脑发了消息，另一台电脑的聊天窗口也能实时收到这条发出的消息并上屏。
	err = utils.WSManager.SendToUser(currentUserID, utils.WSMessage{
		Type:       "chat_message",
		ReceiverID: currentUserID, // 目标是发送者自己
		Data:       wsData,
	})
	if err != nil {
		// 实时推送失败，但消息已持久化，接收者下次上线时可通过 GetChatMessages 拉取历史消息
		// 因此仅打印日志不返回错误
		log.Printf("WS推送给本人 %d 失败: %v", currentUserID, err)
	}

	// 8. 返回成功响应
	responseData := gin.H{
		"messageId": strconv.FormatUint(message.ID, 10), // 转为string
		"sendTime":  message.CreatedAt.Format("2006-01-02 15:04:05"),
		"status":    message.Status,
	}

	response.Success(c, responseData, constant.SendMsgSuccess)
}

// GetSessionList 获取当前用户的会话列表
// GET /api/chat/sessions
func GetSessionList(c *gin.Context) {
	// 0. 从查询参数获取userId
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		// 用户ID缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDLack)
		return
	}
	// 将字符串转换为uint64类型的用户ID
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		// 用户ID格式错误，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgUserIDFormatError)
		return
	}

	// 1.1 获取当前登录用户ID
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	currentUserID := userClaims.UserID

	// 1.2 验证请求的用户ID是否与当前登录用户ID一致
	if userClaims.UserID != userID {
		// 不是本人操作，返回错误响应
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 2. 查询用户的所有会话
	sessions, err := dao.GetUserSessions(currentUserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		log.Println(err)
		return
	}

	// 3. 组装响应数据
	sessionList, err := response.BuildSessionResponses(sessions, currentUserID)
	if err != nil {
		// 响应构建函数内部已经返回了错误常量，这里用err响应即可
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
	}

	response.SuccessWithData(c, sessionList, constant.GetSessionListSuccess)
}

// ConnectWS : WebSocket 连接接口
// GET /api/ws
func ConnectWS(c *gin.Context) {
	// 1. 获取前端发起连接的用户ID (Token 放在 Query 参数里: /api/ws?token=xxx)
	// WebSocket 连接建立时通常不能自定义 Header，所以 Token 常放在 URL 参数中
	token := c.Query("token")
	if token == "" {
		// 原生的WebSocket API不会暴露HTTP响应的Body内容给JS代码，但最好还是返回相应的响应码以及错误信息，方便开发者模式检查。
		response.Fail(c, http.StatusUnauthorized, nil, constant.RequestWithoutToken)
		return
	}

	claims, err := utils.ParseToken(token)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, nil, constant.TokenParseFailed)
		return
	}
	userID := claims.UserID

	// 2. 升级 HTTP 连接为 WebSocket
	conn, err := utils.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.WSConnectFailed)
		return
	}

	// 3. 注册客户端
	client := &utils.Client{
		ID:     userID,
		Socket: conn,
		Send:   make(chan []byte, 256),
	}
	utils.WSManager.Register <- client

	// 4. 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}
