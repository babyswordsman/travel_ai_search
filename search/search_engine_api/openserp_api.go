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

type OpenSerpSearchEngine struct {
	Engines []string
	BaseUrl string
}

type OpenSerpResponse struct {
	Items []OpenSerpItem `json:"items"`
}

type OpenSerpItem struct {
	Rank        int
	Url         string
	Title       string
	Description string
}

func (engine *OpenSerpSearchEngine) Search(ctx context.Context, config *conf.Config, query string) ([]SearchItem, error) {
	reqURL := fmt.Sprintf("%s?engine=%s&text=%s", engine.BaseUrl, engine.Engines[0], url.PathEscape(query))

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
	var resp OpenSerpResponse
	if err := json.Unmarshal(body, &resp.Items); err != nil {
		return nil, fmt.Errorf("unmarshal openserp search response: %s", err)
	}
	searchItems := make([]SearchItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		searchItems = append(searchItems, SearchItem{Link: item.Url, Snippet: item.Description, Title: item.Title})
	}

	return searchItems, nil
}
