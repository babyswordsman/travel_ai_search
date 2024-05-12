package rewrite

import (
	"encoding/json"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/llm/dashscope"

	"github.com/devinyf/dashscopego/qwen"
	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

func TestQueryRewriting(t *testing.T) {
	// queries := []string{
	// 	"俄罗斯的首都是莫斯科，美国的首都是华盛顿，蒙古国的首都是哪里",
	// 	"纽约和香港哪个物价高",
	// 	"去白马寺玩",
	// }

	logger.SetLevel(logger.DebugLevel)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config

	model := &dashscope.DashScopeModel{
		Room:      "chat",
		ModelName: qwen.QwenTurbo,
	}

	rewritingEngine := &LLMQueryRewritingEngine{
		Model: model,
	}

	results, err := rewritingEngine.Rewrite("俄罗斯的首都是莫斯科，美国的首都是华盛顿，蒙古国的首都是哪里", make([]llms.ChatMessage, 0))

	if err != nil {
		t.Error("rewrite query err", err.Error())
	}

	for ind, v := range results {
		t.Log("改写后的查询", ind, ":", v)
	}

	results, err = rewritingEngine.Rewrite("纽约和香港哪个物价高", make([]llms.ChatMessage, 0))

	if err != nil {
		t.Error("rewrite query err", err.Error())
	}

	for ind, v := range results {
		t.Log("改写后的查询", ind, ":", v)
	}
}

func TestJsonEscape(t *testing.T) {
	str := "{\"rewriting_query\":[\"蒙古国的首都\"], \"is_need_rewrite\":true}"
	result := make(map[string]any)
	json.Unmarshal([]byte(str), &result)
	value, ok := result["rewriting_query"]

	if ok {
		rewritingQueries, ok := value.([]any)
		if ok {
			queries := make([]string, 0)
			for _, item := range rewritingQueries {
				if q, ok := item.(string); ok {
					queries = append(queries, q)
				}
			}
			t.Log(queries)
		}

	}
	t.Log(str)
}
