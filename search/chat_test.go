package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/llm"
	"travel_ai_search/search/llm/dashscope"
	"travel_ai_search/search/llm/spark"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/qdrant"
	"travel_ai_search/search/rewrite"
	searchengineapi "travel_ai_search/search/search_engine_api"

	"github.com/devinyf/dashscopego/qwen"
	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"

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

func TestChatStream(t *testing.T) {
	//todo:为测试创建Mock类
	logger.SetLevel(logger.DebugLevel)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config

	tmpKVClient, err := kvclient.InitClient(config)
	if err != nil {
		t.Errorf("init kv client:%s err:%s", config.RedisAddr, err)
		return
	}

	defer tmpKVClient.Close()

	kvclient.InitDetailIdGen()

	tmpVecClient, err := qdrant.InitVectorClient(config)
	if err != nil {
		t.Errorf("init vector client:%s err:%s", config.QdrantAddr, err)
		return
	}

	defer tmpVecClient.Close()

	tmpModelClient := modelclient.InitModelClient(config)
	defer tmpModelClient.Close()

	//llm.InitMemHistoryStoreInstance(5)
	llm.InitKVHistoryStoreInstance(kvclient.GetInstance(), 10)
	//用户历史清理
	llm.GetHistoryStoreInstance().StarCleanTask()

	tokens := int64(0)
	answer := ""

	var searchEngine searchengineapi.SearchEngine
	var prompt llm.Prompt
	var model llm.GenModel
	var rewritingEngine rewrite.QueryRewritingEngine

	//searchEngine = &searchengineapi.GoogleSearchEngine{}
	searchEngine = &searchengineapi.OpenSerpSearchEngine{
		Engines: conf.GlobalConfig.OpenSerpSearch.Engines,
		BaseUrl: conf.GlobalConfig.OpenSerpSearch.Url,
	}
	prompt = &llm.ChatPrompt{
		MaxLength:    1024,
		PromptPrefix: conf.GlobalConfig.PromptTemplate.ChatPrompt,
	}
	model = &dashscope.DashScopeModel{
		ModelName: qwen.QwenTurbo,
		Room:      "chat",
	}

	rewritingEngine = &rewrite.LLMQueryRewritingEngine{
		Model: &dashscope.DashScopeModel{
			ModelName: qwen.QwenTurbo,
			Room:      "chat",
		},
	}

	engine := &ChatEngine{
		SearchEngine:    searchEngine,
		RewritingEnging: rewritingEngine,
		Prompt:          prompt,
		Model:           model,
		Room:            "chat",
	}
	msgListener := make(chan string)
	go func() {
		for v := range msgListener {
			t.Log("receive:", v)
		}
	}()
	answer, tokens = engine.LLMChatStream("香港和纽约哪个房价高，请说明简短的理由", msgListener, make([]llms.ChatMessage, 0))

	if tokens == 0 {
		t.Error("answer:", answer, " tokens:", strconv.FormatInt(tokens, 10))
	}
	t.Log("answer:", answer)
}
