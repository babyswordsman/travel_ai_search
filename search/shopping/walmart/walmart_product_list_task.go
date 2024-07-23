package walmart

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
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
	"github.com/tmc/langchaingo/prompts"
)

type WalmartProductListTask struct {
	output []*detail.RecommendWalmartSkuResponse
	status int
}

func (task *WalmartProductListTask) FormatOutput() string {
	var buf strings.Builder
	for i, sku := range task.output {
		if i > 0 {
			buf.WriteString("\r\n")
		}
		buf.WriteString(fmt.Sprintf("product id:%s, product name:%s, Aisle:%s, price:%.2f",
			sku.ProductId, sku.ProductName, sku.Aisle, sku.ProductPrice))
	}
	return buf.String()
	// if task.Status() != 1 {
	// 	return ""
	// }

	// buf, err := json.Marshal(task.output)
	// if err != nil {
	// 	logger.Errorf("marshal err:%s", err.Error())
	// 	return ""
	// }
	// return string(buf)
}

func (task *WalmartProductListTask) Status() int {
	return task.status
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

// return:
// type  shop or msg
// body  product infor or text msg
// error
func (task *WalmartProductListTask) Run(curUser user.User, room, query string) (llmutil.TaskOutputType, any, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()
	logEntry := logger.WithField("uid", curUser.UserId)
	if err != nil {
		logEntry.Errorf("create llm client err:%s", err.Error())
		task.status = llmutil.RUN_ERR
		return llmutil.CHAT_TYPE_MSG, "", err
	}

	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.WalmartShoppingIntent,
		[]string{"context", "question"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["question"] = query
	inputs["context"] = conversationContext(curUser, room)

	intentMap, err := doLLM(logEntry, llmChain, inputs, ctx)

	if err != nil {
		task.status = llmutil.RUN_ERR
		return llmutil.CHAT_TYPE_MSG, "", common.Errorf("request intent err", err)
	}

	shoppingIntent := task.parseIntent(intentMap)
	if !shoppingIntent.IsShopping {
		return llmutil.CHAT_TYPE_MSG, UNMATCH_SHOPPING_HIT, nil
	}

	resp, err := task.search(&shoppingIntent)
	if err != nil {
		logEntry.Error("search for shopping intent error:", err.Error())
		task.status = llmutil.RUN_ERR
		//TODO: 出默认列表
		return llmutil.CHAT_TYPE_MSG, "", nil
	}

	if len(resp.Hits) == 0 {
		logEntry.Error("search for shopping intent hits empty")
		task.status = llmutil.RUN_ERR
		//TODO:不返空，给几个其他的推荐
		return llmutil.CHAT_TYPE_MSG, EMPTY_SHOPPING_HIT, nil
	}

	//给用户推荐产品
	independentQuery := shoppingIntent.IndependentQuery
	if len(shoppingIntent.IndependentQuery) == 0 {
		//todo:全部上下文
		independentQuery = query
	}
	respSku, err := task.recommend(independentQuery, curUser, resp.Hits)
	if err == nil {
		task.output = respSku
		task.status = llmutil.RUN_DONE
	} else {
		task.status = llmutil.RUN_ERR
	}
	return llmutil.CHAT_TYPE_SHOPPING, respSku, err

}

func (task *WalmartProductListTask) PlanB(curUser user.User, room, query string) (llmutil.TaskOutputType, any, error) {

	if len(query) > 512 {
		query = query[:512]
	}

	logEntry := logger.WithField("uid", curUser.UserId)

	logEntry.Infof("plan b query:%s", query)
	shoppingIntent := ShoppingIntent{
		IsShopping:       true,
		Category:         "",
		ProductName:      query,
		ProductProps:     make(map[string]string),
		IndependentQuery: query,
	}

	resp, err := task.search(&shoppingIntent)
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

	resultList := make([]*detail.RecommendWalmartSkuResponse, 0)
	for _, sku := range resp.Hits {
		var resp detail.RecommendWalmartSkuResponse
		resp.ProductId = sku.Id
		resp.Score = sku.Score
		resp.Reason = sku.ShortDescription
		resp.ProductName = sku.Name
		resp.ProductMainPic = sku.MediumImage
		resp.ProductPrice = sku.SalePrice
		resp.Aisle = sku.Aisle
		resultList = append(resultList, &resp)
	}

	return llmutil.CHAT_TYPE_SHOPPING, resultList, err
}

func (task *WalmartProductListTask) askUser(indepQuestion string, curUser user.User, relDocs []*detail.WalmartSkuResp) (string, error) {
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
	adviceMap, err := doLLM(logEntry, llmChain, inputs, ctx)

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

func (task *WalmartProductListTask) recommend(indepQuestion string, curUser user.User, relDocs []*detail.WalmartSkuResp) ([]*detail.RecommendWalmartSkuResponse, error) {
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
	skuScoreMap, err := doLLMRespList(logEntry, llmChain, inputs, ctx)

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
func (task *WalmartProductListTask) parseIntent(intentMap map[string]any) ShoppingIntent {
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

func (task *WalmartProductListTask) search(intent *ShoppingIntent) (*SkuSearchResponse, error) {
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
		"size":    4,
		"_source": []string{"ItemId", "Timestamp", "Aisle", "ParentItemId", "Color", "MediumImage", "Name", "BrandName", "CategoryPath", "SalePrice", "ShortDescription", "LongDescription"},
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
