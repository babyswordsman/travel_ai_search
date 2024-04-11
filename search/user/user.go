package user

import (
	"time"
	"travel_ai_search/search/conf"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type User struct {
	UserId   string
	UserName string
	//最后活跃时间
	Lasttime time.Time
}

var EmpytUser User = User{
	UserId: "_",
}

func GetCurUser(c *gin.Context) User {
	session := sessions.DefaultMany(c, conf.GlobalConfig.CookieSession)
	obj := session.Get(conf.GlobalConfig.CookieUser)
	if obj == nil {
		return EmpytUser
	} else {
		return obj.(User)
	}
}
