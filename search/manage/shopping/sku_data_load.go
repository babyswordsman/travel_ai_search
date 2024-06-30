package shopping

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
	"travel_ai_search/search/quickwit"
	"travel_ai_search/search/shopping/detail"

	logger "github.com/sirupsen/logrus"
)

func ParseSkuData(path string) int32 {
	parseNum := int32(0)
	f, err := os.Open(path)
	if err != nil {
		logger.Errorln("open file err", err)
		return parseNum
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	skuBatch := make([]string, 0, 10)
	for scanner.Scan() {

		line := scanner.Text()
		line = strings.TrimSpace(line)

		if len(line) <= 2 {
			continue
		}
		skuMap := make(map[string]any)
		err := json.Unmarshal([]byte(line), &skuMap)
		if err != nil {
			logger.WithField("line", line).Errorln("json unmarshal err:", err)
			continue
		}

		skuDoc := transferSku(skuMap)

		buf, err := json.Marshal(skuDoc)

		if err != nil {
			logger.Errorln("Marshal err", err)
			continue
		}
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Debugln(string(buf))
		}
		skuBatch = append(skuBatch, string(buf))
		if len(skuBatch) >= 10 {
			i, err := quickwit.GetInstance().AddDocument("sku_a", skuBatch)
			if err != nil {
				logger.Errorln("add document err", err)
				continue
			}
			logger.Infof("add documnent size:%d", i)
			clear(skuBatch)
		}
		parseNum++
		if parseNum%20 == 19 {
			logger.Infof("deal rows:%d", parseNum)
		}
	}
	if len(skuBatch) > 0 {
		i, err := quickwit.GetInstance().AddDocument("sku_a", skuBatch)
		if err != nil {
			logger.Errorln("add document err", err)
		}
		logger.Infof("add documnent size:%d", i)
		clear(skuBatch)
	}
	return parseNum
}

func JsonToSku(content string) (*detail.SkuDocument, error) {
	skuMap := make(map[string]any)
	err := json.Unmarshal([]byte(content), &skuMap)
	if err != nil {
		return nil, err
	}
	return transferSku(skuMap), nil
}

func transferSku(skuMap map[string]any) *detail.SkuDocument {
	var skuDoc detail.SkuDocument
	skuDoc.Timestamp = time.Now().UnixMilli()
	storeName, ok := skuMap["store"]
	if ok {
		store, ok := storeName.(string)
		if ok {
			skuDoc.StoreName = store
		}
	}

	ProductMainPic, ok := skuMap["productPicture"]
	if ok {
		mainPic, ok := ProductMainPic.(map[string]any)
		if ok {
			pics, ok := mainPic["main"]
			if ok {
				pics, ok := pics.([]any)
				if ok && len(pics) > 0 {
					pic, ok := pics[0].(string)
					if ok {
						skuDoc.ProductMainPic = pic
					}

				}
			}
		}
	}

	ProductName, ok := skuMap["productName"]
	if ok {
		str, ok := ProductName.(string)
		if ok {
			skuDoc.ProductName = str
		}
	}
	Brand, ok := skuMap["brand"]
	if ok {
		str, ok := Brand.(string)
		if ok {
			skuDoc.Brand = str
		}
	}
	FirstLevel, ok := skuMap["firstLevel"]
	if ok {
		str, ok := FirstLevel.(string)
		if ok {
			skuDoc.FirstLevel = str
		}
	}
	SecondLevel, ok := skuMap["secondLevel"]
	if ok {
		str, ok := SecondLevel.(string)
		if ok {
			skuDoc.SecondLevel = str
		}
	}
	ThirdLevel, ok := skuMap["thirdLevel"]
	if ok {
		str, ok := ThirdLevel.(string)
		if ok {
			skuDoc.ThirdLevel = str
		}
	}
	ProductPrice, ok := skuMap["productPrice"]
	if ok {
		v, ok := ProductPrice.(string)
		if ok {
			price, err := strconv.ParseFloat(v, 64)
			if err == nil {
				skuDoc.ProductPrice = float32(price)
			}

		}
	}
	ExtendedProps, ok := skuMap["extendedProperties"]
	if ok {
		props, ok := ExtendedProps.(map[string]any)
		if ok {
			//skuDoc.ExtendedProps = props
			propsBuf, _ := json.Marshal(props)
			skuDoc.ExtendedProps = string(propsBuf)
		}

	}
	CommentSummary, ok := skuMap["productCommentSummary"]

	if ok {
		summary, ok := CommentSummary.(map[string]any)
		if ok {
			//skuDoc.CommentSummary = summary
			summaryBuf, _ := json.Marshal(summary)
			skuDoc.CommentSummary = string(summaryBuf)
		}

	}
	return &skuDoc
}
