package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
	"travel_ai_search/search"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/kvclient"
	"travel_ai_search/search/llm"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/qdrant"
	"travel_ai_search/search/user"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

func CheckSign(c *gin.Context) {
	logger.Infof("start check sign status")
	session := sessions.DefaultMany(c, conf.GlobalConfig.CookieSession)
	obj := session.Get(conf.GlobalConfig.CookieUser)
	if obj == nil {
		//c.Redirect(http.StatusOK, "/login")
		//return

		//暂时不关联用户，分配一个自定义ID
		id, err := uuid.NewUUID()
		if err != nil {
			logger.Errorf("new uuid err:%s", err)
		}
		user := user.User{
			UserId:   id.String(),
			Lasttime: time.Now(),
		}
		buf, _ := json.Marshal(user)
		session.Set(conf.GlobalConfig.CookieUser, string(buf))
		session.Options(sessions.Options{
			Path: "/",

			MaxAge:   3600 * 24 * 30,
			Secure:   false,
			HttpOnly: true,
		})
		session.Save()
		logger.Infof("add new user:%s", user.UserId)
		c.Next()
	} else {

		buf := obj.(string)
		var user user.User
		err := json.Unmarshal([]byte(buf), &user)
		if err != nil {
			logger.Errorf("Unmarshal {%s} cookie error err:%s", buf, err)
			session.Delete(conf.GlobalConfig.CookieUser)
		}
		logger.Infof("user:%s,lasttime:%s", user.UserId, user.Lasttime.Format("2006-01-02 15:04:05"))
		user.Lasttime = time.Now()
		//session.Save()
		c.Next()
	}

}
func init_router(r *gin.Engine) {
	//r.GET("/", search.Index)

	manage_route := r.Group("manage")
	{
		manage_route.GET("/init_data", search.InitData)
	}

	store := cookie.NewStore([]byte(conf.GlobalConfig.CookieCodeKey))

	// store.Options(sessions.Options{
	// 	Path:     "/",
	// 	HttpOnly: true,
	// 	MaxAge:   3600 * 24 * 30,
	// })

	chat_route := r.Group("/")
	chat_route.Use(sessions.SessionsMany([]string{conf.GlobalConfig.CookieSession}, store), CheckSign)
	{

		chat_route.POST("/chat_prompt", search.PrintChatPrompt)
		chat_route.GET("/chat", search.Index)
		chat_route.GET("/chat/stream", search.ChatStream)
		chat_route.GET("/home", search.Home)
		chat_route.GET("/", search.Blog)
		chat_route.GET("/rag/docs", search.UploadForm)
		chat_route.POST("/rag/upload", search.Upload)
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
	if config.LogLevel != "" {
		lvl, err := logger.ParseLevel(config.LogLevel)
		if err != nil {
			logger.Error("parse log level err:", err.Error())
		}
		logger.SetLevel(lvl)
	}
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
	//llm.InitMemHistoryStoreInstance(5)
	llm.InitKVHistoryStoreInstance(kvclient.GetInstance(), 3)
	//用户历史清理
	llm.GetHistoryStoreInstance().StarCleanTask()

	//启动对外服务接口
	r := gin.Default()
	//r.LoadHTMLGlob("resource/*.tmpl")
	r.LoadHTMLFiles("resource/chat.tmpl", "resource/web/index.html", "resource/web/upload.html")
	r.StaticFile("/output.css", "./resource/web/output.css")
	r.StaticFS("/blog", http.Dir("blog"))
	//r.StaticFile("/index.html", "./resource/web/index.html")
	init_router(r)
	logger.Info("start gin: ", config.ServerAddr)
	if len(config.TlsServerAddr) < 1 {
		r.Run(config.ServerAddr)
	} else {
		e := r.RunTLS(config.TlsServerAddr, config.TlsCertPath, config.TlsCertKeyPath)
		if e != nil {
			logger.Info("start err:", e.Error())
			fmt.Println("start err:", e.Error())
		}
	}

}
