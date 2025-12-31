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
	// 1. 获取读锁
	// 为了防止在读取 Clients map 时，有其他协程（如用户注册/注销）同时修改 map 导致崩溃。
	// RLock 允许并发读，但阻塞写。
	manager.Lock.RLock()

	// 2. 从 map 中查找用户
	// 尝试根据 userID 获取对应的 client 连接对象。
	// client 是连接对象，ok 是布尔值，表示该用户是否存在于 map 中（是否在线）。
	client, ok := manager.Clients[userID]

	// 3. 释放读锁
	// 查找操作已完成，必须尽快释放锁，以免阻塞其他协程的写操作（如用户上线）。
	manager.Lock.RUnlock()

	// 4. 判断用户是否在线
	if ok {
		// 5. 序列化消息
		// 将传入的 message 对象（struct 或 map）转换为 JSON 格式的字节切片 ([]byte)。
		// 网络传输通常使用 []byte。
		jsonMessage, err := json.Marshal(message)

		// 6. 错误处理
		// 如果序列化失败（例如结构体里有无法序列化的字段），记录日志并返回错误。
		if err != nil {
			log.Println("消息序列化失败:", err)
			return err
		}

		// 7. 非阻塞发送逻辑 (Select 语句)
		// select 用于处理通道 (channel) 操作。
		select {

		// 8. 尝试发送消息
		// 试图将 jsonMessage 发送到 client.Send 通道中。
		// 这里的 client.Send 通常是一个带缓冲的 channel。
		case client.Send <- jsonMessage:
			// 如果通道未满，消息成功写入，发送流程结束。

		// 9. 通道阻塞处理 (Default 分支)
		// 如果 client.Send 通道已满（说明客户端网络卡顿或处理太慢，导致积压），
		// select 会立即走 default 分支，而不会一直阻塞在这里等待。
		default:
			// 10. 关闭通道
			// 既然发不进去，说明连接可能死掉了或者客户端异常。
			// 关闭 Send 通道，通知写协程退出循环，断开 WebSocket 连接。
			close(client.Send)

			// 11. 从管理器中移除用户
			// 【⚠️注意】：这里存在并发安全隐患！
			// 前面已经在第3步释放了锁，这里直接修改 manager.Clients map 可能会导致 panic (concurrent map writes)。
			// 正确做法通常是需要重新加写锁 (Lock/Unlock)，或者通过注册/注销通道由管理协程统一处理移除。
			delete(manager.Clients, userID)
		}
	}

	// 12. 返回 nil
	// 无论发送成功，还是用户不在线（ok 为 false），方法都返回 nil。
	// 这意味着如果用户不在线，消息会被直接丢弃（“发后即焚”），符合即时通讯中“未读消息走离线接口，不走 WebSocket”的设计逻辑。
	return nil
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
