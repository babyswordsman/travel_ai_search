package llm

import (
	"encoding/json"
	"fmt"
	"testing"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/kvclient"

	"github.com/tmc/langchaingo/llms"
)

func TestKVChatHistoryStore(t *testing.T) {
	config := &conf.Config{
		RedisAddr: "127.0.0.1:9221",
	}
	_, err := kvclient.InitClient(config)
	if err != nil {
		t.Error("init client err", err)
	}

	InitKVHistoryStoreInstance(kvclient.GetInstance(), 10)
	userId := "id-1"
	room := "chat"
	for i := 0; i < 20; i++ {
		GetHistoryStoreInstance().AddChatHistory(userId, room, fmt.Sprintf("问题%d", i), fmt.Sprintf("答案%d", i))
	}

	messages := GetHistoryStoreInstance().LoadChatHistoryForLLM(userId, room)

	t.Log("msg length:", len(messages))
	if len(messages) != 10*2 {
		t.Errorf("actual:%d,expected:%d", len(messages), 10*2)
	}

	msg0 := messages[0]
	msg1 := messages[1]

	t.Logf("actual:%s,%s,expected:%s,%s", msg0.GetContent(), msg0.GetType(), "答案19", "ai")
	t.Logf("actual:%s,%s,expected:%s,%s", msg1.GetContent(), msg1.GetType(), "问题19", "human")

	if msg1.GetContent() != "答案19" && msg1.GetType() != llms.ChatMessageTypeAI {
		t.Errorf("actual:%s,%s,expected:%s,%s", msg1.GetContent(), msg1.GetType(), "答案19", "ai")
	}

	if msg0.GetContent() != "问题19" && msg0.GetType() != llms.ChatMessageTypeHuman {
		t.Errorf("actual:%s,%s,expected:%s,%s", msg0.GetContent(), msg0.GetType(), "问题19", "human")
	}

}

func TestMemoryChatHistoryBuff(t *testing.T) {
	store := &MemoryChatHistoryStore{maxSize: 5}
	userId := "aaaaa"
	room := "chat"
	store.AddChatHistory(userId, room, "我是谁", "你是张三1")
	store.AddChatHistory(userId, room, "我是谁", "你是李四2")

	msgs := store.LoadChatHistoryForLLM(userId, room)

	if len(msgs) != 4 {
		t.Fatalf("expect:%d,actual:%d", 4, len(msgs))
	}
	store.AddChatHistory(userId, room, "我是谁", "你是张三3")
	store.AddChatHistory(userId, room, "我是谁", "你是李四4")
	store.AddChatHistory(userId, room, "我是谁", "你是张三5")
	store.AddChatHistory(userId, room, "我是谁", "你是李四6")
	msgs = store.LoadChatHistoryForLLM(userId, room)
	if len(msgs) != 10 {
		t.Fatalf("expect:%d,actual:%d", 10, len(msgs))
	}

	req := map[string]any{"text": msgs}
	bytes, _ := json.Marshal(req)

	t.Error(string(bytes))

}
