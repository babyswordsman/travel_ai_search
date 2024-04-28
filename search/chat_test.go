package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/llm"
	"travel_ai_search/search/llm/spark"
	"travel_ai_search/search/modelclient"
	searchengineapi "travel_ai_search/search/search_engine_api"

	yaml "gopkg.in/yaml.v3"
)

func TestLLMChatPrompt(t *testing.T) {

	var content = `embedding_model_host: http://127.0.0.1:8080
query_embedding_path: /embedding/query
passage_embedding_path: /embedding/passage
reranker_model_host: http://127.0.0.1:8080
predictor_reranker_path: /reranker/predict
preranking_threshold: 0.4
max_candidates: 5
`
	config := &conf.Config{}
	err := yaml.Unmarshal([]byte(content), config)
	if err != nil {
		t.Errorf("unmarshal err,%s", err)
	}
	conf.GlobalConfig = config

	client := modelclient.InitModelClient(config)
	defer client.Close()

	info, err := os.Stat(search_result_file)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatal("file not exist")
			searchEngine := &searchengineapi.OpenSerpSearchEngine{
				Engines: []string{"google"},
				BaseUrl: "http://llm-search.com/openserp/search",
			}
			searchItems, err := searchEngine.Search(context.Background(), config, "规划一个去四川旅游的行程")
			if err != nil {
				t.Error("e:", err)
				return
			}

			buf, _ := json.Marshal(searchItems)
			writeJson(buf)
		}
	}
	t.Logf("info:%s", info.ModTime().GoString())

	prompt := &llm.TravelPrompt{
		MaxLength:    1024,
		PromptPrefix: "您是一个智能助手",
	}
	model := &spark.SparkModel{Room: "chat"}

	engine := ChatEngine{
		SearchEngine: &MockSearchEngine{},
		Prompt:       prompt,
		Model:        model,
		Room:         "chat",
	}

	str, err := engine.LLMChatPrompt("规划一个去四川旅游的行程")
	if err != nil {
		t.Error("e:", err)
	}
	t.Log(str)
}

var search_result_file = "../data/search_google.json"

func writeJson(buf []byte) {

	file, err := os.Create(search_result_file)
	if err != nil {
		return
	}
	defer file.Close()
	file.Write(buf)
}

type MockSearchEngine struct {
}

func (engine *MockSearchEngine) Search(ctx context.Context, config *conf.Config, query string) ([]searchengineapi.SearchItem, error) {
	buf, err := os.ReadFile(search_result_file)
	if err != nil {
		return make([]searchengineapi.SearchItem, 0), fmt.Errorf("read file err :%s", search_result_file)
	}
	items := make([]searchengineapi.SearchItem, 0)
	err = json.Unmarshal(buf, &items)
	if err != nil {
		return make([]searchengineapi.SearchItem, 0), fmt.Errorf("read file err :%s", search_result_file)
	}
	return items, nil
}
