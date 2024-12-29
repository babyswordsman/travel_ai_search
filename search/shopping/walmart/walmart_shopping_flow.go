package walmart

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/user"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

const SHOPPING_FLOWID = "shopping"

type ShoppingEngine struct {
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
	funName, args, err := funcRoute(curUser, room, input)
	if err != nil {
		return "", nil, fmt.Errorf("no function")
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
		logEntry.Infof("search func argument:query:%s", queryArgs.Query)
		task := &WalmartProductListTask{}
		outputType, output, err := task.Run(curUser, room, query)
		if err == nil {
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
			saveChatHistory(curUser, room, query, task.FormatOutput())
		}
		return outputType, output, err

	case "chat":
		//args := toolCall.FunctionCall.Arguments
		var queryArgs struct {
			Input string `json:"input"`
		}
		if err := json.Unmarshal([]byte(args), &queryArgs); err != nil {
			logEntry.Errorf("chat args(%s) unmarshal err:%s", args, err.Error())
		}
		logEntry.Infof("chat func argument:query:%s", queryArgs.Input)
		task := &WalmartChatTask{}
		outputType, output, err := task.Run(curUser, room, query)
		if err == nil {
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
