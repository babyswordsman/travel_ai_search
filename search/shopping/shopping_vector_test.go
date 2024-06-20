package shopping

import (
	"context"
	"encoding/json"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"

	logger "github.com/sirupsen/logrus"
)

func TestVectorStore(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config
	common.Start_client(config)
	defer common.Stop_client()

	store := NewVector()
	docs, err := store.SimilaritySearch(context.Background(), "被子", 10)
	if err != nil {
		t.Error(err.Error())
	}
	v, _ := json.Marshal(docs)
	t.Log(string(v))
}
