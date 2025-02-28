package llm

import (
	"strings"

	logger "github.com/sirupsen/logrus"
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

// 前端websocket解析时依赖这个类型
type TaskOutputType string

var CHAT_TYPE_CANDIDATE = TaskOutputType("candidate") //参考的文档、商品、网页等
var CHAT_TYPE_TOKENS = TaskOutputType("tokens")       //本次交互的tokens数
var CHAT_TYPE_MSG = TaskOutputType("msg")             //返回的文本信息
var CHAT_TYPE_SHOPPING = TaskOutputType("shop")       //购物信息

var RUN_DONE = 1
var RUN_ERR = 2

type ChatStream struct {
	Type     TaskOutputType `json:"type"` //
	Body     interface{}    `json:"body"`
	Seqno    string         `json:"seqno"` //流式输出时，用于前端区分完整的语句，相同seqno表示同一次llm输出
	Room     string         `json:"room"`
	ChatType string         `json:"chat_type"` // human\ai
}

type GenModel interface {
	/**
	messages:包含历史信息、当前prompt、用户问题
	msgListener:大模型流式返回的时候,流式返回给前端的chan
	return string:大模型全部的生成内容；int64：本次消耗的tokens
	*/
	GetChatRes(messages []llms.ChatMessage, msgListener chan string) (string, int64)
}

/*
*

maxContentLength:输入内容的最大长度，超过长度的历史会话会被截断
*/
func CombineLLMInputWithHistory(systemPrompt string, userInput string, chatHistorys []llms.ChatMessage, maxContentLength int) []llms.ChatMessage {
	systemMsg := llms.SystemChatMessage{
		Content: systemPrompt,
	}
	userMsg := llms.HumanChatMessage{
		Content: userInput,
	}

	contentLength := 0
	contentLength += len(systemMsg.GetContent())
	contentLength += len(userMsg.GetContent())

	msgs := make([]llms.ChatMessage, 0, len(chatHistorys)+2)
	msgs = append(msgs, systemMsg)

	//需要留意聊天记录的顺序
	remain := maxContentLength - len(userMsg.GetContent())
	count := 0
	for i := 0; i < len(chatHistorys); i++ {
		logger.Infof("type:%s,content:%s", chatHistorys[i].GetType(), chatHistorys[i].GetContent())
		if len(chatHistorys[i].GetContent()) == 0 || strings.TrimSpace(chatHistorys[i].GetContent()) == "" {
			continue
		}
		if chatHistorys[i].GetType() != llms.ChatMessageTypeAI && chatHistorys[i].GetType() != llms.ChatMessageTypeHuman {
			continue
		}
		if count%2 == 0 && chatHistorys[i].GetType() != llms.ChatMessageTypeHuman {
			continue
		}
		if count%2 == 1 && chatHistorys[i].GetType() != llms.ChatMessageTypeAI {
			continue
		}
		remain = remain - len(chatHistorys[i].GetContent())
		if remain > 0 {
			msgs = append(msgs, chatHistorys[i])
			count++
		} else {
			break
		}
	}
	msgs = append(msgs, userMsg)
	logger.Infof("add history num:%d", count)
	return msgs
}
