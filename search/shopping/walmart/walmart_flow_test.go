package walmart

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"
	"travel_ai_search/search/common"
	"travel_ai_search/search/conf"
	"travel_ai_search/search/es"
	initclients "travel_ai_search/search/init_clients"
	"travel_ai_search/search/llm"
	llmutil "travel_ai_search/search/llm"
	"travel_ai_search/search/modelclient"
	"travel_ai_search/search/shopping/detail"
	"travel_ai_search/search/user"

	logger "github.com/sirupsen/logrus"
)

func TestShoppingFlow(t *testing.T) {
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
	initclients.Start_client(config)
	defer initclients.Stop_client()
	llm.InitMemHistoryStoreInstance(5)
	engine := ShoppingEngine{}
	_, res, err := engine.Flow(getTestUser(), SHOPPING_ROOM, "I want to buy some crispy cookies")
	if err != nil {
		t.Error("flow err :", err.Error())
	} else {
		t.Log("result:", res)
	}
}

func TestSearchWithShoppingIntent(t *testing.T) {
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
	initclients.Start_client(config)
	defer initclients.Stop_client()
	llm.InitMemHistoryStoreInstance(5)
	engine := ShoppingEngine{}
	shoppingIntentJson := `{"完整信息": "用户在寻找适合夏天的床单。", "问题1": 1, "问题2": "床上用品-床单", "问题3": "夏季透气床单", "问题4": {"材质": "棉质", "颜色": "白色/浅色", "透气性": "良好", "尺寸": "Queen/双人/King/加大号", "季节": "夏季"}}`
	intentMap := make(map[string]any)
	json.Unmarshal([]byte(shoppingIntentJson), &intentMap)
	shoppingIntent := engine.parseIntent(intentMap)
	resp, err := engine.search(&shoppingIntent)
	if err != nil {
		t.Errorf("search err:%s", err)
	}

	if resp.NumHits < 1 {
		t.Errorf("search num is zero")
		return
	}
	t.Log(resp.NumHits)

	for _, doc := range resp.Hits {

		t.Log(doc.Score, " ", doc.Id, " ", doc.Name)
	}
	buf, _ := json.Marshal(resp.Hits)
	t.Log(string(buf))
}

