package types

import "time"

type RuntimeEventCode string

const (
	RuntimeEventCodeNotification = "RuntimeForNotification"
	RuntimeEventCodeCrawl        = "RuntimeForCrawl"
	RuntimeEventCodeClean        = "RuntimeForClean"
)

type MessageOptional struct {
	IsNotify bool   `json:"isNotify"` //
	IsStore  bool   `json:"isStore"`  // 是否保存数据
	ShowType string `json:"showType"` // 展示类型 message | modal | notification
}

type EventData struct {
	CheckHash  string          `json:"checkHash"`  // 消息唯一hash
	Level      string          `json:"level"`      // 级别
	Title      string          `json:"title"`      // 消息标题
	Message    string          `json:"message"`    // 消息主体
	MetaData   interface{}     `json:"metaData"`   // 元数据
	Optional   MessageOptional `json:"optional"`   // 消息选项
	IsRead     bool            `json:"isRead"`     // 是否已读
	ExpiredAt  time.Time       `json:"expiredAt"`  // 过期时间
	ReceivedAt time.Time       `json:"receivedAt"` // 接收时间
}

type RuntimeData struct {
	EventCode RuntimeEventCode
	EventData EventData
	EventTime time.Time
}

func NewRuntimeData(code RuntimeEventCode, data EventData) RuntimeData {
	if data.ReceivedAt.IsZero() {
		data.ReceivedAt = time.Now()
	}
	return RuntimeData{
		EventCode: code,
		EventData: data,
		EventTime: time.Now(),
	}
}
