package dashscope

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/llm"

	"github.com/devinyf/dashscopego"
	"github.com/devinyf/dashscopego/qwen"
	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

type DashScopeModel struct {
	Room      string
	ModelName string
}

func (model *DashScopeModel) GetChatRes(messages []llms.ChatMessage, msgListener chan string) (string, int64) {
	cli := dashscopego.NewTongyiClient(model.ModelName, conf.GlobalConfig.DashScopeLLM.Key)
	dashscopeMessages := make([]dashscopego.TextMessage, 0, len(messages))
	for _, msg := range messages {
		role := ""
		switch msg.GetType() {
		case llms.ChatMessageTypeSystem:
			role = llm.ROLE_SYSTEM
		case llms.ChatMessageTypeHuman:
			role = llm.ROLE_USER
		default:
			role = llm.ROLE_ASSISTANT
		}
		content := qwen.TextContent{Text: msg.GetContent()}
		dashscopeMessages = append(dashscopeMessages, dashscopego.TextMessage{Role: role, Content: &content})
	}

	input := dashscopego.TextInput{
		Messages: dashscopeMessages,
	}

	var seqno = time.Now().UnixNano()
	var answer strings.Builder
	var streamCallbackFn func(ctx context.Context, chunk []byte) error = nil
	if msgListener != nil {
		streamCallbackFn = func(ctx context.Context, chunk []byte) error {
			if msgListener != nil {
				content := string(chunk)
				content = strings.ReplaceAll(content, "\r\n", "<br />")
				content = strings.ReplaceAll(content, "\n", "<br />")
				contentResp := llm.ChatStream{
					ChatType: string(llms.ChatMessageTypeAI),
					Room:     model.Room,
					Type:     llm.CHAT_TYPE_MSG,
					Body:     content, //strings.ReplaceAll(content, "\n", "<br />"),
					Seqno:    strconv.FormatInt(seqno, 10),
				}
				v, _ := json.Marshal(contentResp)
				msgListener <- string(v)
			}
			answer.WriteString(string(chunk))
			return nil
		}
	}

	req := &dashscopego.TextRequest{
		Input:       input,
		StreamingFn: streamCallbackFn,
	}
	if logger.IsLevelEnabled(logger.DebugLevel) {
		reqStr, _ := json.Marshal(req)
		logger.Debug(string(reqStr))
	}

	ctx := context.Background()
	resp, err := cli.CreateCompletion(ctx, req)

	if err != nil {
		logger.Errorf("dashscope err:%s", err.Error())
		return "", 0
	}
	if logger.IsLevelEnabled(logger.DebugLevel) {
		respStr, _ := json.Marshal(resp)
		logger.Debug(string(respStr))
	}

	if msgListener == nil {
		answer.WriteString(resp.Output.Choices[0].Message.Content.Text)
	}

	return answer.String(), int64(resp.Usage.TotalTokens)

}
