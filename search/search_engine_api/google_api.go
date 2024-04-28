package searchengineapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"travel_ai_search/search/conf"

	logger "github.com/sirupsen/logrus"
)

type GoogleSearchEngine struct {
}

func (engine *GoogleSearchEngine) Search(ctx context.Context, config *conf.Config, query string) ([]SearchItem, error) {
	return engine.googleSearch(ctx, config, query)
}

type GoogleSearchResponse struct {
	Items []SearchItem `json:"items"`
}

func (engine *GoogleSearchEngine) googleSearch(ctx context.Context, config *conf.Config, query string) ([]SearchItem, error) {
	queries := make(url.Values)
	queries.Add("hl", config.GoogleCustomSearch.Hl)
	queries.Add("lr", config.GoogleCustomSearch.Lr)
	queries.Add("cr", config.GoogleCustomSearch.Cr)
	queries.Add("q", query)

	logger.Infof("google search query:%s?%s", config.GoogleCustomSearch.Url, queries.Encode())
	//日志不要打印cx和key的值
	if !config.GoogleCustomSearch.IsProxy {
		queries.Add("cx", config.GoogleCustomSearch.Appid)
		queries.Add("key", config.GoogleCustomSearch.Key)
	}

	reqURL := fmt.Sprintf("%s?%s", config.GoogleCustomSearch.Url, queries.Encode())
	response, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("request google search err: %s", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("google search err: %s", err)
	}

	logger.Infof("response:%s", string(body))
	var resp GoogleSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal google search response: %s", err)
	}
	for i:=0;i<len(resp.Items);i++{
		resp.Items[i].IsSearch = true
		resp.Items[i].Score = 0.0
	}
	return resp.Items, nil
}
