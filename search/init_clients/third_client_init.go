package initclients

import (
	"travel_ai_search/search/conf"
	"travel_ai_search/search/es"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/qdrant"
	"travel_ai_search/search/quickwit"

	logger "github.com/sirupsen/logrus"
)

func Start_client(config *conf.Config) {
	//初始化客户端
	_, err := kvclient.InitClient(config)
	if err != nil {
		logger.Errorf("init kv client:%s err:%s", config.RedisAddr, err)
	}

	logger.WithField("redis", config.RedisAddr).Info("redis init")

	//启动ID生成器
	kvclient.StartIdGen()

	_, err = qdrant.InitVectorClient(config)
	if err != nil {
		logger.Errorf("init vector client:%s err:%s", config.QdrantAddr, err)
	}

	logger.WithField("qdrant", config.QdrantAddr).Info("qdrant init")

	modelclient.InitModelClient(config)

	quickwit.InitQuickwitClient(config)

	_,err = es.InitESClient(config)
	if err != nil {
		logger.Errorf("init es client:%v err:%s", config.ESUrl, err)
	}
}

func Stop_client() {
	kvclient.GetInstance().Close()
	qdrant.GetInstance().Close()
	modelclient.GetInstance().Close()
	quickwit.GetInstance().Close()
}
