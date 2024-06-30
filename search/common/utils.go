package common

import (
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"travel_ai_search/search/conf"
)

func ParseLoginCookie() {

}
func GetStack() string {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}

/*
*
带有代码行号的异常输出
*/
func Errorf(format string, err error) error {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf(fmt.Sprintf("[runtime.Caller failed]%s", format), err)
	} else {
		funcName := runtime.FuncForPC(pc).Name()
		fileName := path.Base(file)
		return fmt.Errorf(fmt.Sprintf("[%s.%s:%d]%s", fileName, funcName, line, format), err)
	}

}

func CosineSimilarity(emb1, emb2 []float32) (float32, error) {
	if len(emb1) != len(emb2) {
		return 0, fmt.Errorf("emb1 len:%d not equal emb2 len:%d ", len(emb1), len(emb2))
	}
	var t1 float64
	var t2 float64
	var t3 float64
	for ind := range emb1 {
		t1 += float64(emb1[ind]) * float64(emb2[ind])
		t2 += float64(emb1[ind]) * float64(emb1[ind])
		t3 += float64(emb2[ind]) * float64(emb2[ind])
	}
	cos := t1 / (math.Sqrt(t2) * math.Sqrt(t3))
	return float32(cos), nil
}

func GetTestConfigPath() string {
	wd, err := os.Getwd()
	if err != nil {
		wd = "/"
	}
	if strings.Contains(wd, "/travel_ai_search") {
		dir := wd
		file := ""
		i := 20
		for ; dir != "" && i >= 0; i-- {
			for len(dir) > 0 && os.IsPathSeparator(dir[len(dir)-1]) {
				dir = dir[:len(dir)-1]
			}
			dir, file = filepath.Split(dir)
			if file == "travel_ai_search" {
				wd = filepath.Join(dir, file)
				break
			}
		}
	}
	path := fmt.Sprintf("%s/config/conf_local.yaml", wd)
	return path
}

func GetProjectPath() string {
	wd, err := os.Getwd()
	if err != nil {
		wd = "/"
	}
	if strings.Contains(wd, "/travel_ai_search") {
		dir := wd
		file := ""
		i := 20
		for ; dir != "" && i >= 0; i-- {
			for len(dir) > 0 && os.IsPathSeparator(dir[len(dir)-1]) {
				dir = dir[:len(dir)-1]
			}
			dir, file = filepath.Split(dir)
			if file == "travel_ai_search" {
				wd = filepath.Join(dir, file)
				break
			}
		}
	}
	return wd
}

func GetUploadPath(config *conf.Config) string {
	dir, _ := filepath.Split(config.CrawlerDataPath)
	return dir
}
