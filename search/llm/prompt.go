package llm

import (
	"strconv"
	"strings"
	searchengineapi "travel_ai_search/search/search_engine_api"

	logger "github.com/sirupsen/logrus"
)

type Prompt interface {
	GenPrompt(candidates []searchengineapi.SearchItem) (string, error)
}

type TravelPrompt struct {
	MaxLength    int
	PromptPrefix string
}

func (prompt *TravelPrompt) GenPrompt(candidates []searchengineapi.SearchItem) (string, error) {
	//todo:截断
	buf := strings.Builder{}
	buf.WriteString(prompt.PromptPrefix)
	buf.WriteString("\r\n")
	remain := prompt.MaxLength
	for ind, item := range candidates {
		titleLen := len(item.Title)
		contentLen := len(item.Snippet)
		if remain-titleLen > 0 {
			buf.WriteString("方案" + strconv.Itoa(ind+1) + ":")
			buf.WriteString("\r\n")
			buf.WriteString(item.Title)
			remain = remain - titleLen
		} else {
			break
		}

		if remain-contentLen > 0 {
			buf.WriteString("\r\n")
			buf.WriteString(item.Snippet)
			buf.WriteString("\r\n")
			remain = remain - contentLen
		} else {
			break
		}

	}
	buf.WriteString("\r\n")
	logger.Info(buf.String())
	return buf.String(), nil
}

type ChatPrompt struct {
	MaxLength    int
	PromptPrefix string
}

func (prompt *ChatPrompt) GenPrompt(candidates []searchengineapi.SearchItem) (string, error) {
	//todo:截断
	buf := strings.Builder{}
	buf.WriteString(prompt.PromptPrefix)
	buf.WriteString("\r\n")
	remain := prompt.MaxLength
	for ind, item := range candidates {
		titleLen := len(item.Title)
		contentLen := len(item.Snippet)
		if remain-titleLen > 0 {
			buf.WriteString("资料" + strconv.Itoa(ind+1) + ":")
			buf.WriteString("\r\n")
			buf.WriteString(item.Title)
			remain = remain - titleLen
		} else {
			break
		}

		if remain-contentLen > 0 {
			buf.WriteString("\r\n")
			buf.WriteString(item.Snippet)
			buf.WriteString("\r\n")
			remain = remain - contentLen
		} else {
			break
		}

	}
	buf.WriteString("\r\n")
	logger.Info(buf.String())
	return buf.String(), nil
}