func TestShoppingAdvice(t *testing.T) {
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
	initclients.Start_client(config)
	defer initclients.Stop_client()
	llm.InitMemHistoryStoreInstance(5)
	engine := ShoppingEngine{}

	queryHits := ` [{"timestamp":1720002190218,"store_name":"","product_main_pic":"","product_name":"Apple/苹果 iPhone 15 (A3092) 256GB 黑色 支持移动联通电信5G 双卡双待手机","brand":"Apple","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":5999,"extended_props":"{\"CPU型号\":\"A16\",\"三防标准\":\"IP68\",\"充电功率\":\"以官网信息为准\",\"后摄主像素\":\"4800万像素\",\"商品名称\":\"AppleiPhone15\",\"商品编号\":\"100066896338\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED直屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"256GB\",\"机身色系\":\"黑色系\",\"机身颜色\":\"黑色\",\"泰尔电竞评级\":\"否\",\"运行内存\":\"未公布\",\"风格\":\"科技，商务\"}","comment_summary":"{\"commentCountStr\":\"200万+\",\"generalCountStr\":\"5500+\",\"goodCountStr\":\"36万+\",\"goodRateShow\":95,\"hotCommentTagStatistics\":[{\"name\":\"手感一流\"},{\"name\":\"漂亮大方\"},{\"name\":\"时尚简约\"},{\"name\":\"色彩饱满\"},{\"name\":\"性能一流\"},{\"name\":\"散热性佳\"},{\"name\":\"操作简单\"},{\"name\":\"品质优良\"},{\"name\":\"送给TA\"},{\"name\":\"解锁迅速\"},{\"name\":\"畅快办公\"},{\"name\":\"游戏上分\"},{\"name\":\"按键灵敏\"}],\"imageListCount\":500,\"poorCountStr\":\"1.2万+\",\"videoCountStr\":\"3400+\"}"},{"timestamp":1720002190218,"store_name":"","product_main_pic":"","product_name":"Apple/苹果 iPhone 15 (A3092) 256GB 黑色 支持移动联通电信5G 双卡双待手机","brand":"Apple","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":5999,"extended_props":"{\"CPU型号\":\"A16\",\"三防标准\":\"IP68\",\"充电功率\":\"以官网信息为准\",\"后摄主像素\":\"4800万像素\",\"商品名称\":\"AppleiPhone15\",\"商品编号\":\"100066896338\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED直屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"256GB\",\"机身色系\":\"黑色系\",\"机身颜色\":\"黑色\",\"泰尔电竞评级\":\"否\",\"运行内存\":\"未公布\",\"风格\":\"科技，商务\"}","comment_summary":"{\"commentCountStr\":\"200万+\",\"generalCountStr\":\"5500+\",\"goodCountStr\":\"36万+\",\"goodRateShow\":95,\"hotCommentTagStatistics\":[{\"name\":\"手感一流\"},{\"name\":\"漂亮大方\"},{\"name\":\"时尚简约\"},{\"name\":\"色彩饱满\"},{\"name\":\"性能一流\"},{\"name\":\"散热性佳\"},{\"name\":\"操作简单\"},{\"name\":\"品质优良\"},{\"name\":\"送给TA\"},{\"name\":\"解锁迅速\"},{\"name\":\"畅快办公\"},{\"name\":\"游戏上分\"},{\"name\":\"按键灵敏\"}],\"imageListCount\":500,\"poorCountStr\":\"1.2万+\",\"videoCountStr\":\"3400+\"}"},{"timestamp":1720002190218,"store_name":"","product_main_pic":"","product_name":"Apple/苹果 iPhone 15 (A3092) 256GB 黑色 支持移动联通电信5G 双卡双待手机","brand":"Apple","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":5999,"extended_props":"{\"CPU型号\":\"A16\",\"三防标准\":\"IP68\",\"充电功率\":\"以官网信息为准\",\"后摄主像素\":\"4800万像素\",\"商品名称\":\"AppleiPhone15\",\"商品编号\":\"100066896338\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED直屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"256GB\",\"机身色系\":\"黑色系\",\"机身颜色\":\"黑色\",\"泰尔电竞评级\":\"否\",\"运行内存\":\"未公布\",\"风格\":\"科技，商务\"}","comment_summary":"{\"commentCountStr\":\"200万+\",\"generalCountStr\":\"5500+\",\"goodCountStr\":\"36万+\",\"goodRateShow\":95,\"hotCommentTagStatistics\":[{\"name\":\"手感一流\"},{\"name\":\"漂亮大方\"},{\"name\":\"时尚简约\"},{\"name\":\"色彩饱满\"},{\"name\":\"性能一流\"},{\"name\":\"散热性佳\"},{\"name\":\"操作简单\"},{\"name\":\"品质优良\"},{\"name\":\"送给TA\"},{\"name\":\"解锁迅速\"},{\"name\":\"畅快办公\"},{\"name\":\"游戏上分\"},{\"name\":\"按键灵敏\"}],\"imageListCount\":500,\"poorCountStr\":\"1.2万+\",\"videoCountStr\":\"3400+\"}"},{"timestamp":1720002190218,"store_name":"","product_main_pic":"","product_name":"Apple/苹果 iPhone 15 (A3092) 256GB 黑色 支持移动联通电信5G 双卡双待手机","brand":"Apple","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":5999,"extended_props":"{\"CPU型号\":\"A16\",\"三防标准\":\"IP68\",\"充电功率\":\"以官网信息为准\",\"后摄主像素\":\"4800万像素\",\"商品名称\":\"AppleiPhone15\",\"商品编号\":\"100066896338\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED直屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"256GB\",\"机身色系\":\"黑色系\",\"机身颜色\":\"黑色\",\"泰尔电竞评级\":\"否\",\"运行内存\":\"未公布\",\"风格\":\"科技，商务\"}","comment_summary":"{\"commentCountStr\":\"200万+\",\"generalCountStr\":\"5500+\",\"goodCountStr\":\"36万+\",\"goodRateShow\":95,\"hotCommentTagStatistics\":[{\"name\":\"手感一流\"},{\"name\":\"漂亮大方\"},{\"name\":\"时尚简约\"},{\"name\":\"色彩饱满\"},{\"name\":\"性能一流\"},{\"name\":\"散热性佳\"},{\"name\":\"操作简单\"},{\"name\":\"品质优良\"},{\"name\":\"送给TA\"},{\"name\":\"解锁迅速\"},{\"name\":\"畅快办公\"},{\"name\":\"游戏上分\"},{\"name\":\"按键灵敏\"}],\"imageListCount\":500,\"poorCountStr\":\"1.2万+\",\"videoCountStr\":\"3400+\"}"},{"timestamp":1720002190218,"store_name":"","product_main_pic":"","product_name":"Apple/苹果 iPhone 15 (A3092) 256GB 黑色 支持移动联通电信5G 双卡双待手机","brand":"Apple","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":5999,"extended_props":"{\"CPU型号\":\"A16\",\"三防标准\":\"IP68\",\"充电功率\":\"以官网信息为准\",\"后摄主像素\":\"4800万像素\",\"商品名称\":\"AppleiPhone15\",\"商品编号\":\"100066896338\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED直屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"256GB\",\"机身色系\":\"黑色系\",\"机身颜色\":\"黑色\",\"泰尔电竞评级\":\"否\",\"运行内存\":\"未公布\",\"风格\":\"科技，商务\"}","comment_summary":"{\"commentCountStr\":\"200万+\",\"generalCountStr\":\"5500+\",\"goodCountStr\":\"36万+\",\"goodRateShow\":95,\"hotCommentTagStatistics\":[{\"name\":\"手感一流\"},{\"name\":\"漂亮大方\"},{\"name\":\"时尚简约\"},{\"name\":\"色彩饱满\"},{\"name\":\"性能一流\"},{\"name\":\"散热性佳\"},{\"name\":\"操作简单\"},{\"name\":\"品质优良\"},{\"name\":\"送给TA\"},{\"name\":\"解锁迅速\"},{\"name\":\"畅快办公\"},{\"name\":\"游戏上分\"},{\"name\":\"按键灵敏\"}],\"imageListCount\":500,\"poorCountStr\":\"1.2万+\",\"videoCountStr\":\"3400+\"}"},{"timestamp":1720002190218,"store_name":"","product_main_pic":"","product_name":"Apple/苹果 iPhone 15 (A3092) 256GB 黑色 支持移动联通电信5G 双卡双待手机","brand":"Apple","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":5999,"extended_props":"{\"CPU型号\":\"A16\",\"三防标准\":\"IP68\",\"充电功率\":\"以官网信息为准\",\"后摄主像素\":\"4800万像素\",\"商品名称\":\"AppleiPhone15\",\"商品编号\":\"100066896338\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED直屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"256GB\",\"机身色系\":\"黑色系\",\"机身颜色\":\"黑色\",\"泰尔电竞评级\":\"否\",\"运行内存\":\"未公布\",\"风格\":\"科技，商务\"}","comment_summary":"{\"commentCountStr\":\"200万+\",\"generalCountStr\":\"5500+\",\"goodCountStr\":\"36万+\",\"goodRateShow\":95,\"hotCommentTagStatistics\":[{\"name\":\"手感一流\"},{\"name\":\"漂亮大方\"},{\"name\":\"时尚简约\"},{\"name\":\"色彩饱满\"},{\"name\":\"性能一流\"},{\"name\":\"散热性佳\"},{\"name\":\"操作简单\"},{\"name\":\"品质优良\"},{\"name\":\"送给TA\"},{\"name\":\"解锁迅速\"},{\"name\":\"畅快办公\"},{\"name\":\"游戏上分\"},{\"name\":\"按键灵敏\"}],\"imageListCount\":500,\"poorCountStr\":\"1.2万+\",\"videoCountStr\":\"3400+\"}"}]`
	docs := make([]*detail.WalmartSkuResp, 0)
	err = json.Unmarshal([]byte(queryHits), &docs)
	if err != nil {
		t.Errorf("unmarshal err:%s", err)
	}

	advice, err := engine.askUser("用户在寻找适合拍照的手机。", getTestUser(), docs)
	if err != nil {
		t.Errorf("ask user err:%s", err)
	}
	//"您好，为了更好地帮您挑选拍照功能出色的手机，我想了解一下您的具体需求。您更偏爱哪个屏幕分辨率？后摄主像素一般需要达到多少？还有，对CPU性能有什么特殊要求吗？另外，机身颜色和设计风格也是考虑因素之一吧？请告诉我这些详细信息，谢谢！"
	t.Log(advice)
}

