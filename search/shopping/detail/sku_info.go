package detail

import (
	"encoding/json"
	"strconv"
	"time"
)

type SkuDetail struct {
	Title        string            //标题
	CategoryPath []string          //多级类目
	SkuProps     map[string]string //销售属性
	Describe     string            //描述
	SourceId     string            //原始item_id
}

type RecommendSku struct {
	Id     string
	Score  float64
	Reason string
}

type SkuDocument struct {
	Id             string  `json:"-"`
	Timestamp      int64   `json:"timestamp"`
	StoreName      string  `json:"store_name"`
	ProductMainPic string  `json:"product_main_pic"`
	ProductName    string  `json:"product_name"`
	Brand          string  `json:"brand"`
	FirstLevel     string  `json:"first_level"`
	SecondLevel    string  `json:"second_level"`
	ThirdLevel     string  `json:"third_level"`
	ProductPrice   float64 `json:"product_price"`
	//ExtendedProps  map[string]any `json:"extended_props"`
	//CommentSummary map[string]any `json:"comment_summary"`
	ExtendedProps  string `json:"extended_props"`
	CommentSummary string `json:"comment_summary"`
}
type SkuDocumentResp struct {
	SkuDocument
	Score float64
}

func CrawleJsonToSku(content string) (*SkuDocument, error) {
	skuMap := make(map[string]any)
	err := json.Unmarshal([]byte(content), &skuMap)
	if err != nil {
		return nil, err
	}
	return CrawleTransferSku(skuMap), nil
}

func CrawleTransferSku(skuMap map[string]any) *SkuDocument {
	var skuDoc SkuDocument
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
				skuDoc.ProductPrice = price
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

func EsTransferSku(skuMap map[string]any) *SkuDocumentResp {
	var skuDoc SkuDocumentResp
	skuDoc.Timestamp = time.Now().UnixMilli()
	storeName, ok := skuMap["store_name"]
	if ok {
		store, ok := storeName.(string)
		if ok {
			skuDoc.StoreName = store
		}
	}

	mainPic, ok := skuMap["product_main_pic"]
	if ok {

		mainPic, ok := mainPic.(string)
		if ok {
			skuDoc.ProductMainPic = mainPic
		}

	}

	ProductName, ok := skuMap["product_name"]
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
	FirstLevel, ok := skuMap["first_level"]
	if ok {
		str, ok := FirstLevel.(string)
		if ok {
			skuDoc.FirstLevel = str
		}
	}
	SecondLevel, ok := skuMap["second_level"]
	if ok {
		str, ok := SecondLevel.(string)
		if ok {
			skuDoc.SecondLevel = str
		}
	}
	ThirdLevel, ok := skuMap["third_level"]
	if ok {
		str, ok := ThirdLevel.(string)
		if ok {
			skuDoc.ThirdLevel = str
		}
	}
	ProductPrice, ok := skuMap["product_price"]
	if ok {
		price, ok := ProductPrice.(float64)
		if ok {
			skuDoc.ProductPrice = price
		}
	}
	ExtendedProps, ok := skuMap["extended_props"]
	if ok {
		props, ok := ExtendedProps.(string)
		if ok {
			skuDoc.ExtendedProps = props
		}

	}
	CommentSummary, ok := skuMap["comment_summary"]

	if ok {
		summary, ok := CommentSummary.(string)
		if ok {
			skuDoc.CommentSummary = summary
		}

	}
	return &skuDoc
}
