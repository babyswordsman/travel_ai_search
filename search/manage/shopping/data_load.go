package shopping

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/qdrant"
	"travel_ai_search/search/shopping/detail"

	logger "github.com/sirupsen/logrus"
)

var loadSkuMutex sync.Mutex

func LoadFile(path string) {
	logger.Info("enter method load sku")
	locked := loadSkuMutex.TryLock()
	if locked {
		defer loadSkuMutex.Unlock()
	} else {
		return
	}
	logger.Info("start load sku")
	colls, err := qdrant.GetInstance().ListCollections()
	if err != nil {
		logger.Errorf("load qdrant collections err:%s", err.Error())
		return
	}
	flag := false
	for _, v := range colls {
		if v == qdrant.SKU_COLLECTION {
			flag = true
			break
		}
	}
	if !flag {
		err = createSkuCollection(qdrant.SKU_COLLECTION)
		if err != nil {
			logger.Errorf("create {%s} qdrant collections err:%s", qdrant.SKU_COLLECTION, err.Error())
			return
		}
	} else {
		logger.Infof("qdrant collection:%s existed", qdrant.SKU_COLLECTION)
	}
	f, err := os.Open(path)
	if err != nil {
		logger.Errorf("open file %s err:%s", path, err.Error())
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)
	parseNum := int32(0)
	for {
		fields, err := reader.Read()
		if err != nil {
			logger.Errorf("load file :%s end,reason:%s", path, err.Error())
			break
		}
		v, _ := json.Marshal(fields)

		if len(fields) != 8 {
			logger.Errorf("err len:%d,v:%v", len(fields), string(v))
		}
		skuDetail := detail.SkuDetail{
			Title:        fields[7],
			CategoryPath: []string{fields[1], fields[3], fields[5]},
			SourceId:     fields[6],
		}
		flag, err := CreateSkuIndex(&skuDetail)
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Debugf("create index for:{%s}", skuDetail.Title)
		}
		if err != nil {
			logger.Errorf("create sku index err,reason:%s", err.Error())
		}
		if flag {
			parseNum++
		}
		if parseNum%20 == 19 {
			logger.Infof("deal sku:%d", parseNum)
		}
	}
	logger.Infof("end deal sku:%d", parseNum)
}

func createSkuCollection(name string) error {
	err := qdrant.GetInstance().CreateCollection(name, int32(conf.EMB_VEC_SIZE), 2)
	if err != nil {
		common.Errorf(fmt.Sprintf("create collection:%s", name), err)
	}
	return nil
}

func fetchSkuEmbedding(detail detail.SkuDetail) ([]float32, error) {
	//todo:当前这种文本格式不一定是生成emb有效的方式，可能要探索格式和emb model的匹配、微调
	var tmp strings.Builder
	tmp.WriteString("标题：")
	tmp.WriteString(detail.Title)
	if len(detail.CategoryPath) > 0 {
		tmp.WriteString("\r\n")
		tmp.WriteString("类目：")
		for i, cat := range detail.CategoryPath {
			if i > 0 {
				tmp.WriteString("/")
			}
			tmp.WriteString(cat)
		}
	}

	vector, err := modelclient.GetInstance().PassageEmbedding([]string{tmp.String()})
	if err != nil {
		return make([]float32, 0), err
	}
	return vector[0], nil
}

func CreateSkuIndex(detail *detail.SkuDetail) (bool, error) {

	id, err := kvclient.FetchSkuNextId()
	if err != nil {
		return false, err
	}
	key := conf.SKU_KEY_PREFIX + strconv.FormatUint(id, 10)

	cateBytes, err := json.Marshal(detail.CategoryPath)
	if err != nil {
		logger.Errorf("marshal detail.CategoryPath err:%s", err.Error())
		return false, common.Errorf("marshal detail.CategoryPath ", err)
	}

	values := make([]interface{}, 0)
	values = append(values, conf.DETAIL_TITLE_FIELD, detail.Title)
	values = append(values, "cate", string(cateBytes))
	err = kvclient.GetInstance().HMSet(key, values)
	if err != nil {
		return false, common.Errorf(fmt.Sprintf("set key:{%s}", key), err)
	}

	vector, err := fetchSkuEmbedding(*detail)
	if err != nil {
		return false, common.Errorf(fmt.Sprintf("{%s}fetch embedding ", key), err)
	}

	row := qdrant.NewVectorRow(id, vector)
	row.AppendString("id", key)
	row.AppendString(conf.DETAIL_TITLE_FIELD, detail.Title)

	err = qdrant.GetInstance().AddVector(qdrant.SKU_COLLECTION, []*qdrant.VectorRow{row})
	if err != nil {
		return false, common.Errorf(fmt.Sprintf("{%s}create vector index ", key), err)
	}
	return true, nil
}
