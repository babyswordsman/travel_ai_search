package walmart

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/es"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/modelclient"
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

type ShoppingIntent struct {
	IsShopping       bool
	Category         string
	ProductName      string
	ProductProps     map[string]string
	IndependentQuery string
}

type ShoppingFlowLogHandler struct {
	callbacks.LogHandler
}

const KEY_INDEPENDENT = "Complete information"
const KEY_IS_SHOPPING = "Question 1"
const KEY_CATEGORY = "Question 2"
const KEY_PRODUCT = "Question 3"
const KEY_PROPS = "Question 4"

const SHOPPING_MATCHED = 1
const SHOPPING_IRRELEVANT = 2
const SHOPPING_ROOM = "shop"
const SHOPPING_INDEX_NAME = "wal_sku"

const UNMATCH_SHOPPING_HIT = `我不太清楚您的问题，我是一个购物小助手，请问我购物相关的问题吧`
const EMPTY_SHOPPING_HIT = `暂时没有符合您要求的商品，可以想想还需要购买其他商品吗？`

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

// return:
// type  shop or msg
// body  product infor or text msg
// error
func (engine *ShoppingEngine) Flow(curUser user.User, room, query string) (string, any, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()
	logEntry := logger.WithField("uid", curUser.UserId)
	if err != nil {
		logEntry.Errorf("create llm client err:%s", err.Error())
		return llmutil.CHAT_TYPE_MSG, "", err
	}

	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.WalmartQueryRoute,
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

	intentMap, err := engine.doLLM(logEntry, llmChain, inputs, ctx)

	if err != nil {
		return llmutil.CHAT_TYPE_MSG, "", common.Errorf("request intent err", err)
	}

	shoppingIntent := engine.parseIntent(intentMap)
	if !shoppingIntent.IsShopping {
		return llmutil.CHAT_TYPE_MSG, UNMATCH_SHOPPING_HIT, nil
	}

	resp, err := engine.search(&shoppingIntent)
	if err != nil {
		logEntry.Error("search for shopping intent error:", err.Error())
		//TODO: 出默认列表
		return llmutil.CHAT_TYPE_MSG, "", nil
	}

	if len(resp.Hits) == 0 {
		logEntry.Error("search for shopping intent hits empty")
		//TODO:不返空，给几个其他的推荐
		return llmutil.CHAT_TYPE_MSG, EMPTY_SHOPPING_HIT, nil
	}

	//TODO:需要做相似性截断

	// if len(shoppingIntent.ProductName) > 0 && len(shoppingIntent.ProductProps) == 0 {
	// 	hitThreshold := 20
	// 	if resp.NumHits > hitThreshold {
	// 		//商品比较多，可以建议用户细化需求
	// 		//TODO:如果用户已经交互多次了，就不要再细化了，直接推荐
	// 		independentQuery := shoppingIntent.IndependentQuery
	// 		if len(shoppingIntent.IndependentQuery) == 0 {
	// 			independentQuery = query
	// 		}
	// 		if logger.IsLevelEnabled(logEntry.Level) {
	// 			logEntry.Debugf("prepare generate advice for user,independentQuery:%s", independentQuery)
	// 		}
	// 		advice, err := engine.askUser(independentQuery, curUser, resp.Hits)

	// 		if err != nil {
	// 			logEntry.Error("gen advice error:", err.Error())
	// 			return llmutil.CHAT_TYPE_MSG, "", err
	// 		}
	// 		engine.saveChatHistory(curUser, room, query, advice)
	// 		return llmutil.CHAT_TYPE_MSG, advice, nil
	// 	} else {
	// 		logEntry.Infof("hit sku less than %d", hitThreshold)
	// 	}
	// }

	//给用户推荐产品
	independentQuery := shoppingIntent.IndependentQuery
	if len(shoppingIntent.IndependentQuery) == 0 {
		//todo:全部上下文
		independentQuery = query
	}
	respSku, err := engine.recommend(independentQuery, curUser, resp.Hits)

	return llmutil.CHAT_TYPE_SHOPPING, respSku, err

}

