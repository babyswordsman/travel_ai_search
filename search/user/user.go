package user

import (
	"encoding/json"
	"time"
	"travel_ai_search/search/conf"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
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

		buf := obj.(string)
		var user User
		err := json.Unmarshal([]byte(buf), &user)
		if err != nil {
			logger.Errorf("Unmarshal {%s} cookie error err:%s", buf, err)
			session.Delete(conf.GlobalConfig.CookieUser)
			return EmpytUser
		}
		logger.Infof("cur user:%s", user.UserId)
		return user
	}
}
