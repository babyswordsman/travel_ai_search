package modelclient

import (
	"testing"
	"travel_ai_search/search/conf"

	yaml "gopkg.in/yaml.v3"
)

var content = `embedding_model_host: http://127.0.0.1:8080
query_embedding_path: /embedding/query
passage_embedding_path: /embedding/passage
reranker_model_host: http://127.0.0.1:8080
predictor_reranker_path: /reranker/predict
`

func TestModelService(t *testing.T) {

	config := &conf.Config{}
	err := yaml.Unmarshal([]byte(content), config)
	if err != nil {
		t.Errorf("unmarshal err,%s", err)
	}

	client := InitModelClient(config)
	defer client.Close()

	embeds, err := client.QueryEmbedding([]string{"今天你高兴吗", "发生什么事件了"})
	if err != nil {
		t.Errorf("request model service err,%s", err)
	}
	if len(embeds) != 2 {
		t.Errorf("result err,length:%d", len(embeds))
	}

	if len(embeds[0]) != 768 {
		t.Errorf("result err,length:%d", len(embeds[0]))
	}

	embeds, err = client.PassageEmbedding([]string{"今天真高兴", "事件"})
	if err != nil {
		t.Errorf("request model service err,%s", err)
	}
	if len(embeds) != 2 {
		t.Errorf("result err,length:%d", len(embeds))
	}

	if len(embeds[0]) != 768 {
		t.Errorf("result err,length:%d", len(embeds[0]))
	}

	query_passage := [][2]string{
		{"广西哪里好玩", "广西哪里好玩"},
		{"广西哪里好玩", "广西不好玩"},
		{"广西哪里好玩", "广西的桂林很好玩，还有漓江也很好玩"},
		{"广西哪里好玩", "桂林山水"},
		{"广西哪里好玩", "阳朔"},
		{"广西哪里好玩", "上海外滩很好玩"},
		{"广西哪里好玩", "上海外滩"},
		{"广西哪里好玩", "去泰山旅游"},
	}

	scores, err := client.PredictorRerankerScore(query_passage)
	if err != nil {
		t.Errorf("request model service err,%s", err)
	}
	if len(scores) != len(query_passage) {
		t.Errorf("scores len:%d, query len:%d", len(scores), len(query_passage))
	} else {
		t.Logf("scores:%v", scores)
	}
}