func getTestUser() user.User {
	user := user.User{
		UserId:   "u-1",
		UserName: "test1",
		Lasttime: time.Now(),
	}
	return user
}

func TestShoppingConversationHistory(t *testing.T) {
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
	initclients.Start_client(config)
	defer initclients.Stop_client()
	llm.InitMemHistoryStoreInstance(5)
	engine := ShoppingEngine{}
	query := "推荐个适合拍照的床单"
	resp := "您好，为了更好地帮您挑选拍照功能出色的手机，我想了解一下您的具体需求。您更偏爱哪个屏幕分辨率？后摄主像素一般需要达到多少？还有，对CPU性能有什么特殊要求吗？另外，机身颜色和设计风格也是考虑因素之一吧？请告诉我这些详细信息，谢谢！"
	engine.saveChatHistory(getTestUser(), SHOPPING_ROOM, query, resp)

	content := engine.conversationContext(getTestUser(), SHOPPING_ROOM)

	isContain1 := strings.Contains(content, llmutil.ROLE_USER+":"+query)
	isContain2 := strings.Contains(content, llmutil.ROLE_ASSISTANT+":"+resp)

	if !isContain1 || !isContain2 {
		t.Errorf("content err:%s", content)
	}
	t.Log(content)
}

func TestShoppingFlowWithAdvice(t *testing.T) {
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
	initclients.Start_client(config)
	defer initclients.Stop_client()
	llm.InitMemHistoryStoreInstance(5)
	engine := ShoppingEngine{}
	query := "推荐个适合拍照的手机"
	resp := "您好，为了更好地帮您挑选拍照功能出色的手机，我想了解一下您的具体需求。您更偏爱哪个屏幕分辨率？后摄主像素一般需要达到多少？还有，对CPU性能有什么特殊要求吗？另外，机身颜色和设计风格也是考虑因素之一吧？请告诉我这些详细信息，谢谢！"
	engine.saveChatHistory(getTestUser(), SHOPPING_ROOM, query, resp)

	_, res, err := engine.Flow(getTestUser(), SHOPPING_ROOM, "存储大一点，像素高的")
	if err != nil {
		t.Error("flow err :", err.Error())
	} else {
		t.Log("result:", res)
	}
	//{"完整信息":"用户希望购买一款存储空间大、像素高的手机。", "问题1":1, "问题2":"数码产品-手机-摄影摄像", "问题3":"可能的商品名称：iPhone 12 Pro Max、Samsung Galaxy S21 Ultra、Google Pixel 6 Pro等（具体取决于品牌和配置）", "问题4": {"内存容量": "大", "主摄像头像素": "高"}}
}

