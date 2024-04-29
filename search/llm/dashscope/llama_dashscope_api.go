package dashscope

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/llm"

	"github.com/tmc/langchaingo/llms"

	logger "github.com/sirupsen/logrus"
)

type LlamaDashScope struct {
	Room      string
	ModelName string
}

type DashScopeRequest struct {
	Model string         `json:"model"`
	Input DashScopeInput `json:"input"`
}

type DashScopeInput struct {
	Messages []DashScopeMessage `json:"messages"`
}

type DashScopeMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type DashScopeResponse struct {
	RequestId string          `json:"request_id"`
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Output    DashScopeOutput `json:"output"`
	Usage     DashScopeUsage  `json:"usage"`
}

type DashScopeOutput struct {
	Text string `json:"text"`
}

type DashScopeUsage struct {
	OutputTokens int64 `json:"output_tokens"`
	InputTokens  int64 `json:"input_tokens"`
}

func (model *LlamaDashScope) GetChatRes(messages []llms.ChatMessage, msgListener chan string) (string, int64) {
	client := &http.Client{}

	dashscopeMessages := make([]DashScopeMessage, 0, len(messages))

	for _, msg := range messages {
		switch msg.GetType() {
		case llms.ChatMessageTypeSystem:
			dashscopeMessages = append(dashscopeMessages, DashScopeMessage{Role: llm.ROLE_SYSTEM, Content: msg.GetContent()})
		case llms.ChatMessageTypeHuman:
			dashscopeMessages = append(dashscopeMessages, DashScopeMessage{Role: llm.ROLE_USER, Content: msg.GetContent()})
		default:
			dashscopeMessages = append(dashscopeMessages, DashScopeMessage{Role: llm.ROLE_ASSISTANT, Content: msg.GetContent()})
		}

	}

	reqBody := DashScopeRequest{
		Model: model.ModelName,
		Input: DashScopeInput{
			Messages: dashscopeMessages,
		},
	}

	buf, _ := json.Marshal(reqBody)

	logger.Debug(string(buf))

	req, _ := http.NewRequest("POST", conf.GlobalConfig.DashScopeLLM.HostUrl, bytes.NewBuffer(buf))

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", conf.GlobalConfig.DashScopeLLM.Key))
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("request dashscope err:%s", err.Error())
		return "", 0
	}
	defer resp.Body.Close()

	respBuf, err := io.ReadAll(resp.Body)

	if err != nil {
		logger.Errorf("request dashscope err:%s", err.Error())
		return "", 0
	}
	logger.Debug(string(respBuf))
	if resp.StatusCode != http.StatusOK {
		logger.Errorf("request dashscope status:%d,response:%s, err:%s", resp.StatusCode, string(respBuf), err.Error())
		return "", 0
	}
	result := DashScopeResponse{}
	err = json.Unmarshal(respBuf, &result)
	if err != nil {
		logger.Errorf("request dashscope status:%d,response:%s, err:%s", resp.StatusCode, string(respBuf), err.Error())
		return "", 0
	}

	return result.Output.Text, result.Usage.InputTokens + result.Usage.OutputTokens

}
