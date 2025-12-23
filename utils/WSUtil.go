package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client 代表一个连接的客户端
type Client struct {
	ID     uint64          // 用户ID
	Socket *websocket.Conn // WebSocket连接
	Send   chan []byte     // 待发送的消息通道
}

// Manager WebSocket连接管理器
type Manager struct {
	Clients    map[uint64]*Client // 在线用户映射表: UserID -> Client
	Register   chan *Client       // 注册连接通道
	Unregister chan *Client       // 注销连接通道
	Lock       sync.RWMutex       // 读写锁，保护Clients map
}

// WSManager 全局WebSocket管理器实例
var WSManager = Manager{
	Clients:    make(map[uint64]*Client),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}

// WSMessage 推送给前端的消息结构
type WSMessage struct {
	Type       string      `json:"type"`       // 消息类型: "chat_message", "notification" 等
	ReceiverID uint64      `json:"receiverId"` // 接收者ID
	Data       interface{} `json:"data"`       // 消息负载
}

// Start 启动管理器的主循环 (需要在main.go中调用)
// 函数前置的(manager *Manager)规定了该函数属于Manager类，只能由Manager类的一个实例manager调用
func (manager *Manager) Start() {
	for {
		select {
		case client := <-manager.Register:
			// 有新用户连接
			manager.Lock.Lock()
			// 如果该用户已有连接，先关闭旧连接 (简单的单点登录逻辑)
			if oldClient, ok := manager.Clients[client.ID]; ok {
				close(oldClient.Send)
				delete(manager.Clients, client.ID)
			}
			manager.Clients[client.ID] = client
			manager.Lock.Unlock()
			log.Printf("用户 %d 已连接 WebSocket", client.ID)

		case client := <-manager.Unregister:
			// 用户断开最后一条连接
			manager.Lock.Lock()
			// 只有当 Map 中存的 Client 指针 等于 当前要注销的 Client 指针时，才执行删除和关闭
			//（也就是说如果本次注销是旧连接自己的注销逻辑而不是新连接的强制注销，那么就不会执行，因为Map 中存的 Client 指针）
			// 这防止了：
			// 1. 旧连接注销时把新连接删了
			// 2. 旧连接被 Register 关闭后，ReadPump协程断开触发的 Unregister 再次关闭导致 Panic
			if targetClient, ok := manager.Clients[client.ID]; ok && targetClient == client {
				delete(manager.Clients, client.ID)
				close(client.Send)
			}
			manager.Lock.Unlock()
			log.Printf("用户 %d 断开 WebSocket", client.ID)
		}
	}
}

// SendToUser 向指定用户推送消息
func (manager *Manager) SendToUser(userID uint64, message interface{}) (err error) {
	manager.Lock.RLock()
	client, ok := manager.Clients[userID]
	manager.Lock.RUnlock()

	if ok {
		// 序列化消息
		jsonMessage, err := json.Marshal(message)
		if err != nil {
			log.Println("消息序列化失败:", err)
			return err
		}
		// 将消息放入客户端的发送通道，非阻塞
		select {
		case client.Send <- jsonMessage:
		default:
			// 发送通道已满或阻塞，断开连接
			close(client.Send)
			delete(manager.Clients, userID)
		}
	}
	return nil

	// 否则用户不在线，无需实时发送信息，用户上线后会自动调用相应接口收到消息
}

// WritePump 监听 Send 通道，将消息写入 WebSocket
func (c *Client) WritePump() {
	defer func() {
		err := c.Socket.Close()
		if err != nil {
			return
		}
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// 通道已关闭
				err := c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					return
				}
				return
			}
			// 发送文本消息
			err := c.Socket.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		}
	}
}

// ReadPump 监听 WebSocket 读取 (用于检测断开和处理心跳)
func (c *Client) ReadPump() {
	defer func() {
		WSManager.Unregister <- c
		err := c.Socket.Close()
		if err != nil {
			return
		}
	}()
	for {
		// 这里我们主要只做单向推送，读取循环用于保持连接活跃和检测断开
		_, _, err := c.Socket.ReadMessage()
		if err != nil {
			break
		}
	}
}

// Upgrader HTTP升级为WebSocket的配置
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
