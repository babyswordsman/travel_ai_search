package rewrite

import (
	"encoding/json"
	"fmt"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/llm"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

type QueryRewritingEngine interface {
	/**
	根据原始问题，生成对应的改写问题，方便检索引擎搜索，当改写过程出错，返回原问题
	*/
	Rewrite(query string, chatHistorys []llms.ChatMessage) ([]string, error)
}

type LLMQueryRewritingEngine struct {
	Model llm.GenModel
}

func (engine *LLMQueryRewritingEngine) Rewrite(query string, chatHistorys []llms.ChatMessage) ([]string, error) {
	msgs := llm.CombineLLMInputWithHistory(conf.GlobalConfig.PromptTemplate.QueryRewritingPrompt, query, chatHistorys, 1024)
	answer, tokens := engine.Model.GetChatRes(msgs, nil)
	if tokens == 0 {
		logger.Errorf("query:%s, tokens:%d,rewriting:%s", query, tokens, answer)
	} else {
		logger.Infof("query:%s, tokens:%d,rewriting:%s", query, tokens, answer)
	}

	if tokens == 0 {
		//出错了就原样返回
		return []string{query}, fmt.Errorf("query rewriting err:%s", answer)
	}

	result := make(map[string]any)
	//answer = strings.ReplaceAll(answer, "\\\"", "\"")
	err := json.Unmarshal([]byte(answer), &result)
	if err != nil {
		//出错了就原样返回
		return []string{query}, common.Errorf("query rewriting err:%w", err)
	}
	value, ok := result["rewriting_query"]

	if ok {
		rewritingQueries, ok := value.([]any)
		if ok {
			queries := make([]string, 0)
			for _, item := range rewritingQueries {
				if q, ok := item.(string); ok {
					queries = append(queries, q)
				}
			}
			return queries, nil
		}

	}
	return []string{query}, fmt.Errorf("query rewriting response format fault")
}
