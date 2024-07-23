package walmart

import (
	"encoding/json"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	initclients "travel_ai_search/search/init_clients"
	llmutil "travel_ai_search/search/llm"

	logger "github.com/sirupsen/logrus"
)

func TestRoute(t *testing.T) {
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
	llmutil.InitMemHistoryStoreInstance(5)

	//有没有无麸质的饼干？
	//这些薯片有哪几种口味？
	//你们有售卖进口零食吗？
	query := "有没有适合孩子吃的健康零食"
	tasks, err := route(getTestUser(), SHOPPING_ROOM, SHOPPING_FLOWID, query)
	if err != nil {
		t.Error("err:", err.Error())
		return
	}

	t.Logf("%s size:%d,%v", query, len(tasks), tasks)

	query = "这些薯片有哪几种口味？"
	tasks, err = route(getTestUser(), SHOPPING_ROOM, SHOPPING_FLOWID, query)
	if err != nil {
		t.Error("err:", err.Error())
		return
	}

	t.Logf("%s size:%d,%v", query, len(tasks), tasks)

	query = "今天天气怎么样"
	tasks, err = route(getTestUser(), SHOPPING_ROOM, SHOPPING_FLOWID, query)
	if err != nil {
		t.Error("err:", err.Error())
		return
	}

	t.Logf("%s size:%d,%v", query, len(tasks), tasks)
}

func TestSearchWithIntent(t *testing.T) {
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
	llmutil.InitMemHistoryStoreInstance(5)

	task := WalmartProductListTask{}
	intent := ShoppingIntent{
		IsShopping:       true,
		Category:         "food/chips",
		ProductName:      "chips",
		ProductProps:     make(map[string]string),
		IndependentQuery: "some potato chips",
	}
	resp, err := task.search(&intent)
	if err != nil {
		t.Errorf("search err:%s", err)
		return
	}

	if resp.NumHits < 1 {
		t.Errorf("search num is zero")
		return
	}
	t.Log(resp.NumHits)

	for _, doc := range resp.Hits {

		t.Log(doc.Score, " ", doc.Id, " ", doc.Name)
	}
	buf, _ := json.Marshal(resp.Hits)
	t.Log(string(buf))
}
