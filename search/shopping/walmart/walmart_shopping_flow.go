package walmart

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
	"travel_ai_search/search/conf"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/user"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

const SHOPPING_FLOWID = "shopping"

type ShoppingEngine struct {
}

var chatLoggerOnce sync.Once
var chatLogger *logger.Logger
var chatLogFile *os.File

func GetChatLogger() *logger.Logger {
	chatLoggerOnce.Do(func() {
		var err error
		chatLogFile, err = os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("无法打开日志文件: %v\n", err)
			return
		}
		chatLogger = logger.New()
		chatLogger.SetOutput(chatLogFile)
	})
	return chatLogger
}

func CloseLogFile() {
	if chatLogFile != nil {
		chatLogFile.Close()
	}
}

// func (engine *ShoppingEngine) Flow(curUser user.User, room, input string) (llmutil.TaskOutputType, any, error) {
// 	agentNames, err := route(curUser, room, SHOPPING_FLOWID, input)

// 	if err != nil {
// 		return "", nil, err
// 	}

// 	tasks := make([]shopping.TaskEngine, 0)

// 	for _, name := range agentNames {
// 		switch name {
// 		case "product_list":
// 			tasks = append(tasks, &WalmartProductListTask{})
// 		case "product_detail":
// 			tasks = append(tasks, &WalmartProductDetailTask{})
// 		case "default_chat":
// 			tasks = append(tasks, &WalmartChatTask{})
// 		}
// 	}

// 	if len(tasks) == 0 {
// 		tasks = append(tasks, &WalmartChatTask{})
// 	}

// 	query := input
// 	var outputType llmutil.TaskOutputType
// 	var output any
// 	for _, task := range tasks {
// 		outputType, output, err = task.Run(curUser, room, query)
// 		if err != nil {
// 			taskType := reflect.TypeOf(task).Kind().String()
// 			logger.Infof("exec task %s err :%s", taskType, err.Error())
// 			return "", nil, err
// 		}
// 		query = task.FormatOutput()
// 	}
// 	saveChatHistory(curUser, room, input, query)
// 	return outputType, output, nil
// }

func (engine *ShoppingEngine) Flow(curUser user.User, room, input string) (llmutil.TaskOutputType, any, error) {
	// funName, args, err := funcRoute(curUser, room, input)
	// if err != nil {
	// 	return "", nil, fmt.Errorf("no function")
	// }
	funName, args, err := taskRoute(curUser, room, input)
	if err != nil {
		//return "", nil, fmt.Errorf("no function")
		engine.CallTask(curUser, room, input, "chat", "")
	}
	return engine.CallTask(curUser, room, input, funName, args)
}

func (engine *ShoppingEngine) CallTask(curUser user.User, room string, query string, funName string,
	args string) (llmutil.TaskOutputType, any, error) {
	logEntry := logger.WithField("uid", curUser.UserId)
	switch funName {

	case "searchProducts":
		//args := toolCall.FunctionCall.Arguments
		var queryArgs struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(args), &queryArgs); err != nil {
			logEntry.Errorf("searchProducts args(%s) unmarshal err:%s", args, err.Error())
		}
		logEntry.Infof("search func argument:query:%s,input query:%s", queryArgs.Query, query)
		task := &WalmartProductListTask{}
		outputType, output, err := task.Run(curUser, room, query)
		if err == nil {
			GetChatLogger().Infof("input:%v", query)
			GetChatLogger().Infof("answer:%+v", task.FormatOutput())
			saveChatHistory(curUser, room, query, task.FormatOutput())
		}

		return outputType, output, err

	case "getProductDetail":
		var idArgs struct {
			ProductIds []string `json:"product_ids"`
		}

		if err := json.Unmarshal([]byte(args), &idArgs); err != nil {
			logEntry.Errorf("getProductDetail args(%s) unmarshal err:%s", args, err.Error())
		}
		task := &WalmartProductDetailTask{}
		outputType, output, err := task.Call(curUser, room, idArgs.ProductIds)
		if err == nil {
			GetChatLogger().Infof("input:%v", query)
			GetChatLogger().Infof("answer:%+v", task.FormatOutput())
			saveChatHistory(curUser, room, query, task.FormatOutput())
		}
		return outputType, output, err

	case "chat":
		//args := toolCall.FunctionCall.Arguments
		if len(args) > 0 {
			var queryArgs struct {
				Answer string `json:"answer"`
			}
			if err := json.Unmarshal([]byte(args), &queryArgs); err != nil {
				logEntry.Errorf("chat args(%s) unmarshal err:%s", args, err.Error())
			} else {
				logEntry.Infof("chat func argument:query:%s", queryArgs.Answer)
				GetChatLogger().Infof("input:%v", query)
				GetChatLogger().Infof("answer:%+v", queryArgs.Answer)
				saveChatHistory(curUser, room, query, queryArgs.Answer)
				return llmutil.CHAT_TYPE_MSG, queryArgs.Answer, nil
			}
		}

		task := &WalmartChatTask{}
		outputType, output, err := task.Run(curUser, room, query)
		if err == nil {
			GetChatLogger().Infof("input:%v", query)
			GetChatLogger().Infof("answer:%+v", task.FormatOutput())
			saveChatHistory(curUser, room, query, task.FormatOutput())
		}
		return outputType, output, err

	default:
		logEntry.Errorf("undefined func:%s", funName)
	}
	return "", nil, fmt.Errorf("no function")
}

