package search

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/llm"
	"travel_ai_search/search/llm/spark"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/qdrant"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/schema"
)

func SearchCandidate(query string, threshold float32) ([]map[string]string, error) {
	vectors, err := modelclient.GetInstance().QueryEmbedding([]string{query})
	if err != nil {
		logger.Errorf("query embedding err:%s", err)
		return nil, err
	}

	scores, err := qdrant.GetInstance().Search(qdrant.DETAIL_COLLECTION,
		vectors[0], uint64(conf.GlobalConfig.MaxCandidates), false, true)
	if err != nil {
		logger.Errorf("{%s},search err:%s", query, err)
		return nil, err
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].GetScore() < scores[j].GetScore()
	})

	keys := make([]string, 0, len(scores))
	for i := len(scores) - 1; i >= 0; i-- {
		scoreNode := scores[i]
		logger.WithField("query", query).Infof("score:%f,key:%s", scoreNode.GetScore(), scoreNode.GetPayload()["id"].GetStringValue())
		if threshold > scoreNode.GetScore() {
			continue
		}
		keys = append(keys, scoreNode.GetPayload()["id"].GetStringValue())
	}

	if len(keys) == 0 {
		return make([]map[string]string, 0), nil
	}

	details := make([]map[string]string, 0, len(keys))
	for _, key := range keys {
		detail, err := kvclient.GetInstance().HGetAll(key)
		if err != nil {
			logger.WithField("key", key).Error("fetch detail err", err)
			continue
		}
		details = append(details, detail)
	}

	return details, nil

}

func TravelPrompt(candidates []map[string]string) string {
	//todo:截断
	buf := strings.Builder{}
	buf.WriteString(conf.GlobalConfig.SparkLLM.TravelPrompt)
	buf.WriteString("\r\n")
	remain := 1024
	for ind, detail := range candidates {
		titleLen := len(detail[conf.DETAIL_TITLE_FIELD])
		contentLen := len(detail[conf.DETAIL_CONTENT_FIELD])
		if remain-titleLen > 0 {
			buf.WriteString("方案" + strconv.Itoa(ind+1) + ":")
			buf.WriteString("\r\n")
			buf.WriteString(detail[conf.DETAIL_TITLE_FIELD])
			remain = remain - titleLen
		} else {
			break
		}

		if remain-contentLen > 0 {
			buf.WriteString("\r\n")
			buf.WriteString(detail[conf.DETAIL_CONTENT_FIELD])
			buf.WriteString("\r\n")
			remain = remain - contentLen
		} else {
			break
		}

	}
	buf.WriteString("\r\n")
	logger.Info(buf.String())
	return buf.String()
}

func ChatPrompt(candidates []map[string]string) string {
	//todo:截断
	buf := strings.Builder{}
	buf.WriteString(conf.GlobalConfig.SparkLLM.TravelPrompt)
	buf.WriteString("\r\n")
	remain := 1024
	for ind, detail := range candidates {
		titleLen := len(detail[conf.DETAIL_TITLE_FIELD])
		contentLen := len(detail[conf.DETAIL_CONTENT_FIELD])
		if remain-titleLen > 0 {
			buf.WriteString("方案" + strconv.Itoa(ind+1) + ":")
			buf.WriteString("\r\n")
			buf.WriteString(detail[conf.DETAIL_TITLE_FIELD])
			remain = remain - titleLen
		} else {
			break
		}

		if remain-contentLen > 0 {
			buf.WriteString("\r\n")
			buf.WriteString(detail[conf.DETAIL_CONTENT_FIELD])
			buf.WriteString("\r\n")
			remain = remain - contentLen
		} else {
			break
		}

	}
	buf.WriteString("\r\n")
	logger.Info(buf.String())
	return buf.String()
}

func LLMChatPrompt(query string) string {
	details, err := SearchCandidate(query, conf.GlobalConfig.PreRankingThreshold)
	if err != nil {
		return conf.ErrHint
	}

	if len(details) == 0 {
		return conf.EmptyHint
	}
	prompt := TravelPrompt(details)
	return prompt
}

func LLMChat(query string) (string, int64) {
	details, err := SearchCandidate(query, conf.GlobalConfig.PreRankingThreshold)
	if err != nil {
		return conf.ErrHint, 0
	}

	if len(details) == 0 {
		return conf.EmptyHint, 0
	}
	prompt := TravelPrompt(details)
	systemMsg := schema.SystemChatMessage{
		Content: prompt,
	}
	queryMsg := schema.HumanChatMessage{
		Content: query,
	}
	resp, totalTokens := spark.GetChatRes([]schema.ChatMessage{systemMsg, queryMsg}, nil)
	return resp, totalTokens

}

