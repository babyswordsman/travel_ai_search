package shopping

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"travel_ai_search/search/conf"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/shopping/detail"
	"travel_ai_search/search/user"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

type ShoppingEngine struct {
}

type ShoppingFlowLogHandler struct {
	callbacks.LogHandler
}

type ShoppingIntent struct {
	IsShopping       bool
	Category         string
	ProductName      string
	ProductProps     map[string]string
	IndependentQuery string
}

const KEY_INDEPENDENT = "完整信息"
const KEY_IS_SHOPPING = "问题1"
const KEY_CATEGORY = "问题2"
const KEY_PRODUCT = "问题3"
const KEY_PROPS = "问题4"

const SHOPPING_MATCHED = 1
const SHOPPING_IRRELEVANT = 2
const SHOPPING_ROOM = "shop"

const UNMATCH_SHOPPING_HIT = `我不太清楚您的问题，我是一个购物小助手，请问我购物相关的问题吧`

func NewDashScopeModel() (*openai.LLM, error) {
	opts := make([]openai.Option, 0)
	opts = append(opts, openai.WithBaseURL(conf.GlobalConfig.DashScopeLLM.OpenaiUrl))
	opts = append(opts, openai.WithModel(conf.GlobalConfig.DashScopeLLM.Model))
	opts = append(opts, openai.WithToken(conf.GlobalConfig.DashScopeLLM.Key))
	opts = append(opts, openai.WithCallback(ShoppingFlowLogHandler{}))
	llm, err := openai.New(opts...)
	return llm, err
}

func (engine *ShoppingEngine) conversationContext(curUser user.User, room string) string {
	chatHistorys := llmutil.GetHistoryStoreInstance().LoadChatHistoryForLLM(curUser.UserId, room)
	var strBuilder strings.Builder
	for i, msg := range chatHistorys {
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

func (engine *ShoppingEngine) saveChatHistory(curUser user.User, room, query, response string) {
	llmutil.GetHistoryStoreInstance().AddChatHistory(curUser.UserId, room, query, response)
}

func (engine *ShoppingEngine) Flow(curUser user.User, room, query string) (string, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()

	if err != nil {
		logger.Errorf("create llm client err:%s", err.Error())
		return "", err
	}

	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.QueryRoute,
		[]string{"context", "question"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["question"] = query
	inputs["context"] = engine.conversationContext(curUser, room)

	result, err := llmChain.Call(ctx, inputs)
	if err != nil {
		//todo:重试
		logger.Error("call llm err:", err.Error())
		return "", err
	}

	text, ok := result["text"]

	if !ok {

		buf, _ := json.Marshal(result)
		logger.Errorf("llm response:%s", string(buf))
		return "", errors.New("llm response err")
	}
	content, ok := text.(string)
	if !ok {
		kindStr := reflect.TypeOf(text).Kind().String()
		logger.Errorf("text type:%s", kindStr)
		return "", errors.New("err type:" + kindStr)
	}
	if logger.IsLevelEnabled(logger.DebugLevel) {
		logger.Debugf("%s llm response:%s", query, content)
	}

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 {
		return "", errors.New("invalid json:" + content)
	}
	jsonContent := content[start : end+1]
	intentMap := make(map[string]any)
	err = json.Unmarshal([]byte(jsonContent), &intentMap)
	if err != nil {
		//todo:重试
		logger.Errorf("%s unmarshal err:%s", jsonContent, err.Error())
		return "", err
	}
	shoppingIntent := parseIntent(intentMap)
	if !shoppingIntent.IsShopping {
		return UNMATCH_SHOPPING_HIT,nil
	}

	return "",nil

}

//9解析购物意图
func parseIntent(intentMap map[string]any) ShoppingIntent {
	var shoppingIntent ShoppingIntent
	shoppingIntent.ProductProps = make(map[string]string)
	//{"完整信息":"","问题1":1,"问题2":""，"问题3":"","问题4":{"":""}}
	isShopping, ok := intentMap[KEY_IS_SHOPPING]
	for ok {
		switch v := isShopping.(type) {
		case int:
			if v != SHOPPING_MATCHED {
				logger.Infof("%d is not shopping question", v)
				shoppingIntent.IsShopping = false
				break
			}
		case string:
			intv, err := strconv.Atoi(v)
			if err != nil {
				logger.Infof("is not number {%s} ", v)
				shoppingIntent.IsShopping = false
				break
			}
			if intv != SHOPPING_MATCHED {
				logger.Infof("{%d} is not shopping question", intv)
				shoppingIntent.IsShopping = false
				break
			}
		default:
			logger.Infof("type err ")
			shoppingIntent.IsShopping = false
			break
		}
		shoppingIntent.IsShopping = true
		break
	}
	if !ok {
		logger.Infof("not exists key: KEY_IS_SHOPPING ")
		shoppingIntent.IsShopping = false
	}

	indepQuestion, ok := intentMap[KEY_INDEPENDENT]
	if ok {
		indepQuestionStr, ok := indepQuestion.(string)
		if !ok {
			logger.Errorf("err json format:%v", indepQuestion)
		}
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Debugf("[%s]:%s", KEY_INDEPENDENT, indepQuestionStr)
		}
		shoppingIntent.IndependentQuery = indepQuestionStr
	} else {
		logger.Infof("not exists key: KEY_INDEPENDENT ")
	}

	productName, ok := intentMap[KEY_PRODUCT]
	if ok {
		productNameStr, ok := productName.(string)
		if !ok {
			logger.Errorf("err json format:%v", productName)
		}
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Debugf("[%s]:%s", KEY_PRODUCT, productNameStr)
		}
		shoppingIntent.ProductName = productNameStr
	} else {
		logger.Infof("not exists key: KEY_PRODUCT ")
	}

	category, ok := intentMap[KEY_CATEGORY]
	if ok {
		categoryStr, ok := category.(string)
		if !ok {
			logger.Errorf("err json format:%v", category)
		}
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Debugf("[%s]:%s", KEY_CATEGORY, categoryStr)
		}
		shoppingIntent.Category = categoryStr
	} else {
		logger.Infof("not exists key: KEY_CATEGORY ")
	}

	productProps, ok := intentMap[KEY_PROPS]

	if ok {
		propMap, ok := productProps.(map[string]any)
		if !ok {
			logger.Errorf("err json format:%v", productProps)
		}
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Debugf("[%s]:%v", KEY_PROPS, propMap)
		}

		for k, v := range propMap {
			switch vv := v.(type) {
			case string:
				shoppingIntent.ProductProps[k] = vv
			case int:
				shoppingIntent.ProductProps[k] = strconv.Itoa(vv)
			case int32:
				shoppingIntent.ProductProps[k] = strconv.FormatInt(int64(vv), 10)
			case int64:
				shoppingIntent.ProductProps[k] = strconv.FormatInt(vv, 10)
			case float32:
				shoppingIntent.ProductProps[k] = fmt.Sprintf("%f", vv)
			case float64:
				shoppingIntent.ProductProps[k] = fmt.Sprintf("%f", vv)
			default:
				logger.Infof("invalid type %s:%v", k, v)
			}
		}
	} else {
		logger.Infof("not exists key: KEY_CATEGORY ")
	}
	return shoppingIntent
}

type SkuSearchResponse struct {
	NumHists int                  `json:"num_hits"`
	Hits     []detail.SkuDocument `json:"hits"`
}

func search() {

}
