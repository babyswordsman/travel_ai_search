package quickwit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"

	logger "github.com/sirupsen/logrus"
)

type QuickwitClient struct {
	cli        *http.Client
	server_url string
}

var quickwitCli *QuickwitClient

func InitQuickwitClient(config *conf.Config) *QuickwitClient {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 3,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     time.Minute * 10,
	}
	client := http.Client{
		Transport: transport,
		Timeout:   time.Minute * 15,
	}
	quickwitClient := &QuickwitClient{
		cli:        &client,
		server_url: config.QuickwitUrl,
	}
	quickwitCli = quickwitClient
	return quickwitClient
}

func GetInstance() *QuickwitClient {
	return quickwitCli
}

func (client *QuickwitClient) CreateIndex(indexConfig string) error {
	path, err := url.JoinPath(client.server_url, "/indexes")
	if err != nil {
		return common.Errorf(fmt.Sprintf("%s joinpath err", client.server_url), err)
	}
	resp, err := client.cli.Post(path, "application/yaml", bytes.NewBufferString(indexConfig))
	if err != nil {
		return common.Errorf(fmt.Sprintf("%s create index err", path), err)
	}
	defer resp.Body.Close()
	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.Errorf("read http response err", err)
	}
	if resp.StatusCode != http.StatusOK {
		return common.Errorf("", errors.New(fmt.Sprintf("http status:%d,msg:%s", resp.StatusCode, string(resultBytes))))
	}
	logger.Info("response:", string(resultBytes))
	return nil
}

func (client *QuickwitClient) GetIndexMeta(index string) (string, error) {

	path, err := url.JoinPath(client.server_url, "/indexes", index)
	if logger.IsLevelEnabled(logger.DebugLevel) {
		logger.Debug("path:", path)
	}

	if err != nil {
		return "", common.Errorf(fmt.Sprintf("%s joinpath err", client.server_url), err)
	}
	resp, err := client.cli.Get(path)
	if err != nil {
		return "", common.Errorf(fmt.Sprintf("[%s] get index metadata err", path), err)
	}
	defer resp.Body.Close()
	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", common.Errorf("read http response err", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", common.Errorf("", errors.New(fmt.Sprintf("http status:%d,msg:%s", resp.StatusCode, string(resultBytes))))
	}

	return string(resultBytes), nil
}

func (client *QuickwitClient) AddDocument(index string, docs []string) (int, error) {

	path, err := url.JoinPath(client.server_url, index, "ingest")
	if logger.IsLevelEnabled(logger.DebugLevel) {
		logger.Debug("path:", path)
	}

	if err != nil {
		return 0, common.Errorf(fmt.Sprintf("%s joinpath err", client.server_url), err)
	}
	var buf bytes.Buffer
	for i, doc := range docs {
		if i > 0 {
			buf.WriteString("\r\n")
		}
		buf.WriteString(doc)
	}
	resp, err := client.cli.Post(path, "application/json", &buf)
	if err != nil {
		return 0, common.Errorf(fmt.Sprintf("[%s] add document err", path), err)
	}
	defer resp.Body.Close()
	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, common.Errorf("read http response err", err)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, common.Errorf("", errors.New(fmt.Sprintf("http status:%d,msg:%s", resp.StatusCode, string(resultBytes))))
	} else {
		logger.Infof("add document response:%s", string(resultBytes))
	}
	respMap := make(map[string]int)
	err = json.Unmarshal(resultBytes, &respMap)
	if err != nil {
		return 0, common.Errorf("unmarshal response err", err)
	}
	return respMap["num_docs_for_processing"], nil
}

func (client *QuickwitClient) Search(index string, searchReq map[string]any) ([]byte, error) {
	path, err := url.JoinPath(client.server_url, index, "search")
	if logger.IsLevelEnabled(logger.DebugLevel) {
		logger.Debug("path:", path)
	}

	if err != nil {
		return nil, common.Errorf(fmt.Sprintf("%s joinpath err", client.server_url), err)
	}

	if _, ok := searchReq["query"]; !ok {
		return nil, common.Errorf("", errors.New("search request need contain field:query"))
	}
	reqBytes, _ := json.Marshal(searchReq)

	resp, err := client.cli.Post(path, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, common.Errorf(fmt.Sprintf("[%s] search request err", path), err)
	}
	defer resp.Body.Close()
	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, common.Errorf("read http response err", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, common.Errorf("", errors.New(fmt.Sprintf("http status:%d,msg:%s", resp.StatusCode, string(resultBytes))))
	} else {
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Infof("add document response:%s", string(resultBytes))
		}
	}
	return resultBytes, nil
}

func (client *QuickwitClient) Close() {

}
