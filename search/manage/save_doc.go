package manage

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	docextract "travel_ai_search/search/doc_extract"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/qdrant"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

func extractFile(path string) (string, []schema.Document, error) {
	extractor := docextract.DocconvExtractor{Path: path}
	_, pages, err := extractor.Extract()
	if err != nil {
		return "", make([]schema.Document, 0), fmt.Errorf("split text err:%w", err)
	}
	docs := make([]schema.Document, 0)
	buf := strings.Builder{}
	//长度截断
	for _, page := range pages {
		buf.WriteString(page.Content)
		//todo:上下page是否要拼接，属于table的page是否单独处理
		chunks, err := textsplitter.NewTokenSplitter().SplitText(page.Content)
		if err != nil {
			return "", make([]schema.Document, 0), fmt.Errorf("split text err:%w", err)
		}

		for _, chunk := range chunks {
			doc := schema.Document{
				PageContent: chunk,
				Score:       0.0,
				Metadata: map[string]any{
					"url": path,
				},
			}
			docs = append(docs, doc)
		}
	}
	return buf.String(), docs, nil
}

func deal_chunk(space string, fileName string, docId string, seqno int, document *schema.Document) error {
	id, err := kvclient.FetchDetailNextId()
	if err != nil {
		return common.Errorf("fetch id", err)
	}
	key := docId + "_" + strconv.FormatUint(uint64(seqno), 10)
	tmp := map[string]string{
		conf.DETAIL_TITLE_FIELD:         fileName,
		conf.DETAIL_CONTENT_FIELD:       document.PageContent,
		conf.DETAIL_CONTENT_CHUNK_FIELD: strconv.FormatInt(int64(seqno), 10),
		conf.DETAIL_CONTENT_DOC_FIELD:   docId,
		"space":                         space,
	}
	values := make([]interface{}, 0)
	for k, v := range tmp {
		values = append(values, k, v)
	}
	err = kvclient.GetInstance().HMSet(key, values)
	if err != nil {
		return common.Errorf(fmt.Sprintf("set key:{%s}", key), err)
	}

	vector, err := modelclient.GetInstance().PassageEmbedding([]string{tmp[conf.DETAIL_CONTENT_FIELD]})
	if err != nil {
		return common.Errorf(fmt.Sprintf("filename:{%s-%d}", fileName, seqno), err)
	}

	row := qdrant.NewVectorRow(id, vector[0])
	row.AppendString("id", key)
	row.AppendString(conf.DETAIL_TITLE_FIELD, tmp[conf.DETAIL_TITLE_FIELD])
	row.AppendString("space", space)

	err = qdrant.GetInstance().AddVector(qdrant.DETAIL_COLLECTION, []*qdrant.VectorRow{row})
	if err != nil {
		return common.Errorf(fmt.Sprintf("{%s}create vector index", key), err)
	}
	return nil
}

func deal_doc(space string, fileName string, content string) (string, error) {
	id, err := kvclient.FetchDetailNextId()
	if err != nil {
		return "", common.Errorf("fetch id", err)
	}
	key := conf.DOC_KEY_PREFIX + strconv.FormatUint(id, 10)
	tmp := map[string]string{
		conf.DETAIL_TITLE_FIELD:   fileName,
		conf.DETAIL_CONTENT_FIELD: content,
		"space":                   space,
	}
	values := make([]interface{}, 0)
	for k, v := range tmp {
		values = append(values, k, v)
	}
	err = kvclient.GetInstance().HMSet(key, values)
	if err != nil {
		return "", common.Errorf(fmt.Sprintf("set key:{%s}", key), err)
	}
	return key, nil
}

/*
*
对文档进行分块保存，原始数据保存为：
"doc-id":map[chunk_id]text
*/
func CreateDocIndex(path string, space string) error {
	content, pages, err := extractFile(path)
	if err != nil {
		return common.Errorf(path, err)
	}

	dir, fileName := filepath.Split(path)
	if fileName == "" {
		fileName = dir
	}
	docId, err := deal_doc(space, fileName, content)
	if err != nil {
		return common.Errorf(fileName, err)
	}

	for seqno := range pages {
		err = deal_chunk(space, fileName, docId, seqno+1, &pages[seqno])
		if err != nil {
			return common.Errorf(fmt.Sprintf("%s-%d", fileName, seqno), err)
		}
	}
	return nil
}
