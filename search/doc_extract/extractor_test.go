package docextract

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtracte(t *testing.T) {
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

	dataPath := filepath.Join(wd, "data")
	dir, err := os.Open(dataPath)
	if err != nil {
		t.Errorf("open %s err %s", dataPath, err.Error())
		return
	}

	files, err := dir.Readdir(0)
	if err != nil {
		t.Errorf("open %s err %s", dataPath, err.Error())
		return
	}

	extMap := make(map[string]string)

	extMap[".pdf"] = ""
	//extMap[".docx"] = ""
	//extMap[".xlsx"] = ""
	//extMap[".doc"] = ""
	//extMap[".md"] = ""
	//extMap[".csv"] = ""

	for _, f := range files {
		t.Log(dataPath, f.Name())
		if f.IsDir() {
			continue
		}
		ext := filepath.Ext(f.Name())

		_, ok := extMap[ext]
		t.Logf("name:%s,ext:%s,ok:%t", f.Name(), ext, ok)
		if !ok {
			continue
		}

		extractor := DocconvExtractor{
			Path: filepath.Join(dataPath, f.Name()),
		}
		_, pages, err := extractor.Extract()
		if err != nil {
			t.Errorf("extract file:%s err:%s", extractor.Path, err.Error())
			return
		}
		fmt.Println(pages[0].Content)
		t.Log(extractor.Path)
		t.Log(pages[0].Content)
	}

}
