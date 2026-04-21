package utils

import (
	"context"
	"sync"
)

// aiSessionStreamTaskMap AI会话流式任务管理器，用于存储可跨请求取消的任务
// key: sessionId (会话标识)
// value: context.CancelFunc (上下文取消函数)
var (
	aiSessionStreamTaskMap sync.Map
)

// RegisterAISessionStreamTask 注册AI会话流式任务
// 将sessionId和对应的取消函数存入任务管理器
// sessionId: AI会话的唯一标识
// cancel: 上下文取消函数，用于终止该会话的流式输出
func RegisterAISessionStreamTask(sessionId string, cancel context.CancelFunc) {
	aiSessionStreamTaskMap.Store(sessionId, cancel)
}

// UnregisterAISessionStreamTask 注销AI会话流式任务
// 从任务管理器中移除指定sessionId的任务
// sessionId: 需要注销的AI会话标识
func UnregisterAISessionStreamTask(sessionId string) {
	aiSessionStreamTaskMap.Delete(sessionId)
}

// CancelAISessionStreamTask 取消AI会话流式任务
// 查找指定sessionId的任务并执行取消操作
// sessionId: 需要取消的AI会话标识
// 返回值: true表示成功找到并取消任务，false表示任务不存在或已结束
func CancelAISessionStreamTask(sessionId string) bool {
	if cancelVal, ok := aiSessionStreamTaskMap.Load(sessionId); ok {
		cancel := cancelVal.(context.CancelFunc)
		cancel()
		aiSessionStreamTaskMap.Delete(sessionId)
		return true
	}
	return false
}
