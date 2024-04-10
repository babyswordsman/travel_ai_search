package llm

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
	Type  string      `json:"type"`
	Body  interface{} `json:"body"`
	Seqno string      `json:"seqno"`
}
