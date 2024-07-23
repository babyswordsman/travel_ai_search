package shopping

import (
	"travel_ai_search/search/llm"
	"travel_ai_search/search/user"
)

type TaskEngine interface {
	Run(curUser user.User, room, input string) (llm.TaskOutputType, any, error)
	Status() int //1表示成功，其他表示识别
	FormatOutput() string
}
