package walmart

import (
	"context"
	"encoding/json"
	"strconv"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/es"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/shopping/detail"
	"travel_ai_search/search/user"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/prompts"
)

type WalmartProductDetailTask struct {
	status         int //1表示成功，其他表示识别
	productDetails []*detail.WalmartSkuResp
	output         string
}

func (task *WalmartProductDetailTask) Run(curUser user.User, room, input string) (llmutil.TaskOutputType, any, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()
	logEntry := logger.WithField("uid", curUser.UserId)
	if err != nil {
		logEntry.Errorf("create llm client err:%s", err.Error())
		return llmutil.CHAT_TYPE_MSG, "", err
	}

	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.WalmartExtractProductId,
		[]string{"context", "question"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["question"] = input
	inputs["context"] = conversationContext(curUser, room)
	if logger.IsLevelEnabled(logger.DebugLevel) {
		logEntry.Debugf("WalmartProductDetailTask inputs:%v", inputs)
	}
	idsMap, err := doLLM(logEntry, llmChain, inputs, ctx)
	if err != nil {
		logEntry.Errorf("do llm err:%s", err.Error())
		task.status = llmutil.RUN_ERR
		return llmutil.CHAT_TYPE_MSG, nil, err
	}

	ids := task.parseProductId(idsMap)
	skuResp, err := task.search(ids)
	if err != nil {
		logEntry.Errorf("query ids[%v] err:%s", ids, err.Error())
		task.status = llmutil.RUN_ERR
		return llmutil.CHAT_TYPE_MSG, nil, err
	}

	task.productDetails = skuResp.Hits

	value, err := task.chat(curUser, room, task.FormatDetail())
	if err != nil {
		logEntry.Errorf("chat ids[%v] err:%s", ids, err.Error())
		task.status = llmutil.RUN_ERR
		return llmutil.CHAT_TYPE_MSG, nil, err
	}

	task.output = value
	task.status = llmutil.RUN_DONE
	return llmutil.CHAT_TYPE_MSG, value, nil
}

func (task *WalmartProductDetailTask) Call(curUser user.User, room string, ids []string) (llmutil.TaskOutputType, any, error) {

	logEntry := logger.WithField("uid", curUser.UserId)

	skuResp, err := task.search(ids)
	if err != nil {
		logEntry.Errorf("query ids[%v] err:%s", ids, err.Error())
		task.status = llmutil.RUN_ERR
		return llmutil.CHAT_TYPE_MSG, nil, err
	}

	task.productDetails = skuResp.Hits

	value, err := task.chat(curUser, room, task.FormatDetail())
	if err != nil {
		logEntry.Errorf("chat ids[%v] err:%s", ids, err.Error())
		task.status = llmutil.RUN_ERR
		return llmutil.CHAT_TYPE_MSG, nil, err
	}

	task.output = value
	task.status = llmutil.RUN_DONE
	return llmutil.CHAT_TYPE_MSG, value, nil
}

func (task *WalmartProductDetailTask) chat(curUser user.User, room, input string) (string, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()
	logEntry := logger.WithField("uid", curUser.UserId).WithField("room", room)

	if err != nil {
		logEntry.Errorf("create llm client err:%s", err.Error())
		return "", err
	}

	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.WalmartChat,
		[]string{"context", "question"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["question"] = input
	inputs["context"] = task.FormatOutput()

	str, err := doLLMRespText(logEntry, llmChain, inputs, ctx)

	return str, err
}

func (task *WalmartProductDetailTask) parseProductId(idsMap map[string]any) []string {
	//{"product_ids":[],"isContain":false}
	ids := make([]string, 0)
	if isContainV, ok := idsMap["isContain"]; ok {
		logger.Infof("contain task id:%t", isContainV)
	}
	if productIdsV, ok := idsMap["product id"]; ok {
		if productIds, ok := productIdsV.([]any); ok {
			for _, idV := range productIds {
				switch id := idV.(type) {
				case string:
					ids = append(ids, id)
				case float64:
					ids = append(ids, strconv.FormatFloat(id, 'f', -1, 64))
				default:
					logger.Errorf("err type %v", idV)
				}
			}
		}

	}
	return ids
}

func (engine *WalmartProductDetailTask) search(ids []string) (*SkuSearchResponse, error) {
	matchs := make([]map[string]any, 0)
	for _, id := range ids {
		matchs = append(matchs, map[string]any{
			"term": map[string]any{
				"ItemId": id,
			},
		})
	}
	query := map[string]any{
		"size":    4,
		"_source": []string{"ItemId", "Timestamp", "Aisle", "ParentItemId", "Color", "MediumImage", "Name", "BrandName", "CategoryPath", "SalePrice", "ShortDescription", "LongDescription"},
		"query": map[string]any{
			"bool": map[string]any{
				"should": matchs,
			},
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

func (task *WalmartProductDetailTask) Status() int {
	return task.status
}

func (task *WalmartProductDetailTask) FormatOutput() string {
	if task.Status() != llmutil.RUN_DONE {
		return ""
	}
	return task.output
}

func (task *WalmartProductDetailTask) FormatDetail() string {
	if len(task.output) == 0 {
		return ""
	}

	buf, err := json.Marshal(task.output)
	if err != nil {
		logger.Errorf("marshal err:%s", err.Error())
		return ""
	}
	return string(buf)
}