func (engine *ShoppingEngine) doLLM(logEntry *logger.Entry, llmChain *chains.LLMChain,
	inputs map[string]any, ctx context.Context) (map[string]any, error) {

	jsonContent, err := engine.doLLMRespStr(logEntry, llmChain, inputs, ctx)
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

func (engine *ShoppingEngine) doLLMRespList(logEntry *logger.Entry, llmChain *chains.LLMChain,
	inputs map[string]any, ctx context.Context) ([]map[string]any, error) {

	jsonContent, err := engine.doLLMRespStr(logEntry, llmChain, inputs, ctx)
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

func (engine *ShoppingEngine) doLLMRespStr(logEntry *logger.Entry, llmChain *chains.LLMChain,
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

func (engine *ShoppingEngine) askUser(indepQuestion string, curUser user.User, relDocs []*detail.WalmartSkuResp) (string, error) {
	//追问用户提供一些商品属性相关的喜好,使用查到的商品的属性名做为模板
	logEntry := logger.WithField("uid", curUser.UserId)
	cates := make(map[string]string)

	for _, doc := range relDocs {
		cates[doc.CategoryPath] = ""
	}

	var cateBuild strings.Builder

	for k := range cates {
		cateBuild.WriteString(k)
		cateBuild.WriteString(";")
	}

	var propsBuild strings.Builder

	var conversationContext strings.Builder
	if len(indepQuestion) > 0 {
		conversationContext.WriteString("user question：")
		conversationContext.WriteString(indepQuestion)
		conversationContext.WriteString("\r\n")
	}

	if cateBuild.Len() > 0 {
		conversationContext.WriteString("guessed category：")
		conversationContext.WriteString(cateBuild.String())
		conversationContext.WriteString("\r\n")
	}

	llm, err := NewDashScopeModel()
	if err != nil {
		logEntry.Error("new llm err:", err.Error())
		return "", nil
	}
	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.WalmartAdditionalInfo,
		[]string{"context", "props"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["context"] = conversationContext.String()
	inputs["props"] = propsBuild.String()
	ctx := context.Background()
	adviceMap, err := engine.doLLM(logEntry, llmChain, inputs, ctx)

	if err != nil {
		logEntry.Error("request llm err:", err.Error())
		return "", nil
	}

	existProps, ok := adviceMap["props"]
	if ok {
		logEntry.Infof("exist props:%v", existProps)
	}

	advice, ok := adviceMap["advice"]

	if ok {
		switch v := advice.(type) {
		case string:
			return v, nil
		default:
			return "", fmt.Errorf("llm response advice err type(%s),v:%v", reflect.TypeOf(advice).Kind().String(), v)
		}
	}

	return "", fmt.Errorf("field advice is not exists :%v", adviceMap)

}

func (engine *ShoppingEngine) recommend(indepQuestion string, curUser user.User, relDocs []*detail.WalmartSkuResp) ([]*detail.RecommendWalmartSkuResponse, error) {
	//追问用户提供一些商品属性相关的喜好,使用查到的商品的属性名做为模板
	logEntry := logger.WithField("uid", curUser.UserId)
	idMap := make(map[string]*detail.WalmartSkuResp)
	var skuBuf strings.Builder
	for i, doc := range relDocs {
		skuBuf.WriteString("SKU ID:")
		skuBuf.WriteString(doc.Id)
		skuBuf.WriteString(",product name:")
		skuBuf.WriteString(doc.Name)
		//skuBuf.WriteString(",description:")
		//skuBuf.WriteString(doc.ShortDescription)
		skuBuf.WriteString("\r\n")
		idMap[doc.Id] = relDocs[i]
	}
	logEntry.Info(skuBuf.String())

	llm, err := NewDashScopeModel()
	if err != nil {
		logEntry.Error("new llm err:", err.Error())
		return nil, err
	}
	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.WalmartSkuRecommend,
		[]string{"context", "sku_list"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["context"] = indepQuestion
	inputs["sku_list"] = skuBuf.String()
	ctx := context.Background()
	skuScoreMap, err := engine.doLLMRespList(logEntry, llmChain, inputs, ctx)

	if err != nil {
		logEntry.Error("request llm err:", err.Error())
		return nil, err
	}
	recommendSkuList := make([]detail.RecommendSku, 0, len(skuScoreMap))
	for _, skuMap := range skuScoreMap {
		id := ""
		var score float64 = 0.0
		reason := ""
		idObj, ok := skuMap["id"]
		if ok {
			switch v := idObj.(type) {
			case float64:
				id = strconv.Itoa(int(v))
			case string:
				id = v
			default:
				logEntry.Errorf("pasre llm response score err:%v", v)
			}
		}
		scoreObj, ok := skuMap["score"]
		if ok {
			switch v := scoreObj.(type) {
			case float64:
				score = v
			case string:
				floatV, err := strconv.ParseFloat(v, 64)
				if err != nil {
					logEntry.Errorf("pasre llm response score err:%s", v)
				} else {
					score = floatV
				}
			default:
				logEntry.Errorf("pasre llm response score err:%v", v)
			}
		}
		reasonObj, ok := skuMap["reason"]
		if ok {
			reason = reasonObj.(string)
		}

		sku := detail.RecommendSku{
			Id:     id,
			Score:  score,
			Reason: reason,
		}
		if _, ok := idMap[sku.Id]; ok {
			recommendSkuList = append(recommendSkuList, sku)
		} else {
			logEntry.Errorf("llm generate id err:%s", sku.Id)
		}

	}

	sort.Slice(recommendSkuList, func(i int, j int) bool {
		return recommendSkuList[i].Score > recommendSkuList[j].Score
	})

	resultList := make([]*detail.RecommendWalmartSkuResponse, 0)
	for _, sku := range recommendSkuList {
		var resp detail.RecommendWalmartSkuResponse
		resp.ProductId = sku.Id
		resp.Score = sku.Score
		resp.Reason = sku.Reason
		resp.ProductName = idMap[sku.Id].Name
		resp.ProductMainPic = idMap[sku.Id].MediumImage
		resp.ProductPrice = idMap[sku.Id].SalePrice
		resp.Aisle = idMap[sku.Id].Aisle
		resultList = append(resultList, &resp)
	}

	//buf, err := json.Marshal(resultList)
	return resultList, err
}

// 9解析购物意图
func (engine *ShoppingEngine) parseIntent(intentMap map[string]any) ShoppingIntent {
	var shoppingIntent ShoppingIntent
	shoppingIntent.ProductProps = make(map[string]string)
	//{"完整信息":"","问题1":1,"问题2":""，"问题3":"","问题4":{"":""}}
	isShopping, ok := intentMap[KEY_IS_SHOPPING]

	isShoppingInt := 0
	if ok {
		switch v := isShopping.(type) {
		case float64:
			isShoppingInt = int(v)
		case string:
			intv, err := strconv.Atoi(v)
			if err != nil {
				logger.Infof("key:%s ,v:%s is not number ", KEY_IS_SHOPPING, v)
			} else {
				isShoppingInt = intv
			}

		default:
			logger.Infof("key:%s ,v:%v type err ", KEY_IS_SHOPPING, isShopping)
		}
	} else {
		logger.Infof("not exists key: KEY_IS_SHOPPING ")
	}

	if logger.IsLevelEnabled(logger.DebugLevel) {
		logger.Debugf("[%s]:%d", KEY_IS_SHOPPING, isShoppingInt)
	}
	if isShoppingInt == SHOPPING_MATCHED {
		shoppingIntent.IsShopping = true
	} else if isShoppingInt == SHOPPING_IRRELEVANT {
		shoppingIntent.IsShopping = false
	} else {
		logger.Errorf("llm response [%s]:%d err", KEY_IS_SHOPPING, isShoppingInt)
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
	NumHits int                      `json:"num_hits"`
	Hits    []*detail.WalmartSkuResp `json:"hits"`
}

func (engine *ShoppingEngine) search(intent *ShoppingIntent) (*SkuSearchResponse, error) {
	matchs := make([]map[string]any, 0)
	if len(intent.ProductName) > 0 {
		matchs = append(matchs, map[string]any{"match": map[string]any{
			"Name": map[string]any{"query": intent.ProductName, "boost": 1},
		}})
		matchs = append(matchs, map[string]any{"match": map[string]any{
			"ShortDescription": map[string]any{"query": intent.ProductName,
				"boost": 0.1},
		}})

		matchs = append(matchs, map[string]any{"match": map[string]any{
			"LongDescription": map[string]any{"query": intent.ProductName,
				"boost": 0.1},
		}})
	}
	if len(intent.Category) > 0 {
		matchs = append(matchs, map[string]any{"match": map[string]any{
			"CategoryPath": map[string]any{"query": intent.Category,
				"boost": 0.05},
		}})

	}
	independentQuery := strings.TrimSpace(intent.IndependentQuery)
	if len(independentQuery) == 0 {
		independentQuery = intent.ProductName
	}

	embs, err := modelclient.GetInstance().QueryEmbedding([]string{independentQuery})
	if err != nil {
		return nil, fmt.Errorf("get query emb err")
	}

	//TODO:使用属性进行匹配,目前ES索引没有使用嵌套或者object结构，不能进行嵌套查询

	if len(matchs) == 0 {
		return nil, fmt.Errorf("search with intent,query is empty")
	}
	//TODO:使用属性进行匹配,目前ES索引没有使用嵌套或者object结构，不能进行嵌套查询

	query := map[string]any{
		"size":   4,
		"fields": []string{"ItemId", "Timestamp", "Aisle", "ParentItemId", "Color", "MediumImage", "Name", "BrandName", "CategoryPath", "SalePrice", "ShortDescription", "LongDescription"},
		"query": map[string]any{
			"bool": map[string]any{
				"should": matchs,
			},
		},
		"knn": map[string]any{
			"field":          "DescVector",
			"k":              2,
			"num_candidates": 20,
			"boost":          10,
			"query_vector":   embs[0],
		},
	}
	r, err := es.GetInstance().SearchIndex(SHOPPING_INDEX_NAME, query)
	if err != nil {
		return nil, err
	}
	// Print the response status, number of results, and request duration.
	hits := int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	took := int(r["took"].(float64))
	logger.Info("hits:", hits, ",took:", took)
	// Print the ID and document source for each hit.
	docs := make([]*detail.WalmartSkuResp, 0)
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		if logger.IsLevelEnabled(logger.DebugLevel) {
			for k, v := range hit.(map[string]interface{}) {
				logger.Debugf("k:%s,v:%v", k, v)
			}
		}

		id := hit.(map[string]interface{})["_id"]
		source := hit.(map[string]interface{})["_source"]
		score := hit.(map[string]interface{})["_score"]
		skuDoc := detail.EsToWalmartSku(source.(map[string]interface{}))

		skuDoc.Id = id.(string)
		skuDoc.Score = score.(float64)
		docs = append(docs, &skuDoc)
	}
	resp := &SkuSearchResponse{
		NumHits: hits,
		Hits:    docs,
	}
	return resp, nil
}
