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

func Prompt(candidates []map[string]string) string {
	buf := strings.Builder{}
	buf.WriteString(conf.GlobalConfig.SparkLLM.Prompt)
	buf.WriteString("\r\n")
	for ind, detail := range candidates {
		buf.WriteString("方案" + strconv.Itoa(ind+1) + ":")
		buf.WriteString("\r\n")
		buf.WriteString(detail[conf.DETAIL_TITLE_FIELD])
		//buf.WriteString("\r\n")
		//buf.WriteString(detail[conf.DETAIL_CONTENT_FIELD])
		//buf.WriteString("\r\n")
		buf.WriteString("\r\n")
	}
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
	prompt := Prompt(details)
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
	prompt := Prompt(details)
	systemMsg := llm.Message{Role: llm.ROLE_SYSTEM, Content: prompt}
	queryMsg := llm.Message{Role: llm.ROLE_USER, Content: query}
	resp, totalTokens := spark.GetChatRes([]llm.Message{systemMsg, queryMsg}, nil)
	return resp, totalTokens

}

func LLMChatStreamMock(query string, msgListener chan string) (string, int64) {
	details, err := SearchCandidate(query, conf.GlobalConfig.PreRankingThreshold)
	if err != nil {
		return conf.ErrHint, 0
	}

	if len(details) == 0 {
		return conf.EmptyHint, 0
	}
	time.Sleep(time.Second * 2)

	prompt := Prompt(details)

	candidateResp := llm.ChatStream{
		Type: llm.CHAT_TYPE_CANDIDATE,
		Body: details,
	}
	v, _ := json.Marshal(candidateResp)
	msgListener <- string(v)

	msgResp := llm.ChatStream{
		Type:  llm.CHAT_TYPE_MSG,
		Body:  prompt,
		Seqno: "1",
	}
	v, _ = json.Marshal(msgResp)
	msgListener <- string(v)

	for i := 0; i < 10; i++ {
		msgResp := llm.ChatStream{
			Type:  llm.CHAT_TYPE_MSG,
			Body:  fmt.Sprintf("第%d天", i),
			Seqno: "2",
		}
		v, _ := json.Marshal(msgResp)
		msgListener <- string(v)
	}
	return "sssss", 10
}

func LLMChatStream(query string, msgListener chan string) (answer string, totalTokens int64) {
	details, err := SearchCandidate(query, conf.GlobalConfig.PreRankingThreshold)
	if err != nil {
		return conf.ErrHint, 0
	}

	if len(details) == 0 {
		return conf.EmptyHint, 0
	}
	prompt := Prompt(details)
	systemMsg := llm.Message{Role: llm.ROLE_SYSTEM, Content: prompt}
	queryMsg := llm.Message{Role: llm.ROLE_USER, Content: query}
	entry := logger.WithField("query", queryMsg).WithField("system", systemMsg)
	candidateResp := llm.ChatStream{
		Type: llm.CHAT_TYPE_CANDIDATE,
		Body: details,
	}
	v, _ := json.Marshal(candidateResp)
	msgListener <- string(v)

	answer, totalTokens = spark.GetChatRes([]llm.Message{systemMsg, queryMsg}, msgListener)
	entry.WithField("totalTokens", totalTokens).WithField("answer", answer).Info("[chat]")
	return
}