func TestShoppingRecommend(t *testing.T) {
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
	initclients.Start_client(config)
	defer initclients.Stop_client()
	llm.InitMemHistoryStoreInstance(5)
	engine := ShoppingEngine{}
	query := "推荐个适合拍照的手机"
	resp := "您好，为了更好地帮您挑选拍照功能出色的手机，我想了解一下您的具体需求。您更偏爱哪个屏幕分辨率？后摄主像素一般需要达到多少？还有，对CPU性能有什么特殊要求吗？另外，机身颜色和设计风格也是考虑因素之一吧？请告诉我这些详细信息，谢谢！"
	engine.saveChatHistory(getTestUser(), SHOPPING_ROOM, query, resp)

	queryHits := `[{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"荣耀X50i 一亿像素超清影像 轻羽灵感设计 6.7英寸超窄边护眼全视屏 5G手机 8GB+256GB 杨柳风","brand":"荣耀（HONOR）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":1499,"extended_props":"{\"CPU型号\":\"天玑6020\",\"三防标准\":\"不支持防水\",\"充电功率\":\"26-49W\",\"后摄主像素\":\"1亿像素\",\"商品名称\":\"荣耀荣耀手机\",\"商品编号\":\"100056919535\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"LTPS\",\"机身内存\":\"256GB\",\"机身色系\":\"浅蓝色系\",\"机身颜色\":\"杨柳风\",\"运行内存\":\"8GB\",\"风格\":\"时尚，炫彩\"}","comment_summary":"{\"commentCountStr\":\"20万+\",\"generalCountStr\":\"600+\",\"goodCountStr\":\"6.5万+\",\"goodRateShow\":97,\"hotCommentTagStatistics\":[{\"name\":\"时尚美观\"},{\"name\":\"高端大气\"},{\"name\":\"颜色绚丽\"},{\"name\":\"性能一流\"},{\"name\":\"操作简便\"},{\"name\":\"手感一流\"},{\"name\":\"散热性佳\"},{\"name\":\"送给TA\"},{\"name\":\"解锁迅速\"},{\"name\":\"高效学习\"},{\"name\":\"防摔耐用\"},{\"name\":\"优质耐用\"}],\"imageListCount\":500,\"poorCountStr\":\"800+\",\"videoCountStr\":\"500+\"}"},{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"荣耀Magic V Flip 小折叠屏 4.0英寸大外屏 单反级写真相机 青海湖电池 5G AI 拍照手机 12+256 山茶白","brand":"荣耀（HONOR）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":4999,"extended_props":"{\"CPU型号\":\"第一代骁龙8+\",\"充电功率\":\"50-79W\",\"后摄主像素\":\"5000万像素\",\"商品名称\":\"荣耀Magic\",\"商品编号\":\"100118707150\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"折叠屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"256GB\",\"机身色系\":\"白色系\",\"机身颜色\":\"山茶白\",\"运行内存\":\"12GB\"}","comment_summary":"{\"commentCountStr\":\"2000+\",\"generalCountStr\":\"3\",\"goodCountStr\":\"400+\",\"goodRateShow\":98,\"hotCommentTagStatistics\":[{\"name\":\"时尚美观\"},{\"name\":\"小巧精致\"},{\"name\":\"手感一流\"},{\"name\":\"精美漂亮\"}],\"imageListCount\":387,\"poorCountStr\":\"4\",\"videoCountStr\":\"10+\"}"},{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"荣耀Magic V Flip 小折叠屏 4.0英寸大外屏 单反级写真相机 青海湖电池 5G AI 拍照手机 12+512 山茶白","brand":"荣耀（HONOR）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":5499,"extended_props":"{\"CPU型号\":\"第一代骁龙8+\",\"充电功率\":\"50-79W\",\"后摄主像素\":\"5000万像素\",\"商品名称\":\"荣耀Magic\",\"商品编号\":\"100118707118\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"折叠屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"512GB\",\"机身色系\":\"白色系\",\"机身颜色\":\"山茶白\",\"运行内存\":\"12GB\"}","comment_summary":"{\"commentCountStr\":\"2000+\",\"generalCountStr\":\"3\",\"goodCountStr\":\"400+\",\"goodRateShow\":98,\"hotCommentTagStatistics\":[{\"name\":\"时尚美观\"},{\"name\":\"新颖别致\"},{\"name\":\"手感一流\"},{\"name\":\"性能一流\"},{\"name\":\"触感舒适\"},{\"name\":\"线条流畅\"},{\"name\":\"轻重合适\"}],\"imageListCount\":416,\"poorCountStr\":\"4\",\"videoCountStr\":\"10+\"}"},{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"荣耀Magic6 至臻版 单反级超动态鹰眼相机 荣耀金刚巨犀玻璃 荣耀鸿燕通信 16+512 5G AI手机 墨岩黑","brand":"荣耀（HONOR）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":6999,"extended_props":"{\"CPU型号\":\"第三代骁龙8\",\"三防标准\":\"IP68\",\"充电功率\":\"80-119W\",\"后摄主像素\":\"5000万像素\",\"商品名称\":\"荣耀Magic6\",\"商品编号\":\"100091536885\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED曲面屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"512GB\",\"机身色系\":\"黑色系\",\"机身颜色\":\"墨岩黑\",\"特征特质\":\"AI大模型，卫星通话，LTPO屏幕，高频PWM调光，潜望式长焦，防水防尘，无线充电，NFC，红外遥控，大底主摄\",\"运行内存\":\"16GB\",\"风格\":\"商务，科技\"}","comment_summary":"{\"commentCountStr\":\"5000+\",\"generalCountStr\":\"20+\",\"goodCountStr\":\"1900+\",\"goodRateShow\":97,\"hotCommentTagStatistics\":[{\"name\":\"时尚美观\"},{\"name\":\"时尚简约\"},{\"name\":\"手感一流\"},{\"name\":\"功能强大\"},{\"name\":\"质感俱佳\"},{\"name\":\"操作简便\"},{\"name\":\"优质耐用\"}],\"imageListCount\":500,\"poorCountStr\":\"30+\",\"videoCountStr\":\"50+\"}"},{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"荣耀Magic6 至臻版 单反级超动态鹰眼相机 荣耀金刚巨犀玻璃 荣耀鸿燕通信 16+512 5G AI手机 天穹紫","brand":"荣耀（HONOR）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":6999,"extended_props":"{\"CPU型号\":\"第三代骁龙8\",\"三防标准\":\"IP68\",\"充电功率\":\"80-119W\",\"后摄主像素\":\"5000万像素\",\"商品名称\":\"荣耀Magic6\",\"商品编号\":\"100091536857\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED曲面屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"512GB\",\"机身色系\":\"紫色系\",\"机身颜色\":\"天穹紫\",\"特征特质\":\"AI大模型，卫星通话，LTPO屏幕，高频PWM调光，潜望式长焦，防水防尘，无线充电，NFC，红外遥控，大底主摄\",\"运行内存\":\"16GB\",\"风格\":\"商务，科技\"}","comment_summary":"{\"commentCountStr\":\"5000+\",\"generalCountStr\":\"20+\",\"goodCountStr\":\"1900+\",\"goodRateShow\":97,\"hotCommentTagStatistics\":[{\"name\":\"时尚美观\"},{\"name\":\"线条流畅\"},{\"name\":\"手感一流\"},{\"name\":\"低调奢华\"},{\"name\":\"拍摄功能强\"},{\"name\":\"操作简单\"},{\"name\":\"防摔耐磨\"},{\"name\":\"散热性佳\"}],\"imageListCount\":500,\"poorCountStr\":\"30+\",\"videoCountStr\":\"50+\"}"},{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"荣耀Magic6 至臻版 单反级超动态鹰眼相机 荣耀金刚巨犀玻璃 荣耀鸿燕通信 16GB+1TB 5G AI手机 天穹紫","brand":"荣耀（HONOR）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":7699,"extended_props":"{\"CPU型号\":\"第三代骁龙8\",\"三防标准\":\"IP68\",\"充电功率\":\"80-119W\",\"后摄主像素\":\"5000万像素\",\"商品名称\":\"荣耀Magic6\",\"商品编号\":\"100091536883\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED曲面屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"1TB\",\"机身色系\":\"紫色系\",\"机身颜色\":\"天穹紫\",\"特征特质\":\"AI大模型，卫星通话，LTPO屏幕，高频PWM调光，潜望式长焦，防水防尘，无线充电，NFC，红外遥控，大底主摄\",\"运行内存\":\"16GB\",\"风格\":\"商务，科技\"}","comment_summary":"{\"commentCountStr\":\"5000+\",\"generalCountStr\":\"20+\",\"goodCountStr\":\"1900+\",\"goodRateShow\":97,\"hotCommentTagStatistics\":[{\"name\":\"时尚美观\"},{\"name\":\"质感俱佳\"},{\"name\":\"颜色绚丽\"},{\"name\":\"拍照一流\"},{\"name\":\"安全不伤手\"},{\"name\":\"手感一流\"},{\"name\":\"送给TA\"}],\"imageListCount\":500,\"poorCountStr\":\"30+\",\"videoCountStr\":\"50+\"}"},{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"HUAWEI Pura 70 雪域白 12GB+1TB 超高速风驰闪拍 第二代昆仑玻璃 双超级快充 华为P70智能手机","brand":"华为（HUAWEI）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":6999,"extended_props":"{\"CPU型号\":\"未公布\",\"三防标准\":\"IP68\",\"充电功率\":\"50-79W\",\"后摄主像素\":\"5000万像素\",\"商品名称\":\"华为Pura\",\"商品编号\":\"100107613720\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED直屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"1TB\",\"机身色系\":\"白色系\",\"机身颜色\":\"雪域白\",\"特征特质\":\"AI大模型，轻薄，LTPO屏幕，高频PWM调光，全焦段影像，潜望式长焦，防水防尘，无线充电，NFC，大底主摄\",\"运行内存\":\"12GB\",\"风格\":\"简约风，科技，时尚\"}","comment_summary":"{\"commentCountStr\":\"5万+\",\"generalCountStr\":\"200+\",\"goodCountStr\":\"1.3万+\",\"goodRateShow\":96,\"hotCommentTagStatistics\":[{\"name\":\"时尚美观\"},{\"name\":\"大小适宜\"},{\"name\":\"握感舒适\"},{\"name\":\"线条流畅\"},{\"name\":\"品质优良\"},{\"name\":\"散热性佳\"},{\"name\":\"拍照一流\"},{\"name\":\"送给TA\"},{\"name\":\"使用方便\"},{\"name\":\"按键灵敏\"},{\"name\":\"操作简单\"}],\"imageListCount\":500,\"poorCountStr\":\"200+\",\"videoCountStr\":\"200+\"}"},{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"HUAWEI Pura 70 雪域白 12GB+256GB 超高速风驰闪拍 第二代昆仑玻璃 双超级快充 华为P70智能手机","brand":"华为（HUAWEI）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":5499,"extended_props":"{\"CPU型号\":\"未公布\",\"三防标准\":\"IP68\",\"充电功率\":\"50-79W\",\"后摄主像素\":\"5000万像素\",\"商品名称\":\"华为Pura\",\"商品编号\":\"100095962751\",\"屏幕分辨率\":\"FHD+\",\"屏幕材质\":\"OLED直屏\",\"支持IPv6\":\"支持IPv6\",\"机身内存\":\"256GB\",\"机身色系\":\"白色系\",\"机身颜色\":\"雪域白\",\"特征特质\":\"AI大模型，轻薄，LTPO屏幕，高频PWM调光，全焦段影像，潜望式长焦，防水防尘，无线充电，NFC，大底主摄\",\"运行内存\":\"12GB\",\"风格\":\"简约风，科技，时尚\"}","comment_summary":"{\"commentCountStr\":\"5万+\",\"generalCountStr\":\"200+\",\"goodCountStr\":\"1.3万+\",\"goodRateShow\":96,\"hotCommentTagStatistics\":[{\"name\":\"高端大气\"},{\"name\":\"时尚美观\"},{\"name\":\"手感一流\"},{\"name\":\"超大屏幕\"}],\"imageListCount\":500,\"poorCountStr\":\"200+\",\"videoCountStr\":\"200+\"}"},{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"小米14 徕卡光学镜头 光影猎人900 徕卡75mm浮动长焦 澎湃OS 16+1T 白色 5G AI手机 小米汽车互联","brand":"小米（MI）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":4599,"extended_props":"{\"CPU型号\":\"第三代骁龙8\",\"三防标准\":\"IP68\",\"充电功率\":\"80-119W\",\"后摄主像素\":\"5000万像素\",\"商品名称\":\"小米14\",\"商品编号\":\"100071383566\",\"屏幕分辨率\":\"1.5K\",\"屏幕材质\":\"OLED直屏\",\"机身内存\":\"1TB\",\"机身色系\":\"白色系\",\"机身颜色\":\"白色\",\"运行内存\":\"16GB\",\"风格\":\"大气，时尚，科技\"}","comment_summary":"{\"commentCountStr\":\"50万+\",\"generalCountStr\":\"1000+\",\"goodCountStr\":\"10万+\",\"goodRateShow\":97,\"hotCommentTagStatistics\":[{\"name\":\"时尚美观\"},{\"name\":\"手感一流\"},{\"name\":\"大小适宜\"},{\"name\":\"小巧精致\"},{\"name\":\"品质优良\"},{\"name\":\"散热性佳\"},{\"name\":\"拍摄功能强\"},{\"name\":\"操作方便\"},{\"name\":\"解锁迅速\"},{\"name\":\"畅快办公\"},{\"name\":\"游戏上分\"}],\"imageListCount\":500,\"poorCountStr\":\"1600+\",\"videoCountStr\":\"2200+\"}"},{"timestamp":1720014977011,"store_name":"","product_main_pic":"","product_name":"小米14 徕卡光学镜头 光影猎人900 徕卡75mm浮动长焦 澎湃OS 16+512 白色 5G AI手机 小米汽车互联","brand":"小米（MI）","first_level":"手机通讯","second_level":"手机","third_level":"手机","product_price":4299,"extended_props":"{\"CPU型号\":\"第三代骁龙8\",\"三防标准\":\"IP68\",\"充电功率\":\"80-119W\",\"后摄主像素\":\"5000万像素\",\"商品名称\":\"小米14\",\"商品编号\":\"100071390038\",\"屏幕分辨率\":\"1.5K\",\"屏幕材质\":\"OLED直屏\",\"机身内存\":\"512GB\",\"机身色系\":\"白色系\",\"机身颜色\":\"白色\",\"运行内存\":\"16GB\",\"风格\":\"大气，时尚，科技\"}","comment_summary":"{\"commentCountStr\":\"50万+\",\"generalCountStr\":\"1000+\",\"goodCountStr\":\"10万+\",\"goodRateShow\":97,\"hotCommentTagStatistics\":[{\"name\":\"手感一流\"},{\"name\":\"大小适宜\"},{\"name\":\"性能一流\"},{\"name\":\"时尚美观\"},{\"name\":\"漂亮大方\"},{\"name\":\"散热性佳\"},{\"name\":\"操作简便\"},{\"name\":\"优质耐用\"},{\"name\":\"送给TA\"},{\"name\":\"畅快办公\"},{\"name\":\"解锁迅速\"},{\"name\":\"游戏上分\"}],\"imageListCount\":500,\"poorCountStr\":\"1600+\",\"videoCountStr\":\"2200+\"}"}]`
	skuList := make([]*detail.WalmartSkuResp, 0)
	err = json.Unmarshal([]byte(queryHits), &skuList)
	if err != nil {
		t.Errorf("unmarshal err:%s", err)
	}

	for i := range skuList {
		skuList[i].Id = strconv.Itoa(i)
	}

	res, err := engine.recommend("", getTestUser(), skuList)
	if err != nil {
		t.Error("flow err :", err.Error())
	} else {
		t.Log("result:", res)
	}
}

