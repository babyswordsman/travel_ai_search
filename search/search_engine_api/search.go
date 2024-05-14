package searchengineapi

import (
	"context"
	"travel_ai_search/search/conf"

	"github.com/tmc/langchaingo/schema"
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

func SnakeMerge[T SearchItem | schema.Document](maxItems int, candidates ...[]T) []T {
	mergedItems := make([]T, 0, maxItems)
	maxLen := 0
	for i := range candidates {
		maxLen = max(maxLen, len(candidates[i]))
	}
	for ind := 0; ind < maxLen; ind++ {
		for i := range candidates {
			if ind < len(candidates[i]) {
				mergedItems = append(mergedItems, candidates[i][ind])
				if len(mergedItems) >= maxItems {
					break
				}
			}
		}
		if len(mergedItems) >= maxItems {
			break
		}
	}
	return mergedItems
}
