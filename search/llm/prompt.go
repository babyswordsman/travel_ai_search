package llm

import (
	"fmt"
	"sort"
	"strings"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/modelclient"
	searchengineapi "travel_ai_search/search/search_engine_api"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/schema"
)

type Prompt interface {
	GenPrompt(candidates []schema.Document) (string, []schema.Document, error)
}

type TravelPrompt struct {
	MaxLength    int
	PromptPrefix string
}

func (prompt *TravelPrompt) GenPrompt(candidates []schema.Document) (string, []schema.Document, error) {
	refDocuments := make([]schema.Document, 0, len(candidates))
	//todo:截断,是否添加标题
	buf := strings.Builder{}
	buf.WriteString(prompt.PromptPrefix)
	buf.WriteString("\r\n")
	remain := prompt.MaxLength
	for ind := range candidates {
		item := candidates[ind]
		content := []rune(item.PageContent)
		contentLen := len(content)
		if remain-contentLen > 0 {

			buf.WriteString("\r\n")
			buf.WriteString(string(content))
			buf.WriteString("\r\n")
			remain = remain - contentLen
			refDocuments = append(refDocuments, candidates[ind])
		} else if ind == 0 {
			buf.WriteString("\r\n")
			buf.WriteString(string(content[:remain]))
			buf.WriteString("\r\n")
			remain = 0
			refDocuments = append(refDocuments, candidates[ind])
			break
		} else {
			break
		}

	}
	buf.WriteString("\r\n")
	logger.Debug(buf.String())
	return buf.String(), refDocuments, nil
}

type ChatPrompt struct {
	MaxLength    int
	PromptPrefix string
}

func (prompt *ChatPrompt) GenPrompt(candidates []schema.Document) (string, []schema.Document, error) {
	refDocuments := make([]schema.Document, 0, len(candidates))
	//todo:截断
	buf := strings.Builder{}
	buf.WriteString(prompt.PromptPrefix)
	buf.WriteString("\r\n")
	remain := prompt.MaxLength
	for ind := range candidates {
		item := candidates[ind]
		content := []rune(item.PageContent)
		contentLen := len(content)
		buf.WriteString("\r\n")
		if remain-contentLen > 0 {
			buf.WriteString(string(content))
			remain = remain - contentLen
			refDocuments = append(refDocuments, candidates[ind])
		} else if ind == 0 {
			buf.WriteString(string(content[:remain]))
			remain = 0
			refDocuments = append(refDocuments, candidates[ind])
			break
		} else {
			break
		}
		buf.WriteString("\r\n")
	}
	buf.WriteString("\r\n")
	logger.Info(buf.String())
	return buf.String(), refDocuments, nil
}

func PreRankDoc(query string, candidates []searchengineapi.SearchItem) ([]schema.Document, error) {
	//todo:加载网页，计算相似度可以并行
	docs := make([]schema.Document, 0)
	if len(candidates) > int(conf.GlobalConfig.MaxCandidates) {
		candidates = candidates[:conf.GlobalConfig.MaxCandidates]
	}
	for ind := range candidates {
		item := candidates[ind]
		if item.IsSearch {
			tmps, err := searchengineapi.LoadHtml(item.Link)
			if err != nil {
				logger.Errorf("load html err:%s", err.Error())
				continue
			}
			for i := range tmps {
				tmps[i].Metadata["title"] = item.Title
				tmps[i].Metadata["url"] = item.Link
			}
			docs = append(docs, tmps...)
		} else {
			docs = append(docs, schema.Document{
				PageContent: item.Snippet,
				Score:       item.Score,
				Metadata:    map[string]any{"url": item.Link, "title": item.Title},
			})
		}
	}
	passages := make([]string, 0, len(docs)+1)
	for i := 0; i < len(docs); i++ {
		passages = append(passages, docs[i].PageContent)
	}
	passageEmbeds, err := modelclient.GetInstance().PassageEmbedding(passages)
	if err != nil {
		return docs, common.Errorf("get passage embed err:%w", err)
	}
	queryEmbeds, err := modelclient.GetInstance().QueryEmbedding([]string{query})

	if err != nil {
		return docs, common.Errorf("get query embed err:%w", err)
	}
	if len(passageEmbeds) != len(docs) {
		return docs, fmt.Errorf("len(passageEmbeds):%d != len(docs):%d ", len(passageEmbeds), len(docs))
	}
	for ind, v := range passageEmbeds {
		score, err := common.CosineSimilarity(queryEmbeds[0], v)
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		docs[ind].Score = score
	}

	rankDocs := make([]schema.Document, 0, len(docs))
	for ind := range docs {
		if docs[ind].Score > conf.GlobalConfig.PreRankingThreshold {
			rankDocs = append(rankDocs, docs[ind])
		}
	}

	//倒序
	//todo： 没有过滤相同来源的文档数量
	sort.Slice(rankDocs, func(i, j int) bool {
		return rankDocs[i].Score > rankDocs[j].Score
	})

	return rankDocs, nil
}