func TestSearchWalmartIndex(t *testing.T) {
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
	initclients.Start_client(config)
	defer initclients.Stop_client()

	engine := ShoppingEngine{}
	//Content: {"Complete information":"The user is asking for recommendations for low-calorie snacks.", "Question 1":1, "Question 2":"Snacks", "Question 3":"Low-calorie snacks", "Question 4":{"Brand": "", "Calories per serving": "Low", "Ingredients": ["Natural", "Healthy", "Low-fat"], "Flavors": ["Crunchy", "Sweet", "Savory"], "Allergen-free": "", "Dietary restrictions": ["Vegetarian", "Vegan"]}}
	//
	shoppingIntentJson := `{"Complete information":"The user is looking to purchase crispy cookies.","Question 1":1,"Question 2":"Food & Grocery > Snacks > Baked Goods > Cookies","Question 3":"Crispy Cookies","Question 4":{"Brand": "", "Flavor": "Crispy", "Texture": "Chewy", "Type": "Biscotti or Shortbread", "Ingredients": "小麦，糖，黄油，可能含有鸡蛋和乳制品", "Allergens": ["Gluten", "Dairy (if not vegan)"], "Certifications": "", "Net Weight": "100g", "Package Size": "1 pack", "Price": "$3.99", "Availability": "In stock"}}`
	intentMap := make(map[string]any)
	json.Unmarshal([]byte(shoppingIntentJson), &intentMap)
	intent := engine.parseIntent(intentMap)

	matchs := make([]map[string]any, 0)
	if len(intent.ProductName) > 0 {
		matchs = append(matchs, map[string]any{"match": map[string]any{
			"Name": map[string]any{"query": intent.ProductName, "boost": 1},
		}})
		matchs = append(matchs, map[string]any{"match": map[string]any{
			"ShortDescription": map[string]any{"query": intent.ProductName,
				"boost": 0.1},
		}})

		matchs = append(matchs, map[string]any{"match": map[string]any{
			"LongDescription": map[string]any{"query": intent.ProductName,
				"boost": 0.1},
		}})
	}
	if len(intent.Category) > 0 {
		matchs = append(matchs, map[string]any{"match": map[string]any{
			"CategoryPath": map[string]any{"query": intent.Category,
				"boost": 0.05},
		}})

	}
	query := strings.TrimSpace(intent.IndependentQuery)
	if len(query) == 0 {
		query = intent.ProductName
	}

	embs, err := modelclient.GetInstance().QueryEmbedding([]string{query})
	if err != nil {
		t.Error(err.Error(), len(embs))
	}
	//TODO:使用属性进行匹配,目前ES索引没有使用嵌套或者object结构，不能进行嵌套查询

	queryStruct := map[string]any{
		"size":   10,
		"fields": []string{"ItemId", "Timestamp", "Aisle", "ParentItemId", "Color", "MediumImage", "Name", "BrandName", "CategoryPath", "SalePrice", "ShortDescription", "LongDescription"},
		"query": map[string]any{
			"bool": map[string]any{
				"should": matchs,
			},
		},
		"knn": map[string]any{
			"field":          "DescVector",
			"k":              5,
			"num_candidates": 20,
			"boost":          10,
			"query_vector":   embs[0],
		},
	}

	r, err := es.GetInstance().SearchIndex(SHOPPING_INDEX_NAME, queryStruct)
	if err != nil {
		t.Error(err.Error())
	}
	// Print the response status, number of results, and request duration.
	hits := int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	took := int(r["took"].(float64))
	t.Log("hits:", hits, ",took:", took)
	// Print the ID and document source for each hit.
	docs := make([]*detail.WalmartSkuResp, 0)
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		id := hit.(map[string]interface{})["_id"]
		source := hit.(map[string]interface{})["_source"]
		score := hit.(map[string]interface{})["_score"]
		skuDoc := detail.EsToWalmartSku(source.(map[string]interface{}))

		skuDoc.Id = id.(string)
		skuDoc.Score = score.(float64)
		docs = append(docs, &skuDoc)
	}

	if len(docs) == 0 {
		t.Errorf("empty docs")
	}
	for _, doc := range docs {
		t.Logf("id:%s,score:%f,name:%s", doc.Id, doc.Score, doc.Name)
	}
}

func TestGetIndexInfo(t *testing.T) {
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
	initclients.Start_client(config)
	defer initclients.Stop_client()

	str, err := es.GetInstance().GetIndex(SHOPPING_INDEX_NAME)
	if err != nil {
		t.Error("get index err:", err.Error())
	}
	t.Log(str)
}
