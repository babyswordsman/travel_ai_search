package walmart

import (
	"context"
	"travel_ai_search/search/conf"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/user"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/prompts"
)

type WalmartChatTask struct {
	output string
	status int
}

func (task *WalmartChatTask) Run(curUser user.User, room, input string) (llmutil.TaskOutputType, any, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()
	logEntry := logger.WithField("uid", curUser.UserId)
	if err != nil {
		logEntry.Errorf("create llm client err:%s", err.Error())
		return llmutil.CHAT_TYPE_MSG, "", err
	}

	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		conf.GlobalConfig.PromptTemplate.WalmartChat,
		[]string{"context", "question"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := chains.NewLLMChain(llm, prompt)
	inputs := make(map[string]any)
	inputs["question"] = input
	inputs["context"] = conversationContext(curUser, room)

	str, err := doLLMRespText(logEntry, llmChain, inputs, ctx)
	if err != nil {
		task.status = llmutil.RUN_DONE
		task.output = str
	} else {
		task.status = llmutil.RUN_ERR
	}
	return llmutil.CHAT_TYPE_MSG, str, err
}

func (task *WalmartChatTask) Status() int {

	return task.status
}
func (task *WalmartChatTask) FormatOutput() string {
	return task.output
}
