package walmart

import (
	"context"
	"encoding/json"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	initclients "travel_ai_search/search/init_clients"
	llmutil "travel_ai_search/search/llm"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

var agent_route_prompt = `
你是一个智能需求解析器，负责分析用户与电商购物助手的对话内容，从而决定下一步应执行的任务。当前系统具备以下功能：
1. 调用 search() 检索商品，为用户推荐多个商品。调用该功能时，需要从对话中提取必要的参数，参数格式如下：
{
    "category": str,         # 商品类别（如“手机”“护肤品”）
    "keywords": list,        # 关键词（如“快充”“抗敏感”）
    "attributes": map,      # 属性（如{"color":"红色","price":"1000","size":"长宽：2m*1.4m"}）
}
2. 调用 productDetail() 接口获取详细信息，用于解答用户问题或帮助用户了解已推荐商品。调用该功能时，需要从对话中提取商品 ID 作为参数，参数格式如下：
{
    "product_ids": list        # 商品 ID（从 search 结果中获取,如["1235","90876534"]）
}
3. 对话模式：与用户闲聊或引导用户补充有用的信息。
{
	"question": str 			# 进一步提问内容
}

历史对话上下文:
"""
{{.context}}
"""

用户当前输入：
“”“
{{.input}}
”“”

分析对话时，请遵循以下规则：
1. 如果用户的对话中明确给出了商品类别、至少一个关键词、商品属性（或可从对话中合理推断商品的属性），则调用 search() 接口进行商品检索。
2. 如果用户明确提及想要了解某一或多个具体商品的详细信息（通常是之前推荐过的商品），则调用 productDetail() 接口获取该商品的详细信息。
3. 如果用户的对话信息不足，无法明确商品类别、关键属性、价格范围等必要信息，或者存在模糊不清的表述，进入对话模式，通过友好的提问引导用户补充有用的信息。
4. 只输出以下的json：
{
    "next_task": "search" | "productDetail" | "dialogue",
    "parameters": {
        # 根据 next_task 的值填写相应的参数
        # 如果 next_task 是 "search"，填写 search 接口所需的参数
        # 如果 next_task 是 "productDetail"，填写 productDetail 接口所需的参数
        # 如果 next_task 是 "dialogue"，填写 dialogue 接口所需的参数
    },
    "reason": str  # 简短解释为什么选择执行当前任务
}
`

type TaskDialog struct {
	Context  string `json:"context"`
	Input    string `json:"input"`
	Expected struct {
		NextTask   string                 `json:"next_task"`
		Parameters map[string]interface{} `json:"parameters"`
	} `json:"expected"`
}

var test_dialogs = `
	[
		{
			"context": "",
			"input": "我想买一台笔记本电脑",
			"expected": {
				"next_task": "dialogue",
				"parameters": {}
			}
		},
		{
			"context": "用户：我想买一台笔记本电脑\n助手：您对笔记本电脑有什么具体需求吗？比如预算、品牌或用途？",
			"input": "预算5000左右，主要用来办公",
			"expected": {
				"next_task": "search",
				"parameters": {
					"category": "笔记本电脑",
					"keywords": ["办公"],
					"price_range": "4000-6000",
					"sort_by": "评分"
				}
			}
		},
		{
			"context": "用户：请帮我推荐几款手机\n助手：您对手机有什么具体要求吗？比如价格、品牌或功能？\n用户：3000元以下的华为手机",
			"input": "要拍照好的",
			"expected": {
				"next_task": "search",
				"parameters": {
					"category": "手机",
					"keywords": ["华为", "拍照"],
					"price_range": "0-3000",
					"sort_by": "评分"
				}
			}
		},
		{
			"context": "用户：刚才推荐的第三款手机能详细介绍一下吗？",
			"input": "它的电池容量是多少？",
			"expected": {
				"next_task": "productDetail",
				"parameters": {
					"product_id": "刚才推荐的第三款手机ID"
				}
			}
		},
		{
			"context": "用户：有没有适合送女朋友的礼物？",
			"input": "预算1000-2000",
			"expected": {
				"next_task": "dialogue",
				"parameters": {}
			}
		},
		{
			"context": "用户：有没有适合送女朋友的礼物？\n助手：您女朋友有什么兴趣爱好吗？比如美妆、数码或者首饰？\n用户：她喜欢化妆",
			"input": "想要一套护肤品",
			"expected": {
				"next_task": "search",
				"parameters": {
					"category": "护肤品",
					"keywords": ["套装"],
					"price_range": "1000-2000",
					"sort_by": "评分"
				}
			}
		},
		{
			"context": "用户：我想买运动鞋",
			"input": "耐克的，跑步用",
			"expected": {
				"next_task": "search",
				"parameters": {
					"category": "运动鞋",
					"keywords": ["耐克", "跑步"],
					"price_range": "",
					"sort_by": "评分"
				}
			}
		},
		{
			"context": "用户：刚才看的那双红色跑鞋",
			"input": "有38码的吗？",
			"expected": {
				"next_task": "productDetail",
				"parameters": {
					"product_id": "红色跑鞋ID"
				}
			}
		},
		{
			"context": "用户：天气真好",
			"input": "是啊，适合出去玩",
			"expected": {
				"next_task": "dialogue",
				"parameters": {}
			}
		},
		{
			"context": "用户：帮我找下儿童玩具",
			"input": "3岁男孩玩的，安全无毒的",
			"expected": {
				"next_task": "search",
				"parameters": {
					"category": "儿童玩具",
					"keywords": ["3岁", "男孩", "安全无毒"],
					"price_range": "",
					"sort_by": "评分"
				}
			}
		}
	]

`

func TestTaskRoute(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	logger.SetReportCaller(true)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config
	initclients.Start_client(config)
	defer initclients.Stop_client()
	llmutil.InitMemHistoryStoreInstance(5)

	testDialogs := make([]TaskDialog, 0)
	err = json.Unmarshal([]byte(test_dialogs), &testDialogs)
	if err != nil {
		t.Errorf("%+v", err)
	}

	llm, err := NewDashScopeModel()
	if err != nil {
		t.Errorf("%+v", err)
	}
	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		agent_route_prompt,
		[]string{"context", "input"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	multiRoundDialogs := &TaskDialog{
		Context: "human:请推荐我一些好吃的饼干\nassistant:[{\"product_id\":\"9897362\",\"title\":\"巧克力芝麻饼干\"},{\"product_id\":\"9894362\",\"title\":\"瑞士曲奇卷\"}]\n",
		Input:   "它们含有鸡蛋吗？",
		Expected: struct {
			NextTask   string                 "json:\"next_task\""
			Parameters map[string]interface{} "json:\"parameters\""
		}{
			NextTask: "",
			Parameters: map[string]interface{}{
				"product_ids": "9897362",
			},
		},
	}
	result, err := callLLM(multiRoundDialogs, llm, prompt)
	if err != nil {
		logger.Errorf("%+v", err)
	}
	logger.Info(result)

	// result, err = callLLM(&testDialogs[3], llm, prompt)
	// if err != nil {
	// 	logger.Errorf("%+v", err)
	// }
	logger.Info(result)
}

func callLLM(testDialog *TaskDialog, llm llms.Model, prompt prompts.FormatPrompter) (result any, err error) {
	ctx := context.Background()
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["input"] = testDialog.Input
	inputs["context"] = testDialog.Context
	llmResult, err := llmChain.Call(ctx, inputs)
	if err != nil {
		return "", err
	}
	logger.Info(llmResult)
	// 假设 result 是一个字符串，尝试解析结果为 JSON
	var output map[string]interface{}
	err = json.Unmarshal([]byte(llmResult["text"].(string)), &output)
	if err != nil {
		logger.Errorf("Failed to unmarshal llmChain result for test case with input %s and context %s: %+v", testDialog.Input, testDialog.Context, err)
	}
	// 验证输出是否符合预期
	if output["next_task"] != testDialog.Expected.NextTask {
		logger.Errorf("Expected next_task %s, but got %v for test case with input %s and context %s", testDialog.Expected.NextTask, output["next_task"], testDialog.Input, testDialog.Context)
	}
	return llmResult["text"], nil
}
