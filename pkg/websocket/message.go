package websocket

import (
	"encoding/json"
	"time"
)

// MessageType 定义消息类型
type MessageType string

const (
	OrderCreated   MessageType = "order_created"
	OrderPaid      MessageType = "order_paid"
	OrderCancelled MessageType = "order_cancelled"
	OrderRefunded  MessageType = "order_refunded"
	SystemNotice   MessageType = "system_notice"
)

// Message 定义WebSocket消息结构
type Message struct {
	Type    MessageType `json:"type"`    // 消息类型
	Content string      `json:"content"` // 消息内容
	Data    interface{} `json:"data"`    // 消息数据
	Time    string      `json:"time"`    // 消息时间
}

// NewOrderMessage 创建订单相关消息
func NewOrderMessage(msgType MessageType, content string, data interface{}) ([]byte, error) {
	msg := Message{
		Type:    msgType,
		Content: content,
		Data:    data,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	}

	return json.Marshal(msg)
}

// NewSystemNotice 创建系统通知消息
func NewSystemNotice(content string) ([]byte, error) {
	return NewOrderMessage(SystemNotice, content, nil)
}
