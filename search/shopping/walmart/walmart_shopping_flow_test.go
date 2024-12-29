package walmart

import (
	"encoding/json"
	"testing"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/user"

	"github.com/bytedance/mockey"
	logger "github.com/sirupsen/logrus"
	"github.com/smartystreets/goconvey/convey"
	//"github.com/smartystreets/goconvey/convey"
)

func TestFuncRoute(t *testing.T) {
	logger.SetLevel(logger.DebugLevel)
	logger.SetReportCaller(true)
	path := common.GetTestConfigPath()
	t.Log("config path：", path)
	config, err := conf.ParseConfig(path)
	if err != nil {
		t.Error("parse config err ", err.Error())
		return
	}
	conf.GlobalConfig = config
	llmutil.InitMemHistoryStoreInstance(10)

	mockey.PatchConvey("test funRoute", t, func() {
		mockey.PatchConvey("success", func() {

			funName, args, err := funcRoute(user.EmpytUser, "shopping", "Recommend several kinds of delicious biscuits to me")
			convey.So(err, convey.ShouldBeNil)
			convey.So(funName, convey.ShouldBeIn, []string{"searchProducts", "getProductDetail", "chat"})
			argMap := make(map[string]any)
			err = json.Unmarshal([]byte(args), &argMap)
			convey.So(err, convey.ShouldBeNil)
			switch funName {
			case "searchProducts":
				v, ok := argMap["query"]
				convey.So(ok, convey.ShouldBeTrue)
				convey.Printf("query:%v", v)
			case "getProductDetail":
				v, ok := argMap["product_ids"]
				convey.So(ok, convey.ShouldBeTrue)
				convey.Printf("product_ids:%v", v)
			case "chat":
				v, ok := argMap["input"]
				convey.So(ok, convey.ShouldBeTrue)
				convey.Printf("input:%v", v)
			}
		})
	})
}
