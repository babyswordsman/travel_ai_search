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

func RewriteQuery(model llm.GenModel, query string, chatHistorys []llms.ChatMessage) ([]string, error) {
	systemMsg := llms.SystemChatMessage{
		Content: conf.GlobalConfig.PromptTemplate.QueryRewritingPrompt,
	}
	userMsg := llms.HumanChatMessage{
		Content: query,
	}

	contentLength := 0
	contentLength += len(systemMsg.GetContent())
	contentLength += len(userMsg.GetContent())

	msgs := make([]llms.ChatMessage, 0, len(chatHistorys)+2)
	msgs = append(msgs, systemMsg)
	//todo: 暂时只接受最长1024的长度，给prompt留了1024，后续再改成限制总长度
	//需要留意聊天记录的顺序
	remain := 1024 - len(userMsg.GetContent())
	for i := len(chatHistorys) - 1; i >= 0; i-- {
		remain = remain - len(chatHistorys[i].GetContent())
		if remain > 0 {
			msgs = append(msgs, chatHistorys[i])
		} else {
			break
		}
	}
	msgs = append(msgs, userMsg)

	answer, tokens := model.GetChatRes(msgs, nil)
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
			queries := make([]string,0)
			for _,item := range rewritingQueries{
				if q,ok:=item.(string);ok{
					queries = append(queries, q)
				}
			}
			return queries, nil
		}

	}
	return []string{query}, fmt.Errorf("query rewriting response format fault")
}
