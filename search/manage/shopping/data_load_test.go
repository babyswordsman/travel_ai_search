package shopping

import (
	"encoding/json"
	"os"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/es"
	initclients "travel_ai_search/search/init_clients"
	"travel_ai_search/search/quickwit"
	"travel_ai_search/search/shopping/detail"

	logger "github.com/sirupsen/logrus"
)

func TestParseData(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config
	initclients.Start_client(config)
	defer initclients.Stop_client()

	wd, _ := os.Getwd()
	println("wd:%s", wd)
	LoadFile("../../../data/shopping_test_data.csv")
}

func TestSkuParseData(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	logger.SetReportCaller(true)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config
	initclients.Start_client(config)
	defer initclients.Stop_client()

	root := common.GetProjectPath()
	ParseSkuData(root + "/data/jd.txt")
}

type TestSkuSearchResponse struct {
	NumHists int                  `json:"num_hits"`
	Hits     []detail.SkuDocument `json:"hits"`
}

func TestSearchSku(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config
	initclients.Start_client(config)
	defer initclients.Stop_client()
	req := map[string]any{"query": " third_level:电子产品-手机  OR first_level: IN [电子通讯] OR second_level:电子产品-手机"}
	buf, err := quickwit.GetInstance().Search("sku_a", req)
	if err != nil {
		t.Error("search err", err.Error())
	}
	var resp TestSkuSearchResponse
	err = json.Unmarshal(buf, &resp)
	if err != nil {
		t.Error("search err", err.Error())
	}

	v, err := json.Marshal(resp)
	if err != nil {
		t.Error("search err", err.Error())
	}
	t.Log(string(v))
}

func TestParseWalmartSku(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config
	initclients.Start_client(config)
	defer initclients.Stop_client()

	projectRoot := common.GetProjectPath()
	delInfo, err := es.GetInstance().DeleteIndex("wal_sku")
	if err != nil {
		t.Error("delete index wal_sku err ", err.Error())
		return
	} else {
		t.Logf("delete index %s", delInfo)
	}

	total := LoadWalmartSkuFiles(projectRoot + "/data/walmart")
	t.Logf("add sku :%d", total)
}
