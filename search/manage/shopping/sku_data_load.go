package shopping

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"travel_ai_search/search/es"
	"travel_ai_search/search/shopping/detail"

	logger "github.com/sirupsen/logrus"
)

var sku_index_mapping string = `
	
  {
    "settings": {
    "analysis": {
      "analyzer": {
        "std_analyzer": { 
          "type":      "standard",
          "stopwords": "_english_"
        }
      }
    }
  },
    "mappings": {
      "properties": {
        "timestamp": {
          "type":"date","format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"
        },
        "store_name": {
        //"comment":"店铺名称",
          "type":"text",
          "analyzer": "std_analyzer"
        },
        "product_main_pic": {
          //"comment": "商品主图",
          "type": "text",
          "analyzer": "std_analyzer"
        },
        "product_name": {
          //"comment": "商品名称",
          "type": "text",
          "analyzer": "std_analyzer"
        },
        "brand": {
          //"comment": "品牌名",
          "type": "text",
          "analyzer": "std_analyzer"
        },
        "first_level": {
          //"comment": "一级类目",
          "type": "text",
          "analyzer": "std_analyzer"
        },
        "second_level": {
          //"comment": "二级类目",
          "type": "text",
          "analyzer": "std_analyzer"
        },
        "third_level": {
          //"comment": "三级类目",
          "type": "text",
          "analyzer": "std_analyzer"
        },
        "product_price": {
          //"comment": "价格",
          "type": "double"
          
        },
        "extended_props": {
          //"comment": "商品扩展属性",
          "type": "text",
          "analyzer": "std_analyzer"
        },
        "comment_summary": {
          //"comment": "商品评论信息",
          "type": "text",
          "analyzer": "std_analyzer"
        }
      }
    }
  }
	`

func ParseSkuData(path string) int32 {
	logger.Info("enter method ParseSkuData")
	locked := loadSkuMutex.TryLock()
	if locked {
		defer loadSkuMutex.Unlock()
	} else {
		return 0
	}
	logger.Info("start load sku")
	index_name := "sku"
	_, err := es.GetInstance().GetIndex(index_name)
	if err != nil {
		err = es.GetInstance().CreateIndex(index_name, sku_index_mapping)
		if err != nil {
			logger.Errorf("create index err:%s", err.Error())
			return 0
		}
	}

	parseNum := int32(0)
	f, err := os.Open(path)
	if err != nil {
		logger.Errorln("open file err", err)
		return parseNum
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

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

		skuDoc := detail.CrawleTransferSku(skuMap)
		//skuDoc.Id = fmt.Sprintf("sku-%d-%d", time.Now().UnixMilli(), parseNum)
		buf, err := json.Marshal(skuDoc)

		if err != nil {
			logger.Errorln("Marshal err", err)
			continue
		}
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Debugln(string(buf))
		}

		i, err := es.GetInstance().AddDocument(index_name, fmt.Sprintf("id-%d-%d", time.Now().UnixMilli(), parseNum), string(buf))
		if err != nil {
			logger.Errorln("add document err", err)
			continue
		}
		logger.Infof("add documnent size:%d", i)

		parseNum++
		if parseNum%20 == 19 {
			logger.Infof("deal rows:%d", parseNum)
		}
	}

	return parseNum
}

var walmart_sku_index_mapping string = `
	
	{
	  "settings": {
	  "analysis": {
		"analyzer": {
		  "std_analyzer": { 
			"type":      "standard",
			"stopwords": "_english_"
		  }
		}
	  }
	},
	  "mappings": {
		"properties": {
		  "timestamp": {
			"type":"date","format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"
		  },
		  "aisle":{
		  	"type":"keyword"
		  },
		  "parentItemId":{
		  	"type": "long"
		  },
		  "color": {
			"type":"keyword"
		  },
		  "mediumImage": {
			"type": "keyword"
		  },
		  "name": {
			"type": "text",
			"analyzer": "std_analyzer"
		  },
		  "brandName": {
			"type": "keyword"
		  },
		  "categoryPath": {
			"type": "text",
			"analyzer": "std_analyzer"
		  },
		  "salePrice": {
			"type": "double"
		  },
		  "shortDescription": {
			"type": "text",
			"analyzer": "std_analyzer"
		  },
		  "longDescription": {
			"type": "text",
			"analyzer": "std_analyzer"
		  }
		}
	  }
	}
	  `

func LoadWalmartSkuFiles(parentDir string) int32 {
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		logger.Errorf("read %s err", parentDir)
		return 0
	}
	total := int32(0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := fmt.Sprintf("%s/%s", parentDir, e.Name())
		num := ParseWalmartSkuData(path)
		if num == 0 {
			logger.Errorf("parse %s err", path)
		} else {
			logger.Infof("%s items :%d", path, num)
		}
		total += num
	}
	return total
}
func ParseWalmartSkuData(path string) int32 {
	logger.Info("enter method ParseWalmartSkuData")
	locked := loadSkuMutex.TryLock()
	if locked {
		defer loadSkuMutex.Unlock()
	} else {
		return 0
	}
	logger.Info("start load wal_sku")
	index_name := "wal_sku"
	_, err := es.GetInstance().GetIndex(index_name)
	if err != nil {
		err = es.GetInstance().CreateIndex(index_name, walmart_sku_index_mapping)
		if err != nil {
			logger.Errorf("create index err:%s", err.Error())
			return 0
		}
	}

	parseNum := int32(0)

	buf, err := os.ReadFile(path)
	if err != nil {
		logger.Errorf("read file:%s err:%s", path, err.Error())
		return 0
	}
	respMap := make(map[string]any)
	err = json.Unmarshal(buf, &respMap)

	if err != nil {
		logger.Errorf("parse file err:%s", err.Error())
		return 0
	}

	items := respMap["items"].([]any)
	rand.Seed(time.Now().UnixMilli())
	for _, itemMap := range items {
		randomNumber := rand.Intn(10) + 1
		itemMapNew := itemMap.(map[string]any)
		itemMapNew["aisle"] = fmt.Sprintf("A%d", randomNumber)
		skuDoc := detail.ParseWalmartSku(itemMapNew)

		buf, err := json.Marshal(skuDoc)

		if err != nil {
			logger.Errorln("Marshal err", err)
			continue
		}
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Debugln(string(buf))
		}

		i, err := es.GetInstance().AddDocument(index_name, strconv.FormatInt(skuDoc.ItemId, 10), string(buf))
		if err != nil {
			logger.Errorln("add document err", err)
			continue
		}
		logger.Infof("add documnent size:%d", i)

		parseNum++
		if parseNum%20 == 19 {
			logger.Infof("deal rows:%d", parseNum)
		}
	}

	return parseNum
}
