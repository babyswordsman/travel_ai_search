package es

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	logger "github.com/sirupsen/logrus"
)

type ESClient struct {
	cli        *es8.Client
	server_url []string
}

var esCli *ESClient

func InitESClient(config *conf.Config) (*ESClient, error) {
	if len(config.ESUrl) == 0 {
		return nil, errors.New("config.ESUrl is empty")
	}
	cfg := es8.Config{
		Addresses: config.ESUrl,
	}
	client, err := es8.NewClient(cfg)
	if err != nil {
		return nil, common.Errorf("new es client err", err)
	}
	esClient := &ESClient{
		cli:        client,
		server_url: config.ESUrl,
	}
	esCli = esClient
	return esCli, nil

}

func GetInstance() *ESClient {
	return esCli
}

func (client *ESClient) GetIndex(index string) (string, error) {
	req := esapi.IndicesGetMappingRequest{
		Index: []string{index},
	}
	resp, err := req.Do(context.Background(), client.cli)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("status err :" + resp.Status())
	}
	defer resp.Body.Close()
	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respContent), nil

}

func (client *ESClient) CreateIndex(index, sku_index_mapping string) error {
	// 初始化 Elasticsearch 客户端

	buf := bytes.NewBufferString(sku_index_mapping)

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  buf,
	}
	resp, err := req.Do(context.Background(), client.cli)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.Errorf("create index["+index+"] read response err", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create index[%s] code:%s,msg:%s", index, resp.Status(), string(respContent))
	}

	return nil
}

func (client *ESClient) DeleteIndex(index string) (string, error) {
	// 初始化 Elasticsearch 客户端

	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}
	resp, err := req.Do(context.Background(), client.cli)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", common.Errorf("delete index["+index+"] read response err", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("delete index[%s] code:%s,msg:%s", index, resp.Status(), string(respContent))
	}

	return string(respContent), nil
}

func (client *ESClient) AddDocument(index, docId, doc string) (int64, error) {

	buf := bytes.NewBufferString(doc)
	req := esapi.IndexRequest{
		Index:      index,
		Body:       buf,
		DocumentID: docId,
	}

	resp, err := req.Do(context.Background(), client.cli)
	if err != nil {
		return 0, common.Errorf(fmt.Sprintf("[%s] add doc[%s] err", index, docId), err)
	}

	defer resp.Body.Close()
	respContent, err := io.ReadAll(resp.Body)

	if err != nil {
		return 0, common.Errorf(fmt.Sprintf("[%s] add doc[%s] err", index, docId), err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusIMUsed {
		return 0, fmt.Errorf("[%s] add doc [%s] status:%s ,msg:%s ", index, docId, resp.Status(), string(respContent))
	}
	//{"_index":"sku","_id":"2024-06-30T18:52:30","_version":1,"result":"created","_shards":{"total":2,"successful":1,"failed":0},"_seq_no":0,"_primary_term":1}
	resultMap := make(map[string]any)
	json.Unmarshal(respContent, &resultMap)
	shardsResp, ok := resultMap["_shards"]
	if ok {
		shardsMap, ok := shardsResp.(map[string]any)
		if ok {
			num, ok := shardsMap["successful"]
			if ok {
				switch v := num.(type) {
				case int:
					//fmt.Println("type:int")
					return int64(v), nil
				case int32:
					//fmt.Println("type:int32")
					return int64(v), nil
				case int64:
					//fmt.Println("type:int64")
					return int64(v), nil
				case float32:
					//fmt.Println("type:float32")
					return int64(v), nil
				case float64:
					//fmt.Println("type:float64")
					return int64(v), nil
				default:
					return 0, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("parse response err,msg:%s", string(respContent))
}

func (client *ESClient) SearchIndex(index string, query map[string]any) (map[string]any, error) {

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("[%s]  encoding query err: %s ", index, err.Error())
	}

	if logger.IsLevelEnabled(logger.DebugLevel) {
		logger.Debugf("es [%s] query:%s", index, buf.String())
	}
	// Perform the search request.
	res, err := client.cli.Search(
		client.cli.Search.WithContext(context.Background()),
		client.cli.Search.WithIndex(index),
		client.cli.Search.WithBody(&buf),
		client.cli.Search.WithTrackTotalHits(true),
		client.cli.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("[%s] getting response err: %s", index, err.Error())
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("[%s] http code:%s parse response err: %s", index, res.Status(), err.Error())
		} else {
			// Print the response status and error information.
			return nil, fmt.Errorf("[%s] status:[%s] %s: %s", index,
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}
	var r map[string]any
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("[%s] parse response err: %s", index, err.Error())
	}
	return r, nil
}

func (client *ESClient) Close() {

}
