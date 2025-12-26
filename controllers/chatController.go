package controllers

import (
	"errors"
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
	"gorm.io/gorm"
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

// SearchChatMessages 搜索聊天记录接口
func SearchChatMessages(c *gin.Context) {
	// 从查询参数获取userId
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		// 用户ID缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDLack)
		return
	}

	// 从查询参数获取searchKey
	searchKey := c.Query("searchKey")
	if searchKey == "" {
		// 搜索关键词缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.SearchKeyLack)
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

	// 搜索指定用户的聊天记录
	messages, err := dao.SearchChatMessagesByUser(userID, searchKey)
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

	// 返回搜索到的聊天记录列表
	response.SuccessWithData(c, responseData, constant.GetChatMessageSuccess)
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
		"type":         "chat_message", // 消息类型标记
	}

	// 推送给接收者 (如果在线)
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
		// 仅表示实时推送失败，但消息已持久化，接收者下次上线时可通过 GetChatMessages 拉取历史消息
		response.Fail(c, http.StatusInternalServerError, nil, constant.SendRealTimeMsgFailed)
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
	var sessionList []response.SessionResponse

	for _, session := range sessions {
		// 3.1 获取双方用户信息
		user1, err1 := dao.GetUserByID(session.User1ID)
		user2, err2 := dao.GetUserByID(session.User2ID)
		if err1 != nil || err2 != nil {
			continue // 如果找不到用户，跳过该异常会话
		}

		// 3.2 获取最后一条消息
		lastMsg, errMsg := dao.GetLastMessageBySessionID(session.ID)
		lastContent := ""
		lastTime := ""
		if errMsg == nil {
			lastContent = lastMsg.Content
			lastTime = lastMsg.CreatedAt.Format("2006-01-02 15:04:05")
		} else if !errors.Is(errMsg, gorm.ErrRecordNotFound) {
			// 如果是 RecordNotFound 说明是新会话没消息，这是正常的；其他错误则要响应
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			log.Println(errMsg)
		}

		// 3.3 统计未读消息数 (我是接收者，所以统计别人发给我的未读)
		unreadCount, _ := dao.CountUnreadMessages(session.ID, currentUserID)

		// 3.4 构建响应对象
		item := response.SessionResponse{
			SessionID:   session.ID,
			UserID1:     user1.ID,
			Avatar1:     utils.GetFileURL(user1.Avatar),
			Username1:   user1.Username,
			UserID2:     user2.ID,
			Avatar2:     utils.GetFileURL(user2.Avatar),
			Username2:   user2.Username,
			LastMessage: lastContent,
			LastTime:    lastTime,
			UnreadCount: uint64(unreadCount),
		}
		sessionList = append(sessionList, item)
	}

	// 4. 返回结果
	if sessionList == nil {
		sessionList = []response.SessionResponse{}
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
