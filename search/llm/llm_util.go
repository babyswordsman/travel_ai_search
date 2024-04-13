package llm

import (
	"context"
	"fmt"
	"sync"
	"time"
	"travel_ai_search/search/common"
	"travel_ai_search/search/user"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

var ROLE_SYSTEM = "system"
var ROLE_USER = "user"
var ROLE_ASSISTANT = "assistant"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
	GetChatRes(messages []schema.ChatMessage, msgListener chan string) (string, int64)
}

type UserChatHistory struct {
	Room     string
	Lasttime time.Time
	ChatBuff *memory.ConversationWindowBuffer
}

var userChatHistorys = &sync.Map{}

func AddChatHistory(userId string, room string, query, response string) error {
	v, ok := userChatHistorys.LoadOrStore(fmt.Sprintf("%s-%s", room, userId), &UserChatHistory{})
	history := v.(*UserChatHistory)
	if !ok {
		history.ChatBuff = memory.NewConversationWindowBuffer(5,
			memory.WithReturnMessages(true),
			memory.WithInputKey("query"),
			memory.WithOutputKey("response"))
	}
	history.Lasttime = time.Now()
	err := history.ChatBuff.SaveContext(context.Background(),
		map[string]any{history.ChatBuff.InputKey: query},
		map[string]any{history.ChatBuff.OutputKey: response})
	if err != nil {
		logger.Errorf("add %s chat history err:%s", userId, err)
	}
	return err
}

/*
*
加载用户对话记录
*/
func LoadChatHistory(userId string, room string) []schema.ChatMessage {
	if userId == user.EmpytUser.UserId {
		return make([]schema.ChatMessage, 0)
	}
	v, ok := userChatHistorys.Load(fmt.Sprintf("%s-%s", room, userId))
	if !ok {
		return make([]schema.ChatMessage, 0)
	}
	history := v.(*UserChatHistory)
	tmp := make(map[string]any)
	historyMsgs, err := history.ChatBuff.LoadMemoryVariables(context.Background(), tmp)
	if err != nil {
		return make([]schema.ChatMessage, 0)
	}
	v, ok = historyMsgs[history.ChatBuff.MemoryKey]
	if !ok {
		return make([]schema.ChatMessage, 0)
	}

	return v.([]schema.ChatMessage)

}

func InitUserChatHistory() {

	go func() {
		tick := time.NewTicker(time.Minute * 5)
		for {
			select {
			case v := <-tick.C:
				{
					logger.Infof("start clean chat history,%s", v.Format("2006-01-02 15:04:05"))
					go func() {
						defer func() {
							if err := recover(); err != nil {
								logger.Errorf("panic err is %s \r\n %s", err, common.GetStack())
							}
							curNow := time.Now()
							var delNum int = 0
							var remainNum int = 0

							userChatHistorys.Range(func(k any, v any) bool {
								last := v.(*UserChatHistory).Lasttime
								if curNow.Sub(last) > time.Minute*10 {
									logger.Infof("del %s,lasttime:%s", k.(string), last.Format("2016-01-02 15:04:05"))
									userChatHistorys.Delete(k)
									delNum++
								} else {
									remainNum++
								}
								return true
							})

							logger.Infof("del %d,remain %d", delNum, remainNum)
						}()
					}()
				}

			}
		}

	}()
}
