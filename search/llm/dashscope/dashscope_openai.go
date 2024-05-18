package dashscope

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/llm"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type CompletionResponse struct {
	ID      string  `json:"id,omitempty"`
	Created float64 `json:"created,omitempty"`
	Choices []struct {
		FinishReason string      `json:"finish_reason,omitempty"`
		Index        float64     `json:"index,omitempty"`
		Logprobs     interface{} `json:"logprobs,omitempty"`
		Text         string      `json:"text,omitempty"`
	} `json:"choices,omitempty"`
	Model  string `json:"model,omitempty"`
	Object string `json:"object,omitempty"`
	Usage  struct {
		CompletionTokens float64 `json:"completion_tokens,omitempty"`
		PromptTokens     float64 `json:"prompt_tokens,omitempty"`
		TotalTokens      float64 `json:"total_tokens,omitempty"`
	} `json:"usage,omitempty"`
}

type DashScopeOpenAIModel struct {
	Room      string
	ModelName string
}

func (model *DashScopeOpenAIModel) GetChatRes(messages []llms.ChatMessage, msgListener chan string) (string, int64) {
	client, err := model.newClient()
	if err != nil {
		return "", 0
	}

	contents := make([]llms.MessageContent, 0, len(messages))
	for _, msg := range messages {
		contents = append(contents, llms.MessageContent{
			Role:  msg.GetType(),
			Parts: []llms.ContentPart{llms.TextPart(msg.GetContent())},
		})
	}

	if logger.IsLevelEnabled(logger.DebugLevel) {
		buf, _ := json.Marshal(contents)
		logger.Debug(string(buf))
	}
	var seqno = time.Now().UnixNano()
	rsp, err := client.GenerateContent(context.Background(), contents,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			logger.Infoln("chunk:", string(chunk))
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
			return nil
		}))

	if err != nil {
		fmt.Println("request err:", err.Error())
		return "", 0
	}

	c1 := rsp.Choices[0]
	/**
	GenerationInfo: map[string]any{
		"CompletionTokens": result.Usage.CompletionTokens,
		"PromptTokens":     result.Usage.PromptTokens,
		"TotalTokens":      result.Usage.TotalTokens,
	},
	*/
	buf, _ := json.Marshal(rsp)
	logger.Infoln(string(buf))

	tokens, ok := c1.GenerationInfo["TotalTokens"].(int)
	if !ok {
		buf, _ := json.Marshal(c1.GenerationInfo)
		logger.Infoln(string(buf))
	}
	logger.Infoln("TotalTokens:", tokens)
	logger.Infoln("content:", c1.Content)
	logger.Infoln("stop:", c1.StopReason)
	//todo: dashscope bug
	{
		inputBuf, _ := json.Marshal(contents)
		outputBuf, _ := json.Marshal(rsp)
		tokens = len(inputBuf) + len(outputBuf)
	}
	return c1.Content, int64(tokens)
}

func (model *DashScopeOpenAIModel) newClient() (llms.Model, error) {
	opts := make([]openai.Option, 0)
	opts = append(opts, openai.WithBaseURL(conf.GlobalConfig.DashScopeLLM.OpenaiUrl))
	opts = append(opts, openai.WithModel(model.ModelName))
	opts = append(opts, openai.WithToken(conf.GlobalConfig.DashScopeLLM.Key))
	llm, err := openai.New(opts...)
	return llm, err
}
