package shopping

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"travel_ai_search/search/es"
	"travel_ai_search/search/shopping/detail"

	logger "github.com/sirupsen/logrus"
)

func ParseSkuData(path string) int32 {
	parseNum := int32(0)
	f, err := os.Open(path)
	if err != nil {
		logger.Errorln("open file err", err)
		return parseNum
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {

		line := scanner.Text()
		line = strings.TrimSpace(line)

		if len(line) <= 2 {
			continue
		}
		skuMap := make(map[string]any)
		err := json.Unmarshal([]byte(line), &skuMap)
		if err != nil {
			logger.WithField("line", line).Errorln("json unmarshal err:", err)
			continue
		}

		skuDoc := detail.CrawleTransferSku(skuMap)
		//skuDoc.Id = fmt.Sprintf("sku-%d-%d", time.Now().UnixMilli(), parseNum)
		buf, err := json.Marshal(skuDoc)

		if err != nil {
			logger.Errorln("Marshal err", err)
			continue
		}
		if logger.IsLevelEnabled(logger.DebugLevel) {
			logger.Debugln(string(buf))
		}

		i, err := es.GetInstance().AddDocument("sku", fmt.Sprintf("id-%d-%d", time.Now().UnixMilli(), parseNum), string(buf))
		if err != nil {
			logger.Errorln("add document err", err)
			continue
		}
		logger.Infof("add documnent size:%d", i)

		parseNum++
		if parseNum%20 == 19 {
			logger.Infof("deal rows:%d", parseNum)
		}
	}

	return parseNum
}
