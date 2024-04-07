package modelclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"travel_ai_search/search/conf"
)

type ModelClient struct {
	cli                  *http.Client
	queryEmbeddingPath   string
	passageEmbeddingPath string
	rerankerPath         string
}

var modelCli *ModelClient

func GetInstance() *ModelClient {
	return modelCli
}
func InitModelClient(config *conf.Config) *ModelClient {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 3,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     time.Minute * 10,
	}
	client := http.Client{
		Transport: transport,
		Timeout:   time.Minute * 15,
	}
	modelClient := &ModelClient{
		cli:                  &client,
		queryEmbeddingPath:   config.EmbeddingModelHost + config.QueryEmbeddingPath,
		passageEmbeddingPath: config.EmbeddingModelHost + config.PassageEmbeddingPath,
		rerankerPath:         config.RerankerModelHost + config.PredictorRerankerPath,
	}
	modelCli = modelClient
	return modelClient
}

func (modelClient *ModelClient) QueryEmbedding(queries []string) ([][]float32, error) {
	body := map[string][]string{
		"queries": queries,
	}
	embeddings := make([][]float32, 0)
	buf, err := json.Marshal(body)
	if err != nil {
		return embeddings, fmt.Errorf("json marshal err,[%v]", body)
	}
	resp, err := modelClient.cli.Post(modelClient.queryEmbeddingPath, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return embeddings, fmt.Errorf("[query embedding] err:%s", err)
	}
	defer resp.Body.Close()
	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return embeddings, fmt.Errorf("[query embedding] read response err:%s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return embeddings, fmt.Errorf("[query embedding] err,status:%d,body:%s", resp.StatusCode, respContent)
	}
	respStruct := make(map[string][][]float32)
	err = json.Unmarshal(respContent, &respStruct)
	if err != nil {
		return embeddings, fmt.Errorf("[query embedding] unmarshal{%s} response err:%s", respContent, err)
	}
	return respStruct["embs"], nil
}

func (modelClient *ModelClient) PassageEmbedding(passages []string) ([][]float32, error) {
	body := map[string][]string{
		"passages": passages,
	}
	embeddings := make([][]float32, 0)
	buf, err := json.Marshal(body)
	if err != nil {
		return embeddings, fmt.Errorf("json marshal err,[%v]", body)
	}
	resp, err := modelClient.cli.Post(modelClient.passageEmbeddingPath, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return embeddings, fmt.Errorf("[passage embedding] err:%s", err)
	}
	defer resp.Body.Close()
	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return embeddings, fmt.Errorf("[passage embedding] read response err:%s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return embeddings, fmt.Errorf("[passage embedding] err,status:%d,body:%s", resp.StatusCode, respContent)
	}

	respStruct := make(map[string][][]float32)
	err = json.Unmarshal(respContent, &respStruct)
	if err != nil {
		return embeddings, fmt.Errorf("[query embedding] unmarshal{%s} response err:%s", respContent, err)
	}
	return respStruct["embs"], nil
}

/*
*
精排打分，格式为query_passage：[["北京哪里好玩","推荐你去爬长城"],["query","answer"]]
@return 打分列表,按传入顺序排序
*/
func (modelClient *ModelClient) PredictorRerankerScore(query_passage [][2]string) ([]float32, error) {
	body := map[string][][2]string{
		"q_p_pairs": query_passage,
	}
	scores := make([]float32, 0)
	buf, err := json.Marshal(body)
	if err != nil {
		return scores, fmt.Errorf("json marshal err,[%v]", body)
	}
	resp, err := modelClient.cli.Post(modelClient.rerankerPath, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return scores, fmt.Errorf("[predictor reranker] err:%s", err)
	}
	defer resp.Body.Close()
	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return scores, fmt.Errorf("[predictor reranker] read response err:%s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return scores, fmt.Errorf("[predictor reranker] err,status:%d,body:%s", resp.StatusCode, respContent)
	}

	respStruct := make(map[string][]float32)
	err = json.Unmarshal(respContent, &respStruct)
	if err != nil {
		return scores, fmt.Errorf("[query embedding] unmarshal{%s} response err:%s", respContent, err)
	}
	return respStruct["scores"], nil
}

func (modelClient *ModelClient) Close() {
	//empty
}
