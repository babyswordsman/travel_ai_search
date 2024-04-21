package searchengineapi

import (
	"context"
	"encoding/json"
	"testing"
	"travel_ai_search/search/conf"
)

func TestOpenSerpSearch(t *testing.T) {
	path := "/workspace/travel_ai_search/config/conf_local.yaml"
	conf, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err : ", err.Error())
	}
	engine := OpenSerpSearchEngine{
		Engines: conf.OpenSerpSearch.Engines,
		BaseUrl: conf.OpenSerpSearch.Url,
	}

	items, err := engine.Search(context.Background(), conf, "llama3效果怎么样")
	if err != nil {
		t.Error("search err : ", err.Error())
	}
	if len(items) == 0 {
		t.Error("search item is null", len(items))
	}
	buf, _ := json.Marshal(items)
	t.Log(string(buf))

}
