package search

import (
	"encoding/json"
	"html/template"
	"net/http"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/llm"
	"travel_ai_search/search/manage"
	"travel_ai_search/search/user"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gorilla/websocket"
	logger "github.com/sirupsen/logrus"
)

type ChatRequest struct {
	Context string `json:"context"`
	Query   string `json:"query" binding:"required"`
}

func InitData(c *gin.Context) {

	num := manage.ParseData(conf.GlobalConfig, manage.CreateIndex)
	c.JSON(http.StatusOK, gin.H{
		"num": num,
	})
}

func PrintChatPrompt(c *gin.Context) {

	req := ChatRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		logger.Errorf("parse {%s} err:%s", c.GetString(gin.BodyBytesKey), err)
		c.JSON(http.StatusOK, gin.H{
			"prompt": conf.ErrHint,
		})
		return
	}
	logger.WithField("req", c.GetString(gin.BodyBytesKey)).Info("request chat prompt")
	resp := LLMChatPrompt(req.Query)
	c.JSON(http.StatusOK, gin.H{
		"prompt": resp,
	})
}

var chatUpgrader = websocket.Upgrader{}

func ChatStream(ctx *gin.Context) {
	w, r := ctx.Writer, ctx.Request
	c, err := chatUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("chat upgrade:%s", err)
		return
	}

	defer c.Close()

	curUser := user.GetCurUser(ctx)

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			logger.Errorf("chat read msg:%s", err)
			break
		}
		//ping pong close已经由框架代理了
		//暂时只支持单轮
		switch mt {
		case websocket.TextMessage:
			{
				logger.Infof("read msg:%s", message)
				msgData := make(map[string]string)
				err := json.Unmarshal([]byte(message), &msgData)
				if err != nil {
					logger.Errorf("json unmarshal %s err:%s", message, err)
					break
				}

				msgListener := make(chan string, 10)
				go func(user user.User, room string, query string) {
					defer func() {
						if err := recover(); err != nil {
							logger.Errorf("panic err is %s \r\n %s", err, common.GetStack())

							contentResp := llm.ChatStream{
								Type: llm.CHAT_TYPE_TOKENS,
								Body: 0,
							}
							v, _ := json.Marshal(contentResp)
							msgListener <- string(v)

						}
						close(msgListener)
					}()
					tokens := int64(0)
					if conf.GlobalConfig.SparkLLM.IsMock {
						_, tokens = LLMChatStreamMock(room, query, msgListener, llm.LoadChatHistory(curUser.UserId))

					} else {
						_, tokens = LLMChatStream(room, query, msgListener, llm.LoadChatHistory(curUser.UserId))
					}

					contentResp := llm.ChatStream{
						Type: llm.CHAT_TYPE_TOKENS,
						Body: tokens,
					}
					v, _ := json.Marshal(contentResp)
					msgListener <- string(v)
				}(curUser, string(msgData["room"]), string(msgData["input"]))
				for respMsg := range msgListener {
					c.WriteMessage(mt, []byte(respMsg))
				}
				//maybe close
				break
			}
		default:
			{
				logger.Errorf("chat read msg type:%d,msg:%v", mt, message)
				break
			}
		}
	}

}

func Chat(c *gin.Context) {

	req := ChatRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		logger.Errorf("parse {%s} err:%s", c.GetString(gin.BodyBytesKey), err)
		c.JSON(http.StatusOK, gin.H{
			"prompt": conf.ErrHint,
		})
		return
	}
	//logger.WithField("req", c.GetString(gin.BodyBytesKey)).Info("request chat ")
	resp, tokens := LLMChat(req.Query)
	logger.WithField("req", c.GetString(gin.BodyBytesKey)).WithField("chat", resp).Info("request chat")
	c.JSON(http.StatusOK, gin.H{
		"chat":        resp,
		"totalTokens": tokens,
	})
}

func Home(c *gin.Context) {
	c.HTML(http.StatusOK, "chat.tmpl", gin.H{
		"server": template.JSEscapeString(conf.GlobalConfig.ChatAddr),
	})
}

func Index(c *gin.Context) {
	c.Redirect(http.StatusFound, "/index.html")
}
