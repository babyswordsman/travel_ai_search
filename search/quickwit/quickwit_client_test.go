package quickwit

import (
	"io"
	"os"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
)

func TestCreateIndex(t *testing.T) {
	indexFile, err := os.Open(common.GetProjectPath() + "/config/quickwit_sku.yaml")
	if err != nil {
		t.Error(err.Error())
	}

	defer indexFile.Close()
	indexConfigBuf, err := io.ReadAll(indexFile)
	if err != nil {
		t.Error("read index config file err ", err.Error())
		return
	}
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config
	client := InitQuickwitClient(config)
	err = client.CreateIndex(string(indexConfigBuf))
	if err != nil {
		t.Error("create index err ", err.Error())
		return
	}
}
