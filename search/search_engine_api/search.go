package searchengineapi

import (
	"context"
	"travel_ai_search/search/conf"
)

type SearchItem struct {
	Title    string  `json:"title"`
	Link     string  `json:"link"`
	Snippet  string  `json:"snippet"`
	IsSearch bool    `json:"is_search"`
	Score    float32 `json:"score"`
}

type SearchEngine interface {
	Search(ctx context.Context, config *conf.Config, query string) ([]SearchItem, error)
}
