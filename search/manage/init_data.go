package manage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"travel_ai_search/search/conf"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/qdrant"

	logger "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type TripDetail struct {
	Title  []string `json:"title"`
	Detail []string `json:"detail"`
}

func getHtmlNodeName(type_ html.NodeType) string {
	name := ""
	switch type_ {
	case html.ErrorNode:
		name = "error"
	case html.TextNode:
		name = "text"
	case html.DocumentNode:
		name = "document"
	case html.ElementNode:
		name = "element"
	case html.CommentNode:
		name = "comment"
	case html.DoctypeNode:
		name = "doctype"

	case html.RawNode:
		name = "raw"
	default:
		name = "other"
	}
	return name
}

func traverseNode(node *html.Node, values []string) []string {
	//str := fmt.Sprintln("type:", getHtmlNodeName(node.Type), "atom:", node.DataAtom.String(), "data:", node.Data, "attr:", node.Attr)
	//fmt.Println(str)
	//if node.Type != html.ElementNode && node.Type != html.TextNode {
	//	return values
	//}
	if node.Type == html.CommentNode {
		return values
	}
	if node.Type == html.TextNode {
		text := strings.TrimSpace(node.Data)
		if text != "" {
			values = append(values, text)
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		values = traverseNode(child, values)
	}
	if node.DataAtom == atom.P || node.DataAtom == atom.H3 || node.DataAtom == atom.Div {
		values = append(values, "\r\n")
	}
	return values
}
func ExtractText(body string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		logger.Errorln("parse html err", err)
		return nil, err
	}
	values := make([]string, 20)
	values = traverseNode(doc, values)

	return values, nil
}

func extractRawTxt(detail *TripDetail) (*strings.Builder, *strings.Builder, error) {
	var titleBuilder strings.Builder
	for ind, v := range detail.Title {
		if ind > 0 {
			titleBuilder.WriteString(",")
		}
		titleBuilder.WriteString(v)
	}

	var detailBuilder strings.Builder
	for ind, v := range detail.Detail {
		rawTexts, err := ExtractText(v)
		if err != nil {
			logger.Errorf("extract text {%s},err:%s", v, err)
			continue
		}
		if ind > 0 {
			detailBuilder.WriteString("\r\n")
		}
		moreLine := 0
		last := ""
		//抽取的文本中有多个换行符
		for _, txt := range rawTexts {

			if txt == "\r\n" {
				moreLine++
			} else {
				moreLine = 0
			}
			if moreLine > 1 {
				continue
			}
			if last != "" && last != "\r\n" {
				detailBuilder.WriteString("\r\n")
			}
			detailBuilder.WriteString(txt)
			last = txt
		}
	}
	return &titleBuilder, &detailBuilder, nil
}

func fetchEmbedding(detail map[string]string) ([]float32, error) {
	var tmp strings.Builder
	tmp.WriteString(detail[conf.DETAIL_TITLE_FIELD])
	tmp.WriteString("\r\n")
	tmp.WriteString(detail[conf.DETAIL_CONTENT_FIELD])
	vector, err := modelclient.GetInstance().PassageEmbedding([]string{tmp.String()})
	if err != nil {
		return make([]float32, 0), err
	}
	return vector[0], nil
}

func CreateIndex(detail *TripDetail) (bool, error) {
	title, content, err := extractRawTxt(detail)
	if err != nil {
		return false, err
	}

	id, err := kvclient.FetchDetailNextId()
	if err != nil {
		return false, err
	}
	key := conf.DETAIL_KEY_PREFIX + strconv.FormatUint(id, 10)
	tmp := map[string]string{
		conf.DETAIL_TITLE_FIELD:   title.String(),
		conf.DETAIL_CONTENT_FIELD: content.String(),
	}
	values := make([]interface{}, 0)
	for k, v := range tmp {
		values = append(values, k, v)
	}
	err = kvclient.GetInstance().HMSet(key, values)
	if err != nil {
		return false, fmt.Errorf("set key:{%s},err:%s", key, err)
	}

	vector, err := fetchEmbedding(tmp)
	if err != nil {
		return false, fmt.Errorf("{%s}fetch embedding err:%s", key, err)
	}

	row := qdrant.NewVectorRow(id, vector)
	row.AppendString("id", key)
	row.AppendString(conf.DETAIL_TITLE_FIELD, tmp[conf.DETAIL_TITLE_FIELD])

	err = qdrant.GetInstance().AddVector(qdrant.DETAIL_COLLECTION, []*qdrant.VectorRow{row})
	if err != nil {
		return false, fmt.Errorf("{%s}create vector index err:%s", key, err)
	}
	return true, nil
}

func ParseData(config *conf.Config, dealFunc func(*TripDetail) (bool, error)) int32 {
	parseNum := int32(0)
	f, err := os.Open(config.CrawlerDataPath)
	if err != nil {
		logger.Errorln("open file err", err)
		return parseNum
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	tripDetail := &TripDetail{}
	clean(tripDetail)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Debug("load line:", line)

		err := json.Unmarshal([]byte(line), tripDetail)
		if err != nil {
			logger.WithField("line", line).Errorln("json unmarshal err:", err)
			continue
		}
		if len(tripDetail.Title) == 0 {
			continue
		}
		parseNum++
		_, err = dealFunc(tripDetail)
		if err != nil {
			logger.Errorln("deal data err", err)
		}
		if parseNum%20 == 19 {
			logger.Infof("deal rows:%d", parseNum)
		}
	}
	return parseNum
}

func clean(tripDetail *TripDetail) {
	tripDetail.Detail = tripDetail.Detail[:0]
	tripDetail.Title = tripDetail.Title[:0]
}
