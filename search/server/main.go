package main

import (
	"flag"
	"fmt"
	"os"
	"travel_ai_search/search"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/qdrant"

	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

func init_router(r *gin.Engine) {
	//r.GET("/", search.Index)

	manage_route := r.Group("manage")
	{
		manage_route.GET("/init_data", search.InitData)
	}

	chat_route := r.Group("llm")
	{
		chat_route.POST("/chat_prompt", search.ChatPrompt)
		chat_route.POST("/chat", search.Chat)
		chat_route.GET("/chat/stream", search.ChatStream)
		chat_route.GET("/home", search.Home)

	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./conf.yaml", "config path,example:./conf.yaml")

	flag.Parse()

	logger.SetReportCaller(true)

	content, err := os.ReadFile(configPath)
	if err != nil {
		logger.WithField("config", configPath).Error(" read file err:", err)
		fmt.Printf("read file:%s err:%s", configPath, err)
		return
	}
	config := &conf.Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		logger.WithField("config", configPath).Error("parse yaml err:", err)
		fmt.Printf("parse yaml:%s err:%s", configPath, err)
		return
	}

	conf.GlobalConfig = config
	//初始化客户端
	tmpKVClient, err := kvclient.InitClient(config)
	if err != nil {
		logger.Errorf("init kv client:%s err:%s", config.RedisAddr, err)
	}

	defer tmpKVClient.Close()
	logger.WithField("redis", config.RedisAddr).Info("redis init")

	kvclient.InitDetailIdGen()

	tmpVecClient, err := qdrant.InitVectorClient(config)
	if err != nil {
		logger.Errorf("init vector client:%s err:%s", config.QdrantAddr, err)
	}

	defer tmpVecClient.Close()
	logger.WithField("qdrant", config.QdrantAddr).Info("qdrant init")

	tmpModelClient := modelclient.InitModelClient(config)
	defer tmpModelClient.Close()
	logger.WithFields(logger.Fields{"embedding": config.EmbeddingModelHost,
		"reranker": config.RerankerModelHost}).Info("model client init")

	//启动对外服务接口
	r := gin.Default()
	r.LoadHTMLGlob("resource/*.tmpl")
	r.StaticFile("/output.css", "./resource/web/output.css")
	r.StaticFile("/", "./resource/web/index.html")
	init_router(r)
	logger.Info("start gin: ", config.ServerAddr)
	r.Run(config.ServerAddr)
}
