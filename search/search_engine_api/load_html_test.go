package searchengineapi

import (
	"testing"

	logger "github.com/sirupsen/logrus"
)

func TestLoadHtml(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	documents, err := LoadHtml("https://www.163.com/dy/article/J0NR1TKH05299A13.html?clickfrom=w_sports")
	if err != nil {
		t.Errorf("load html error:%s", err)
	}

	for i, doc := range documents {
		t.Logf("%d:score:%f,meta:%v", i, doc.Score, doc.Metadata)
		t.Log(doc.PageContent)
	}
}
