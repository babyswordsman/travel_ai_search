package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"travel_ai_search/search/common"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/user"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
)

//暂时不考虑缓存

type UserChatHistory struct {
	Room     string
	Lasttime time.Time
	ChatBuff *memory.ConversationWindowBuffer
}

var historyStore ChatHistoryStore

func InitKVHistoryStoreInstance(client *kvclient.KVClient, maxSize int) {
	historyStore = &KVChatHistoryStore{client: client, maxSize: int64(maxSize)}
}

func InitMemHistoryStoreInstance(maxSize int) {
	historyStore = &MemoryChatHistoryStore{maxSize: int64(maxSize)}
}

func GetHistoryStoreInstance() ChatHistoryStore {
	return historyStore
}

type ChatHistoryStore interface {
	AddChatHistory(userId string, room string, query, response string) error
	LoadChatHistoryForLLM(userId string, room string) []llms.ChatMessage
	LoadChatHistoryForHuman(userId string, room string) []llms.ChatMessage
	StarCleanTask()
}

type MemoryChatHistoryStore struct {
	maxSize int64
}

var userChatHistorys = &sync.Map{}

func (store *MemoryChatHistoryStore) AddChatHistory(userId string, room string, query, response string) error {
	v, ok := userChatHistorys.LoadOrStore(GetKey(userId, room), &UserChatHistory{})
	history := v.(*UserChatHistory)
	if !ok {
		history.ChatBuff = memory.NewConversationWindowBuffer(int(store.maxSize),
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
func (store *MemoryChatHistoryStore) LoadChatHistoryForHuman(userId string, room string) []llms.ChatMessage {
	if userId == user.EmpytUser.UserId {
		return make([]llms.ChatMessage, 0)
	}

	v, ok := userChatHistorys.Load(GetKey(userId, room))
	if !ok {
		return make([]llms.ChatMessage, 0)
	}
	history := v.(*UserChatHistory)
	tmp := make(map[string]any)
	historyMsgs, err := history.ChatBuff.LoadMemoryVariables(context.Background(), tmp)
	if err != nil {
		return make([]llms.ChatMessage, 0)
	}
	v, ok = historyMsgs[history.ChatBuff.MemoryKey]
	if !ok {
		return make([]llms.ChatMessage, 0)
	}

	return v.([]llms.ChatMessage)

}

func (store *MemoryChatHistoryStore) LoadChatHistoryForLLM(userId string, room string) []llms.ChatMessage {
	//todo:LlM maybe reverse
	return store.LoadChatHistoryForHuman(userId, room)
}

func (store *MemoryChatHistoryStore) StarCleanTask() {

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

type KVChatHistoryStore struct {
	client  *kvclient.KVClient
	maxSize int64
}

func GetKey(userId, room string) string {
	return fmt.Sprintf("%s-%s", room, userId)
}

func (store *KVChatHistoryStore) AddChatHistory(userId string, room string, query, response string) error {
	messages := make([]*Message, 0, 2)
	timestamp := time.Now().UnixMilli()
	messages = append(messages,
		&Message{Role: ROLE_USER, Content: query, Timestamp: timestamp})
	messages = append(messages,
		&Message{Role: ROLE_ASSISTANT, Content: response, Timestamp: timestamp})

	buf, err := json.Marshal(messages)
	if err != nil {
		logger.Errorf("marshal err:%s", err)
	}
	key := GetKey(userId, room)
	err = store.client.LPush(key, buf)
	if err != nil {
		logger.Errorf("lpush %s err:%s", key, err)
	}
	status, err1 := store.client.LTrim(key, 0, store.maxSize-1)
	if err1 != nil {
		logger.Errorf("ltrim %s err:%s", key, err)
	} else {
		logger.Infof("ltrim %s status:%s", key, status)
	}

	return err
}
func (store *KVChatHistoryStore) LoadChatHistoryForLLM(userId string, room string) []llms.ChatMessage {
	key := GetKey(userId, room)
	values, err := store.client.LRange(key, 0, store.maxSize-1)
	msgs := make([]llms.ChatMessage, 0, len(values)*2+1)
	if err != nil {
		return msgs
	}
	for _, str := range values {
		messages := make([]*Message, 0, 2)
		err := json.Unmarshal([]byte(str), &messages)
		if err != nil {
			logger.Errorf("unmarshal %s err:%s", str, err)
		}
		for _, m := range messages {
			msgs = append(msgs, m)
		}

	}
	return msgs
}

func (store *KVChatHistoryStore) LoadChatHistoryForHuman(userId string, room string) []llms.ChatMessage {
	key := GetKey(userId, room)
	values, err := store.client.LRange(key, 0, store.maxSize-1)
	msgs := make([]llms.ChatMessage, 0, len(values)*2+1)
	if err != nil {
		return msgs
	}
	//列表是时间倒序的
	for i := len(values) - 1; i >= 0; i-- {
		str := values[i]
		messages := make([]*Message, 0, 2)
		err := json.Unmarshal([]byte(str), &messages)
		if err != nil {
			logger.Errorf("unmarshal %s err:%s", str, err)
		}
		for _, m := range messages {
			msgs = append(msgs, m)
		}

	}
	return msgs
}
func (store *KVChatHistoryStore) StarCleanTask() {
	//empty
}
