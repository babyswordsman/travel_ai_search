package es

import (
	"encoding/json"
	"testing"
	"time"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/shopping/detail"
)

func start(t *testing.T) {
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config

	InitESClient(conf.GlobalConfig)
}

func stop() {
	GetInstance().Close()
}

func TestCreateIndex(t *testing.T) {
	start(t)
	defer stop()
	index := "sku"
	sku_index_mapping := `
	
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

	err := GetInstance().CreateIndex(index, sku_index_mapping)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetIndex(t *testing.T) {
	start(t)
	defer stop()
	str, err := GetInstance().GetIndex("sku")
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(str)
}

func TestDeleteIndex(t *testing.T) {
	start(t)
	defer stop()
	str, err := GetInstance().DeleteIndex("sku")
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("del success:" + str)
}

func TestAddIndex(t *testing.T) {
	start(t)
	defer stop()
	line := `{"brand":"Apple","productName":"Apple/苹果 iPhone 15 (A3092) 256GB 黑色 支持移动联通电信5G 双卡双待手机","productNumber":"商品编号：100066896338","productPrice":"5999.00","productLink":"https://item.jd.com/100066896338.html\n","firstLevel":"手机通讯","secondLevel":"手机","thirdLevel":"手机","store":"","productPicture":{"main":["https://img13.360buyimg.com/n1/s450x450_jfs/t1/245925/18/12621/39175/66753e71F7b337d75/956610f0ad946d52.jpg.avif\n"],"other":null},"productCommentSummary":{"commentCountStr":"200万+","goodRateShow":95,"goodCountStr":"36万+","generalCountStr":"5500+","poorCountStr":"1.2万+","imageListCount":500,"videoCountStr":"3400+","hotCommentTagStatistics":[{"name":"手感一流"},{"name":"漂亮大方"},{"name":"时尚简约"},{"name":"色彩饱满"},{"name":"性能一流"},{"name":"散热性佳"},{"name":"操作简单"},{"name":"品质优良"},{"name":"送给TA"},{"name":"解锁迅速"},{"name":"畅快办公"},{"name":"游戏上分"},{"name":"按键灵敏"}]},"sales_guarantee":"厂家服务\n        \n        \n                                                                                                本商品质保期周期1年质保，在此时间范围内可提交维修申请，具体请以厂家服务为准。\n                                                                                                                                                    如因质量问题或故障，凭厂商维修中心或特约维修点的质量检测证明，享受7日内退货，15日内换货，15日以上在质保期内享受免费保修等三包服务！(注:如厂家在商品介绍中有售后保障的说明,则此商品按照厂家说明执行售后保障服务。)\n                                                                                                                    \n                                \n            \n            京东承诺\n        \n        \n                            京东平台卖家销售并发货的商品，由平台卖家提供发票和相应的售后服务。请您放心购买！\n                                        注：因厂家会在没有任何提前通知的情况下更改产品包装、产地或者一些附件，本司不能确保客户收到的货物与商城图片、产地、附件说明完全一致。只能确保为原厂正货！并且保证与当时市场上同样主流新品一致。若本商城没有及时更新，请大家谅解！\n        \n                                \n            \n             正品行货             \n        \n                        京东商城向您保证所售商品均为正品行货，京东自营商品开具机打发票或电子发票。\n                                                        无忧退货\n        \n            客户购买京东自营商品7日内（含7日，自客户收到商品之日起计算），在保证商品完好的前提下，可无理由退货。（部分商品除外，详情请见各商品细则）","extendedProperties":{"CPU型号":"A16","三防标准":"IP68","充电功率":"以官网信息为准","后摄主像素":"4800万像素","商品名称":"AppleiPhone15","商品编号":"100066896338","屏幕分辨率":"FHD+","屏幕材质":"OLED直屏","支持IPv6":"支持IPv6","机身内存":"256GB","机身色系":"黑色系","机身颜色":"黑色","泰尔电竞评级":"否","运行内存":"未公布","风格":"科技，商务"},"relevant_product":null}`
	doc, err := detail.CrawleJsonToSku(line)
	if err != nil {
		t.Error(err.Error())
	}
	v, err := json.Marshal(doc)
	if err != nil {
		t.Error(err.Error())
	}
	num, err := GetInstance().AddDocument("sku", time.Now().Format("2006-01-02T15:04:05"), string(v))
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("success:", num)
}

func TestSearchIndex(t *testing.T) {
	start(t)
	defer stop()

	query := map[string]interface{}{
		"size": 2,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"match": map[string]interface{}{
						"third_level": "电子产品-手机",
					}},
					{"match": map[string]interface{}{
						"second_level": "电子产品-手机",
					}},
					{"match": map[string]interface{}{
						"first_level": "电子产品-手机",
					}},
					{"match": map[string]interface{}{
						"extended_props": "颜色:红色",
					}},
					{"match": map[string]interface{}{
						"product_name": "大内存红色手机",
					}},
				},
			},
		},
	}
	r, err := GetInstance().SearchIndex("sku", query)
	if err != nil {
		t.Error(err.Error())
	}
	// Print the response status, number of results, and request duration.
	hits := int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	took := int(r["took"].(float64))
	t.Log("hits:", hits, ",took:", took)
	// Print the ID and document source for each hit.
	docs := make([]*detail.SkuDocumentResp, 0)
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		id := hit.(map[string]interface{})["_id"]
		source := hit.(map[string]interface{})["_source"]
		score := hit.(map[string]interface{})["_score"]
		skuDoc := detail.EsTransferSku(source.(map[string]interface{}))

		skuDoc.Id = id.(string)
		skuDoc.Score = score.(float64)
		docs = append(docs, skuDoc)
	}

	if len(docs) == 0 {
		t.Errorf("empty docs")
	}
	for _, doc := range docs {
		t.Logf("id:%s,name:%s,props:%s", doc.Id, doc.ProductName, doc.ExtendedProps)
	}
}
