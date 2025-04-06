package walmart

import (
	"encoding/json"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/user"

	"github.com/bytedance/mockey"
	logger "github.com/sirupsen/logrus"
	"github.com/smartystreets/goconvey/convey"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestFuncRoute(t *testing.T) {
	/**
	StopReason: tool_calls
	GenerationInfo:
	        PromptTokens: 156
	         TotalTokens: 173
	    CompletionTokens: 17
	FuncCall:  searchProducts {"query":"biscuits"}
	time="2025-01-01T11:51:01+08:00" level=info msg="toolcall num:1" func=travel_ai_search/search/shopping/walmart.funcRoute file="/home/casperding/Desktop/workspace/travel_ai_search/search/shopping/walmart/walmart_shopping_flow.go:210" uid=_
	time="2025-01-01T11:51:01+08:00" level=info msg="toolcall:searchProducts({\"query\":\"biscuits\"})" func=travel_ai_search/search/shopping/walmart.funcRoute file="/home/casperding/Desktop/workspace/travel_ai_search/search/shopping/walmart/walmart_shopping_flow.go:213" uid=_
	✔✔✔✔query:biscuits
	*/
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
	llmutil.InitMemHistoryStoreInstance(10)

	mockey.PatchConvey("test funRoute", t, func() {
		mockey.PatchConvey("success", func() {
			choices := make([]*llms.ContentChoice, 0)
			choice := &llms.ContentChoice{
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "searchProducts",
							Arguments: "{\"query\":\"biscuits\"}",
						},
					},
				},
			}
			choices = append(choices, choice)
			resp := &llms.ContentResponse{
				Choices: choices,
			}
			mockey.Mock((*openai.LLM).GenerateContent).Return(resp, nil).Build()

			funName, args, err := funcRoute(user.EmpytUser, "shopping", "Recommend several kinds of delicious biscuits to me")
			convey.So(err, convey.ShouldBeNil)
			convey.So(funName, convey.ShouldBeIn, []string{"searchProducts", "getProductDetail", "chat"})
			argMap := make(map[string]any)
			err = json.Unmarshal([]byte(args), &argMap)
			convey.So(err, convey.ShouldBeNil)
			switch funName {
			case "searchProducts":
				v, ok := argMap["query"]
				convey.So(ok, convey.ShouldBeTrue)
				convey.Printf("query:%v", v)
			case "getProductDetail":
				v, ok := argMap["product_ids"]
				convey.So(ok, convey.ShouldBeTrue)
				convey.Printf("product_ids:%v", v)
			case "chat":
				v, ok := argMap["input"]
				convey.So(ok, convey.ShouldBeTrue)
				convey.Printf("input:%v", v)
			default:
				t.Errorf("err funcName:%v", funName)
			}
		})

		mockey.PatchConvey("test task run success", func() {
			choices := make([]*llms.ContentChoice, 0)
			choice := &llms.ContentChoice{
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "searchProducts",
							Arguments: "{\"query\":\"biscuits\"}",
						},
					},
				},
			}
			choices = append(choices, choice)
			resp := &llms.ContentResponse{
				Choices: choices,
			}
			mockey.Mock((*openai.LLM).GenerateContent).Return(resp, nil).Build()
			query := "Recommend several kinds of delicious biscuits to me"
			funName, args, err := funcRoute(user.EmpytUser, "shopping", query)
			convey.So(err, convey.ShouldBeNil)
			convey.So(funName, convey.ShouldBeIn, []string{"searchProducts", "getProductDetail", "chat"})
			argMap := make(map[string]any)
			err = json.Unmarshal([]byte(args), &argMap)
			convey.So(err, convey.ShouldBeNil)
			engine := &ShoppingEngine{}

			intentMapStr := `{"Complete information":"Crispy cookies","Question 1":1,"Question 2":"Food","Question 3":"Cookies","Question 4":{"Texture":"Crispy"}}`
			intentMap := make(map[string]any)
			err = json.Unmarshal([]byte(intentMapStr), &intentMap)
			convey.So(err, convey.ShouldBeNil)
			mockey.Mock(doLLM).Return(intentMap, nil).Build()

			engine.CallTask(user.EmpytUser, "shopping", query, funName, args)
		})
	})
}

func TestFuncRouteWithProductDetail(t *testing.T) {
	/**
			Entering LLM with messages:
	Role: human
	Text: I want to buy some crispy cookies
	Role: ai
	Text: product id:5109506439, product name:Soeos Fortune Cookies, Approx. 40pcs Crispy Cookies, Aisle:A2, price:13.70
	product id:10292640, product name:Ginger Snaps Cookies, Ginger Cookies, 16 oz, Aisle:A1, price:6.24
	product id:46028132, product name:Arnott's Tim Tam Chewy Caramel Cookies, 6.2 oz. Tray, Aisle:A3, price:6.84
	product id:21905208, product name:Stella D'oro Cookies, Margherite Cookies, 12 oz, Aisle:A5, price:9.00
	Role: human
	Text: I like to eat sweets, which one is more suitable for me?
	Exiting LLM with response:
	StopReason: tool_calls
	GenerationInfo:
	    CompletionTokens: 31
	        PromptTokens: 324
	         TotalTokens: 355
	FuncCall:  getProductDetail {"product_ids":["5109506439","10292640","46028132","21905208"]}

	*/
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
	llmutil.InitMemHistoryStoreInstance(10)

	mockey.PatchConvey("test funRoute", t, func() {
		mockey.PatchConvey("success", func() {
			choices := make([]*llms.ContentChoice, 0)
			choice := &llms.ContentChoice{
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "getProductDetail",
							Arguments: "{\"product_ids\":[\"5109506439\",\"10292640\",\"46028132\",\"21905208\"]}",
						},
					},
				},
			}
			choices = append(choices, choice)
			resp := &llms.ContentResponse{
				Choices: choices,
			}
			mockey.Mock((*openai.LLM).GenerateContent).Return(resp, nil).Build()

			funName, args, err := funcRoute(user.EmpytUser, "shopping", "Recommend several kinds of delicious biscuits to me")
			convey.So(err, convey.ShouldBeNil)
			convey.So(funName, convey.ShouldBeIn, []string{"searchProducts", "getProductDetail", "chat"})
			argMap := make(map[string][]string)
			err = json.Unmarshal([]byte(args), &argMap)
			convey.So(err, convey.ShouldBeNil)

			productIds, ok := argMap["product_ids"]
			convey.So(ok, convey.ShouldBeTrue)
			convey.Printf("product_ids:%v", productIds)

			engine := &ShoppingEngine{}
			query := "I like to eat sweets, which one is more suitable for me?"

			// 				docs := make([]*detail.WalmartSkuResp, 0)

			// skuMap1 := map[string]string{"id":"5109506439", "name":"Soeos Fortune Cookies, Approx. 40pcs Crispy Cookies", "Aisle":"A2", "price":"13.70"}
			// skuMap2 := map[string]string{"id":"10292640", "name":"Ginger Snaps Cookies, Ginger Cookies, 16 oz", "Aisle":"A1", "price":"6.24"}
			// skuMap3 := map[string]string{"id":"46028132", "name":"Arnott's Tim Tam Chewy Caramel Cookies, 6.2 oz. Tray", "Aisle":"A3", "price":"6.84"}
			// skuMap4 := map[string]string{"id":"21905208", "name":"Stella D'oro Cookies, Margherite Cookies, 12 oz", "Aisle":"A5", "price":"9.00"}

			engine.CallTask(user.EmpytUser, "shopping", query, funName, args)
		})
	})
}
