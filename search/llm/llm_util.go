package llm

import (
	"github.com/tmc/langchaingo/llms"
)

var ROLE_SYSTEM = "system"
var ROLE_USER = "user"
var ROLE_ASSISTANT = "assistant"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	//单位毫秒:time.Now().UnixMilli()
	Timestamp int64 `json:"timestamp"`
}

// GetType gets the type of the message.
func (msg *Message) GetType() llms.ChatMessageType {
	switch msg.Role {
	case ROLE_SYSTEM:
		return llms.ChatMessageTypeSystem
	case ROLE_ASSISTANT:
		return llms.ChatMessageTypeAI
	case ROLE_USER:
		return llms.ChatMessageTypeHuman
	default:
		return llms.ChatMessageTypeTool
	}
}

// GetContent gets the content of the message.
func (msg *Message) GetContent() string {
	return msg.Content
}

func (msg *Message) GetTimestamp() int64 {
	return msg.Timestamp
}

var CHAT_TYPE_CANDIDATE = "candidate"
var CHAT_TYPE_TOKENS = "tokens"
var CHAT_TYPE_MSG = "msg"

type ChatStream struct {
	Type     string      `json:"type"`
	Body     interface{} `json:"body"`
	Seqno    string      `json:"seqno"`
	Room     string      `json:"room"`
	ChatType string      `json:"chat_type"`
}

type GenModel interface {
	/**
	messages:包含历史信息、当前prompt、用户问题
	msgListener:大模型流式返回的时候,流式返回给前端的chan
	return string:大模型全部的生成内容；int64：本次消耗的tokens
	*/
	GetChatRes(messages []llms.ChatMessage, msgListener chan string) (string, int64)
}
