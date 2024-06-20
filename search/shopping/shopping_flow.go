package shopping

import (
	"context"
	"travel_ai_search/search/conf"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/vectorstores"
)

type ShoppingEngine struct {
}

type ShoppingFlowLogHandler struct {
	callbacks.LogHandler
}

func NewDashScopeModel() (*openai.LLM, error) {
	opts := make([]openai.Option, 0)
	opts = append(opts, openai.WithBaseURL(conf.GlobalConfig.DashScopeLLM.OpenaiUrl))
	opts = append(opts, openai.WithModel(conf.GlobalConfig.DashScopeLLM.Model))
	opts = append(opts, openai.WithToken(conf.GlobalConfig.DashScopeLLM.Key))
	opts = append(opts, openai.WithCallback(ShoppingFlowLogHandler{}))
	llm, err := openai.New(opts...)
	return llm, err
}

func (engine *ShoppingEngine) Flow(query string) (string, error) {
	ctx := context.Background()
	llm, err := NewDashScopeModel()

	if err != nil {
		logger.Errorf("create llm client err:%s", err.Error())
		return "", err
	}

	combinedStuffQAChain := chains.LoadStuffQA(llm)
	combinedQuestionGeneratorChain := chains.LoadCondenseQuestionGenerator(llm)

	qdrantStore := NewVector()

	chain := chains.NewConversationalRetrievalQA(
		combinedStuffQAChain,
		combinedQuestionGeneratorChain,
		vectorstores.ToRetriever(qdrantStore, int(conf.GlobalConfig.MaxCandidates)),
		memory.NewConversationBuffer(memory.WithReturnMessages(true)),
	)
	result, err := chains.Run(ctx, chain, query)
	logger.Infof("query:%s,result:%s", query, result)
	return result, err
}
