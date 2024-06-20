package shopping

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/qdrant"

	logger "github.com/sirupsen/logrus"

	go_client "github.com/qdrant/go-client/qdrant"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

const VEC_ROWID_KEY = "rowid"
const REL_ITEMID_KEY = "id"

type QdrantStore struct {
	embedder       embeddings.Embedder
	collectionName string
	topK           int
	client         *qdrant.VectorClient
	scoreThreshold float32
}

func NewVector() QdrantStore {
	embedder := NewEmbedder()
	client := qdrant.GetInstance()
	s := QdrantStore{
		embedder:       embedder,
		client:         client,
		collectionName: qdrant.SKU_COLLECTION,
		topK:           int(conf.GlobalConfig.MaxCandidates),
		scoreThreshold: conf.GlobalConfig.PreRankingThreshold,
	}
	return s
}

func (s QdrantStore) AddDocuments(ctx context.Context,
	docs []schema.Document,
	_ ...vectorstores.Option,
) ([]string, error) {
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors,
		err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, errors.New("number of vectors from embedder does not match number of documents")
	}

	//todo:batch
	rows := make([]*qdrant.VectorRow, 0, len(vectors))
	rowIds := make([]string, 0, len(vectors))
	for i := 0; i < len(docs); i++ {
		rowid := docs[i].Metadata[VEC_ROWID_KEY]
		id, err := rowid.(uint64)

		if !err {
			vkind := reflect.ValueOf(rowid)
			return nil, fmt.Errorf("{%v} type is not int64,type:%s", rowid, vkind.Kind().String())
		}
		row := qdrant.NewVectorRow(id, vectors[i])
		rows = append(rows, row)
		rowIds = append(rowIds, strconv.FormatUint(id, 10))
		for key, value := range docs[i].Metadata {

			vkind := reflect.ValueOf(value)

			if vkind.Kind() == reflect.String {
				row.AppendString(key, vkind.String())
			} else if vkind.CanInt() {
				row.AppendNum(key, int64(vkind.Int()))
			} else if vkind.CanFloat() {
				row.AppendDouble(key, vkind.Float())
			} else {
				return nil, common.Errorf(fmt.Sprintf("unimplement type :{%s}", vkind.Kind().String()), errors.ErrUnsupported)
			}
		}

	}

	err = qdrant.GetInstance().AddVector(qdrant.DETAIL_COLLECTION, rows)
	if err != nil {
		return nil, common.Errorf(fmt.Sprintf("{%s}insert vector index ", qdrant.DETAIL_COLLECTION), err)
	}
	return rowIds, nil
}

func (s QdrantStore) SimilaritySearch(ctx context.Context,
	query string, numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {

	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, common.Errorf(fmt.Sprintf("query:{%s} embedQuery err", query), err)
	}

	points, err := s.client.SearchWithFilter(s.collectionName, vector, nil, uint64(s.topK), false, true)
	if err != nil {
		return nil, common.Errorf(fmt.Sprintf("query:{%s} search vector err", query), err)
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].GetScore() < points[j].GetScore()
	})

	docs := make([]schema.Document, 0, len(points))
	for i := len(points) - 1; i >= 0; i-- {
		scoreNode := points[i]
		if s.scoreThreshold > scoreNode.GetScore() {
			continue
		}
		doc := schema.Document{
			PageContent: "",
			Score:       scoreNode.Score,
		}

		meta := make(map[string]any)
		meta[VEC_ROWID_KEY] = scoreNode.Id.GetNum()

		for k, x := range scoreNode.GetPayload() {
			switch v := x.GetKind().(type) {
			case *go_client.Value_StringValue:
				meta[k] = v.StringValue
				if k == "title" {
					doc.PageContent = v.StringValue
				}
			case *go_client.Value_DoubleValue:
				meta[k] = v.DoubleValue
			case *go_client.Value_IntegerValue:
				meta[k] = v.IntegerValue
			default:
				return nil, common.Errorf(fmt.Sprintf("Unimplemented type:{%v} k:%s", v, k), errors.ErrUnsupported)
			}

		}

		doc.Metadata = meta
		docs = append(docs, doc)
		if logger.IsLevelEnabled(logger.DebugLevel) {
			buf, _ := json.Marshal(doc)
			logger.Debugf("search candidate:{%s}", string(buf))
		}
	}

	return docs, nil

}
