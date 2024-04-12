package search

import (
	"encoding/json"
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
		// if len(details) == 0 {
		// 	return conf.EmptyHint, 0
		// }
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
	for i := len(chatHistorys) - 1; i >= 0; i-- {
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
	answer := `亲爱的游客，以下是我们为您精心策划的四川旅游行程，全程10天，
	带您领略四川的自然风光和人文风情。\n\n第一天：成都\n抵达成都后，您可以自由活动，
	逛逛锦里古街，品尝美食，感受成都的悠闲生活。夜宿成都。\n\n第二天：成都-都江堰-青城山\n早餐后，
	前往都江堰，游览都江堰水利工程。午餐后，前往青城山，游览道教名山青城山。夜宿都江堰。\n\n第三天：都江堰-峨眉山\n早餐后，
	乘车前往峨眉山，游览峨眉山金顶、清音阁等景点。夜宿峨眉山。\n\n第四天：峨眉山-乐山大佛\n早餐后，前往乐山，
	游览世界最大的石刻佛像乐山大佛。午餐后，游览乐山市区。夜宿乐山。\n\n第五天：乐山-泸沽湖\n早餐后，
	乘车前往泸沽湖，游览美丽的泸沽湖，体验摩梭文化。夜宿泸沽湖。\n\n第六天：泸沽湖-康定\n早餐后，乘车前往康定，游览康定情歌广场，
	体验浓郁的藏族风情。夜宿康定。\n\n第七天：康定-海螺沟\n早餐后，乘车前往海螺沟，游览冰川森林公园，
	欣赏壮观的冰川景观。夜宿海螺沟。\n\n第八天：海螺沟-稻城亚丁\n早餐后，乘车前往稻城亚丁，游览稻城亚丁景区，欣赏美丽的雪山、
	湖泊和草原。夜宿稻城。\n\n第九天：稻城-新龙\n早餐后，乘车前往新龙，游览新龙红草地、毛垭大草原，感受高原牧场的美丽。
	夜宿新龙。\n\n第十天：新龙-成都\n早餐后，乘车返回成都，结束美好的四川之旅。\n\n注意事项：\n1. 四川地区海拔较高，
	请注意防晒、保暖。\n2. 请尊重当地民俗风情，不要随意拍照。\n3. 自驾游请确保驾驶技术熟练，注意安全。\n
	4. 保持环境卫生，不要乱丢垃圾。\n5. 请随身携带身份证、现金、银行卡等重要物品。\n6. 请遵守景区规定，不要攀爬危险区域。\n
	7. 请按照导游安排的行程进行，不要私自离团。\n\n祝您在四川度过一个愉快的旅程！`
	contents := strings.Split(answer, "，")
	for _, txt := range contents {
		txt = strings.ReplaceAll(txt, "\\r\\n", "<br />")
		txt = strings.ReplaceAll(txt, "\\n", "<br />")
		msgResp := llm.ChatStream{
			Type:  llm.CHAT_TYPE_MSG,
			Body:  txt + "\n",
			Seqno: seqno,
		}
		v, _ := json.Marshal(msgResp)
		msgListener <- string(v)
	}

	return "sssss", 10
}

func LLMChatStream(room string, query string, msgListener chan string, chatHistorys []schema.ChatMessage) (answer string, totalTokens int64) {
	var prompt string
	logger.Infof("room:%s,query:%s", room, query)
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
	for i := len(chatHistorys) - 1; i >= 0; i-- {
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