var availableTools = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getProductDetail",
			Description: "get detailed information of the products, such as ID, product name, product description(contains taste/purpose/ingredient/component list etc)",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"product_ids": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
						"description": "list of the product id",
					},
				},
				"required": []string{"product_ids"},
			},
		},
	},
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "searchProducts",
			Description: "searches for product lists. ",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "the rewrited query words for searching",
					},
				},
				"required": []string{"query"},
			},
		},
	},
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "chat",
			Description: "When no suitable agent is found, choose me to enter the dialogue mode",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"input": map[string]any{
						"type":        "string",
						"description": "an independent and complete question that the user wants to express",
					},
				},
				"required": []string{"input"},
			},
		},
	},
}

func funcRoute(curUser user.User, room, query string) (string, string, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()
	logEntry := logger.WithField("uid", curUser.UserId)
	if err != nil {
		logEntry.Errorf("create llm client err:%s", err.Error())
		return "", "", err
	}
	chatHistorys := llmutil.GetHistoryStoreInstance().LoadChatHistoryForLLM(curUser.UserId, room)
	messageHistory := make([]llms.MessageContent, 0)
	for _, history := range chatHistorys {

		if msg, ok := history.(*llmutil.Message); ok {
			if (time.Now().UnixMilli() - msg.GetTimestamp()) > 10*60*1000 {
				continue
			}
		}

		messageHistory = append(messageHistory, llms.TextParts(history.GetType(), history.GetContent()))
	}
	messageHistory = append(messageHistory, llms.TextParts(llms.ChatMessageTypeHuman, query))

	resp, err := llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		logEntry.Errorf("call llm err:%s", err.Error())
		return "", "", err
	}
	logEntry.Infof("toolcall num:%d", len(resp.Choices[0].ToolCalls))

	for _, toolCall := range resp.Choices[0].ToolCalls {
		logEntry.Infof("toolcall:%s(%s)", toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)
		return toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments, nil
	}
	return "", "", fmt.Errorf("no function")
}

func taskRoute(curUser user.User, room, query string) (string, string, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()
	logEntry := logger.WithField("uid", curUser.UserId)
	if err != nil {
		logEntry.Errorf("create llm client err:%s", err.Error())
		return "", "", err
	}

	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.TasKRouter,
		[]string{"context", "input"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["input"] = query
	inputs["context"] = conversationContext(curUser, room)
	if logger.IsLevelEnabled(logger.DebugLevel) {
		logEntry.Debugf("taskRoute inputs:%v", inputs)
	}
	taskRouteMap, err := doLLM(logEntry, llmChain, inputs, ctx)
	if err != nil {
		logEntry.Errorf("do llm err:%s", err.Error())
	}
	if logger.IsLevelEnabled(logger.DebugLevel) {
		logEntry.Debugf("taskRoute response taskRouteMap:%+v", taskRouteMap)
	}
	if reason, ok := taskRouteMap["reason"]; ok {
		logEntry.Infof("taskRoute reason:%+v", reason)
	}
	taskName := ""
	if nextTaskName, ok := taskRouteMap["next_task"]; ok {
		taskName = nextTaskName.(string)
		logEntry.Infof("next_task:%v", nextTaskName)
	}

	params, paramExist := taskRouteMap["parameters"]

	//"search" | "productDetail" | "dialogue"
	switch taskName {
	case "search":
		//if paramExist {
		//paramMap := params.(map[string]interface{})
		//category, ok := paramMap["category"]
		//keywords
		//attributes
		//}
		return "searchProducts", "", nil
	case "productDetail":
		if paramExist {
			paramMap := params.(map[string]interface{})
			productIds, ok := paramMap["product_ids"]
			productIdStrList := make([]string, 0)
			if ok {
				productIdList := productIds.([]interface{})
				for _, productId := range productIdList {
					switch productId := productId.(type) {
					case string:
						productIdStrList = append(productIdStrList, productId)
					case float64:
						productIdStrList = append(productIdStrList, fmt.Sprintf("%d", int(productId)))
					}
				}
				if len(productIdStrList) > 0 {
					buf, err := json.Marshal(map[string][]string{"product_ids": productIdStrList})
					if err != nil {
						return "getProductDetail", "", fmt.Errorf("productIds marshal err:%s", err.Error())
					}
					return "getProductDetail", string(buf), nil
				}
			}
		}
		return "getProductDetail", "", fmt.Errorf("parameter error")
	case "dialogue":
		if paramExist {
			paramMap := params.(map[string]interface{})
			question, ok := paramMap["question"]

			if ok {
				buf, err := json.Marshal(map[string]string{"answer": question.(string)})
				if err != nil {
					return "chat", "", fmt.Errorf("parameter error")
				}
				return "chat", string(buf), nil
			}
		}
		return "chat", "", fmt.Errorf("parameter error")
	default:
		return "", "", fmt.Errorf("no function")
	}
}
