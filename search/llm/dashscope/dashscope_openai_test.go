package dashscope

import (
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

func TestOpenaiStream(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config

	model := &DashScopeOpenAIModel{
		Room:      "chat",
		ModelName: conf.GlobalConfig.DashScopeLLM.Model,
	}

	msgListener := make(chan string, 10)

	go model.GetChatRes([]llms.ChatMessage{
		llms.SystemChatMessage{Content: "您是一个幽默而又尖酸刻薄的时尚专栏作家"},
		llms.HumanChatMessage{Content: "写个150字的笑话"},
	}, msgListener)
	for str := range msgListener {
		t.Log("chan:", str)
	}

}