func LLMChatStreamMock(room string, query string, msgListener chan string, chatHistorys []schema.ChatMessage) (string, int64) {
	var prompt string

	if room == "travel" {
		details, err := SearchCandidate(query, conf.GlobalConfig.PreRankingThreshold)
		if err != nil {
			return conf.ErrHint, 0
		}
		if len(details) == 0 {
			return conf.EmptyHint, 0
		}
		prompt = TravelPrompt(details)
		candidateResp := llm.ChatStream{
			Type: llm.CHAT_TYPE_CANDIDATE,
			Body: details,
		}
		v, _ := json.Marshal(candidateResp)
		msgListener <- string(v)
	} else {
		prompt = conf.GlobalConfig.SparkLLM.ChatPrompt
		//todo: search engine
	}

	//systemMsg := llm.Message{Role: llm.ROLE_SYSTEM, Content: prompt}
	//queryMsg := llm.Message{Role: llm.ROLE_USER, Content: query}
	systemMsg := schema.SystemChatMessage{
		Content: prompt,
	}
	userMsg := schema.HumanChatMessage{
		Content: query,
	}

	contentLength := 0
	contentLength += len(systemMsg.GetContent())
	contentLength += len(userMsg.GetContent())

	msgs := make([]schema.ChatMessage, 0, len(chatHistorys)+2)
	msgs = append(msgs, systemMsg)
	//todo: 暂时只接受最长1024的长度，给prompt留了1024，后续再改成限制总长度
	remain := 1024 - len(userMsg.GetContent())
	for i := len(chatHistorys) - 1; i <= 0; i-- {
		remain = remain - len(chatHistorys[i].GetContent())
		if remain > 0 {
			msgs = append(msgs, chatHistorys[i])
		} else {
			break
		}
	}
	msgs = append(msgs, userMsg)

	seqno := strconv.FormatInt(time.Now().UnixMilli(), 10)
	for _, msg := range msgs {
		msgResp := llm.ChatStream{
			Type:  llm.CHAT_TYPE_MSG,
			Body:  msg.GetContent(),
			Seqno: seqno,
		}
		v, _ := json.Marshal(msgResp)
		msgListener <- string(v)
	}
	for i := 0; i < 10; i++ {
		msgResp := llm.ChatStream{
			Type:  llm.CHAT_TYPE_MSG,
			Body:  fmt.Sprintf("第%d天", i),
			Seqno: seqno,
		}
		v, _ := json.Marshal(msgResp)
		msgListener <- string(v)
	}
	return "sssss", 10
}

func LLMChatStream(room string, query string, msgListener chan string, chatHistorys []schema.ChatMessage) (answer string, totalTokens int64) {
	var prompt string

	if room == "travel" {
		details, err := SearchCandidate(query, conf.GlobalConfig.PreRankingThreshold)
		if err != nil {
			return conf.ErrHint, 0
		}
		if len(details) == 0 {
			return conf.EmptyHint, 0
		}
		prompt = TravelPrompt(details)
		candidateResp := llm.ChatStream{
			Type: llm.CHAT_TYPE_CANDIDATE,
			Body: details,
		}
		v, _ := json.Marshal(candidateResp)
		msgListener <- string(v)
	} else {
		prompt = conf.GlobalConfig.SparkLLM.ChatPrompt
		//todo: search engine
	}
	systemMsg := schema.SystemChatMessage{
		Content: prompt,
	}
	userMsg := schema.HumanChatMessage{
		Content: query,
	}
	entry := logger.WithField("query", userMsg).WithField("system", systemMsg)

	contentLength := 0
	contentLength += len(systemMsg.GetContent())
	contentLength += len(userMsg.GetContent())

	msgs := make([]schema.ChatMessage, 0, len(chatHistorys)+2)
	msgs = append(msgs, systemMsg)
	//todo: 暂时只接受最长1024的长度，给prompt留了1024，后续再改成限制总长度
	remain := 1024 - len(userMsg.GetContent())
	for i := len(chatHistorys) - 1; i <= 0; i-- {
		remain = remain - len(chatHistorys[i].GetContent())
		if remain > 0 {
			msgs = append(msgs, chatHistorys[i])
		} else {
			break
		}
	}
	msgs = append(msgs, userMsg)
	answer, totalTokens = spark.GetChatRes(msgs, msgListener)
	entry.WithField("totalTokens", totalTokens).WithField("answer", answer).Info("[chat]")
	return
}
