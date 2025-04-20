package walmart

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/user"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

func NewDashScopeModel() (*openai.LLM, error) {
	opts := make([]openai.Option, 0)
	opts = append(opts, openai.WithBaseURL(conf.GlobalConfig.DashScopeLLM.OpenaiUrl))
	opts = append(opts, openai.WithModel(conf.GlobalConfig.DashScopeLLM.Model))
	opts = append(opts, openai.WithToken(conf.GlobalConfig.DashScopeLLM.Key))
	opts = append(opts, openai.WithCallback(ShoppingFlowLogHandler{}))
	llm, err := openai.New(opts...)
	return llm, err
}

func conversationContext(curUser user.User, room string) string {
	chatHistorys := llmutil.GetHistoryStoreInstance().LoadChatHistoryForLLM(curUser.UserId, room)
	var strBuilder strings.Builder
	for i, msg := range chatHistorys {
		if msg, ok := msg.(*llmutil.Message); ok {
			if (time.Now().UnixMilli() - msg.GetTimestamp()) > 10*60*1000 {
				continue
			}
		}
		role := ""
		switch msg.GetType() {
		case llms.ChatMessageTypeSystem:
			role = llmutil.ROLE_SYSTEM
		case llms.ChatMessageTypeHuman:
			role = llmutil.ROLE_USER
		default:
			role = llmutil.ROLE_ASSISTANT
		}
		if i > 0 {
			strBuilder.WriteString("\r\n")
		}
		strBuilder.WriteString(role)
		strBuilder.WriteString(":")
		strBuilder.WriteString(msg.GetContent())
	}
	return strBuilder.String()
}

func saveChatHistory(curUser user.User, room, query, response string) {
	llmutil.GetHistoryStoreInstance().AddChatHistory(curUser.UserId, room, query, response)
}

func doLLM(logEntry *logger.Entry, llmChain *chains.LLMChain,
	inputs map[string]any, ctx context.Context) (map[string]any, error) {

	jsonContent, err := doLLMRespJsonStr(logEntry, llmChain, inputs, ctx)
	if err != nil {
		return nil, err
	}
	respMap := make(map[string]any)
	err = json.Unmarshal([]byte(jsonContent), &respMap)
	if err != nil {
		//todo:重试
		logEntry.Errorf("%s unmarshal err:%s", jsonContent, err.Error())
		return nil, err
	}
	return respMap, nil
}

func doLLMRespList(logEntry *logger.Entry, llmChain *chains.LLMChain,
	inputs map[string]any, ctx context.Context) ([]map[string]any, error) {

	jsonContent, err := doLLMRespJsonStr(logEntry, llmChain, inputs, ctx)
	if err != nil {
		return nil, err
	}
	respMap := make([]map[string]any, 0)
	err = json.Unmarshal([]byte(jsonContent), &respMap)
	if err != nil {
		//todo:重试
		logEntry.Errorf("%s unmarshal err:%s", jsonContent, err.Error())
		return nil, err
	}
	return respMap, nil
}

func doLLMRespJsonStr(logEntry *logger.Entry, llmChain *chains.LLMChain,
	inputs map[string]any, ctx context.Context) (string, error) {

	llmStartTime := time.Now().UnixMilli()
	result, err := llmChain.Call(ctx, inputs)
	llmEndTime := time.Now().UnixMilli()

	logEntry.Infof("llm time:%d", llmEndTime-llmStartTime)
	if err != nil {
		//todo:重试
		logEntry.Error("call llm err:", err.Error())
		return "", err
	}

	text, ok := result["text"]

	if !ok {

		buf, _ := json.Marshal(result)
		logEntry.Errorf("llm response:%s", string(buf))
		return "", errors.New("llm response err")
	}
	content, ok := text.(string)
	if !ok {
		kindStr := reflect.TypeOf(text).Kind().String()
		logEntry.Errorf("text type:%s", kindStr)
		return "", errors.New("err type:" + kindStr)
	}

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")

	list_start := strings.Index(content, "[")
	list_end := strings.LastIndex(content, "]")

	if list_start >= 0 && list_start < start {
		start = list_start
		end = list_end
	}

	if start == -1 || end == -1 {
		return "", common.Errorf(fmt.Sprintf("parse json:%d,%d", start, end), errors.New("invalid json:"+content))
	}
	jsonContent := content[start : end+1]
	return jsonContent, nil
}

func doLLMRespText(logEntry *logger.Entry, llmChain *chains.LLMChain,
	inputs map[string]any, ctx context.Context) (string, error) {

	llmStartTime := time.Now().UnixMilli()
	result, err := llmChain.Call(ctx, inputs)
	llmEndTime := time.Now().UnixMilli()

	logEntry.Infof("llm time:%d", llmEndTime-llmStartTime)
	if err != nil {
		//todo:重试
		logEntry.Error("call llm err:", err.Error())
		return "", err
	}

	text, ok := result["text"]

	if !ok {

		buf, _ := json.Marshal(result)
		logEntry.Errorf("llm response:%s", string(buf))
		return "", errors.New("llm response err")
	}
	content, ok := text.(string)
	if !ok {
		kindStr := reflect.TypeOf(text).Kind().String()
		logEntry.Errorf("text type:%s", kindStr)
		return "", errors.New("err type:" + kindStr)
	}
	return content, nil
}

func route(curUser user.User, room, flowId, query string) ([]string, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()
	logEntry := logger.WithField("uid", curUser.UserId)
	if err != nil {
		logEntry.Errorf("create llm client err:%s", err.Error())
		return nil, err
	}

	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.AgentRouting,
		[]string{"context", "agent_list", "question"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["question"] = query
	inputs["context"] = conversationContext(curUser, room)
	inputs["agent_list"] = conf.AgentTemplate(conf.GlobalConfig, flowId)

	selectMap, err := doLLM(logEntry, llmChain, inputs, ctx)
	if err != nil {
		logEntry.Errorf("do llm err:%s", err.Error())
		return nil, err
	}

	if reason, ok := selectMap["reason"]; ok {
		logEntry.Infof("route reason:%s", reason)
	}
	agents := make([]string, 0)
	if agentV, ok := selectMap["agents"]; ok {
		logEntry.Infof("agents:%v", agentV)
		switch agentNameList := agentV.(type) {
		case []any:
			for _, agentNameV := range agentNameList {
				switch name := agentNameV.(type) {
				case string:
					agents = append(agents, name)
				default:
					logEntry.Errorf("err agent name type,%v", agentNameV)
				}
			}
		default:
			logEntry.Errorf("err agents type,%v", agentV)
		}
	}
	logEntry.Infof("agents: %v\n", agents)
	return agents, nil
}
