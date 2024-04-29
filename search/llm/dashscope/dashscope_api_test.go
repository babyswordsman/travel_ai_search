package dashscope

import (
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"

	"github.com/devinyf/dashscopego/qwen"
	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

func TestStream(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config

	model := &DashScopeModel{
		Room:      "chat",
		ModelName: qwen.QwenTurbo,
	}

	msgListener := make(chan string, 10)

	go model.GetChatRes([]llms.ChatMessage{
		llms.HumanChatMessage{Content: "写个150字的笑话"},
	}, msgListener)
	for str := range msgListener {
		t.Log("chan:", str)
	}
}
