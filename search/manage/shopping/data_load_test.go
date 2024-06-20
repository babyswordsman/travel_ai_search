package shopping

import (
	"os"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"

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
	common.Start_client(config)
	defer common.Stop_client()

	wd, _ := os.Getwd()
	println("wd:%s", wd)
	LoadFile("../../../data/shopping_test_data.csv")
}
