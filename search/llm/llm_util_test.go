package llm

import (
	"encoding/json"
	"testing"
	"time"
)

func TestChatHistoryBuff(t *testing.T) {
	userId := "aaaaa"
	room := "chat"
	AddChatHistory(userId, room, "我是谁", "你是张三1")
	AddChatHistory(userId, room, "我是谁", "你是李四2")

	msgs := LoadChatHistory(userId, room)

	if len(msgs) != 4 {
		t.Fatalf("expect:%d,actual:%d", 4, len(msgs))
	}
	AddChatHistory(userId, room, "我是谁", "你是张三3")
	AddChatHistory(userId, room, "我是谁", "你是李四4")
	AddChatHistory(userId, room, "我是谁", "你是张三5")
	AddChatHistory(userId, room, "我是谁", "你是李四6")
	msgs = LoadChatHistory(userId, room)
	if len(msgs) != 10 {
		t.Fatalf("expect:%d,actual:%d", 10, len(msgs))
	}

	req := map[string]any{"text": msgs}
	bytes, _ := json.Marshal(req)

	t.Error(string(bytes))

}

type TestMsg struct {
	Id string
	V  int
}

func test1(v any) string {
	msg := v.(TestMsg)
	msg.Id = "test1"
	msg.V++

	str, _ := json.Marshal(msg)
	return string(str)
}

func test2(msg *TestMsg) string {
	msg.Id = "test2"
	msg.V++

	str, _ := json.Marshal(msg)
	return string(str)
}

func test3(msg any) string {
	v := msg.(*TestMsg)
	str, _ := json.Marshal(v)
	return string(str)
}

func test4(msg any) string {
	v := msg.(TestMsg)
	str, _ := json.Marshal(v)
	return string(str)
}

func TestAny(t *testing.T) {
	msg := TestMsg{
		Id: "a",
		V:  1,
	}

	str1 := test1(msg)
	origin, _ := json.Marshal(msg)
	if str1 == string(origin) {
		t.Errorf("str1:%s,origin:%s", str1, origin)
	}

	msg = TestMsg{
		Id: "a",
		V:  1,
	}
	str2 := test2(&msg)

	origin, _ = json.Marshal(msg)
	if str2 != string(origin) {
		t.Errorf("str2:%s,origin:%s", str2, origin)
	}

	tmp := msg
	tmp.Id = "tmp"

	origin, _ = json.Marshal(msg)
	str3, _ := json.Marshal(tmp)
	if string(str3) == string(origin) {
		t.Errorf("str3:%s,origin:%s", str3, origin)
	}

	tmp2 := &msg
	str4 := test3(tmp2)
	origin, _ = json.Marshal(msg)

	if str4 != string(origin) {
		t.Errorf("str4:%s,origin:%s", str4, origin)
	}

	var a = 0
	var b = 0

	go func() {
		time.Sleep(time.Millisecond)
		a++
		b++
	}()
	a = 10
	b = 10
	time.Sleep(time.Millisecond * 3)
	//t.Errorf("a:%d,b:%d", a, b)

}
