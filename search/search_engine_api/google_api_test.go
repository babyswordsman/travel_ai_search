package searchengineapi

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"travel_ai_search/search/conf"

	yaml "gopkg.in/yaml.v3"
)

func parseConfig(configPath string) (*conf.Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {

		return nil, fmt.Errorf("read file:%s err:%s", configPath, err)
	}
	config := &conf.Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, fmt.Errorf("parse yaml:%s err:%s", configPath, err)
	}
	return config, nil
}

var configPath = flag.String("config", "config path", "conf.yaml")

func TestGoogleSearch(t *testing.T) {
	flag.Parse()

	t.Logf("exe:%s", getExePath())
	t.Logf("abs:%s", getAbsPath())
	t.Logf("wdp:%s", getWorkingDirPath())
	config, err := parseConfig(*configPath)
	if err != nil {
		t.Errorf("path:%s,err:%v", *configPath, err)
	}
	googleSearchEngine := &GoogleSearchEngine{}
	items, err := googleSearchEngine.Search(context.Background(), config, "世界 和平 伊朗与以色列问题")
	if err != nil {
		t.Errorf("search err:%v", err)
	}
	buf, _ := json.Marshal(items)
	t.Logf("items:%v", string(buf))
}

func getExePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)
	fmt.Println("exePath:", exePath)
	return exePath
}

func getAbsPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	fmt.Println("absPath:", dir)
	return dir
}

func getWorkingDirPath() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println("workingDirPath:", dir)
	return dir
}
