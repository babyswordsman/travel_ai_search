package searchengineapi

import (
	"context"
	"sort"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/qdrant"

	logger "github.com/sirupsen/logrus"
)

type LocalSearchEngine struct {
}

func (engine *LocalSearchEngine) Search(ctx context.Context, config *conf.Config, query string) ([]SearchItem, error) {
	return engine.internalSearch(query, config.PreRankingThreshold)
}
func (engine *LocalSearchEngine) internalSearch(query string, threshold float32) ([]SearchItem, error) {
	vectors, err := modelclient.GetInstance().QueryEmbedding([]string{query})
	if err != nil {
		logger.Errorf("query embedding err:%s", err)
		return nil, err
	}

	scores, err := qdrant.GetInstance().Search(qdrant.DETAIL_COLLECTION,
		vectors[0], uint64(conf.GlobalConfig.MaxCandidates), false, true)
	if err != nil {
		logger.Errorf("{%s},search err:%s", query, err)
		return nil, err
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].GetScore() < scores[j].GetScore()
	})

	keys := make([]string, 0, len(scores))
	scoreMap := make(map[string]float32)
	for i := len(scores) - 1; i >= 0; i-- {
		scoreNode := scores[i]
		logger.WithField("query", query).Infof("score:%f,key:%s", scoreNode.GetScore(), scoreNode.GetPayload()["id"].GetStringValue())
		if threshold > scoreNode.GetScore() {
			continue
		}
		key := scoreNode.GetPayload()["id"].GetStringValue()
		scoreMap[key] = scoreNode.GetScore()
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return make([]SearchItem, 0), nil
	}

	details := make([]SearchItem, 0, len(keys))
	for _, key := range keys {
		detail, err := kvclient.GetInstance().HGetAll(key)
		if err != nil {
			logger.WithField("key", key).Error("fetch detail err", err)
			continue
		}
		v, ok := scoreMap[key]
		if !ok {
			v = 0.0
		}
		item := SearchItem{
			Title:    detail[conf.DETAIL_TITLE_FIELD],
			Snippet:  detail[conf.DETAIL_CONTENT_FIELD],
			Link:     "/detail?id=" + key,
			Score:    v,
			IsSearch: false,
		}
		details = append(details, item)
	}

	return details, nil

}
