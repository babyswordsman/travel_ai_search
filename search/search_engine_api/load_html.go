package searchengineapi

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	readability "github.com/go-shiori/go-readability"

	logger "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/schema"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

func get_header() (header map[string]string) {
	header = make(map[string]string)
	header["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"
	header["Accept-Encoding"] = "gzip, deflate"
	header["Accept-Language"] = "zh-CN,zh;q=0.9"
	header["Sec-Ch-Ua"] = "\"Google Chrome\";v=\"123\", \"Not:A-Brand\";v=\"8\", \"Chromium\";v=\"123\""
	header["Sec-Ch-Ua-Mobile"] = "?0"
	header["Sec-Ch-Ua-Platform"] = "\"macOS\""
	header["Sec-Fetch-Dest"] = "document"
	header["Sec-Fetch-Mode"] = "navigate"
	header["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
	header["Referer"] = "https://www.google.com/"
	return
}
func LoadHtml(url string) ([]schema.Document, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range get_header() {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return make([]schema.Document, 0), err
	}

	var html_detail []byte
	{
		defer resp.Body.Close()

		//todo:先都读出来，方便调试
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return make([]schema.Document, 0), err
		}

		html_detail = body
	}
	contentEncoding := resp.Header.Get("Content-Encoding")
	logger.Infof("response content-encoding:%s", contentEncoding)
	switch contentEncoding {
	case "gzip":

		zipReader, err := gzip.NewReader(bytes.NewBuffer(html_detail))
		if err != nil {
			return make([]schema.Document, 0), fmt.Errorf("gzip new reader err:%w", err)
		}
		html_detail, err = io.ReadAll(zipReader)
		if err != nil {
			return make([]schema.Document, 0), fmt.Errorf("gzip read err:%w", err)
		}
	case "deflat":
		flateReader := flate.NewReader(bytes.NewBuffer(html_detail))

		html_detail, err = io.ReadAll(flateReader)
		if err != nil {
			return make([]schema.Document, 0), fmt.Errorf("flate read err:%w", err)
		}
	}

	if resp.StatusCode != http.StatusOK {
		return make([]schema.Document, 0), fmt.Errorf("status:%d", resp.StatusCode)
	}
	mimetype := resp.Header.Get("Content-Type")
	if mimetype == "" {
		mimetype = resp.Header.Get("content-type")
	}
	logger.Infof("response charset:%s", mimetype)
	encoding, charsetName, certain := charset.DetermineEncoding(html_detail, mimetype)
	logger.Infof("certain:%t,charset:%s", certain, charsetName)
	if certain && charsetName != "utf-8" {
		html_detail, _, err = transform.Bytes(encoding.NewDecoder(), html_detail)
		if err != nil {
			return make([]schema.Document, 0), fmt.Errorf("transform bytes err:%w", err)
		}
	}
	if logger.IsLevelEnabled(logger.DebugLevel) {
		detail := string(html_detail)
		logger.Debugf("url:%s,status:%d,html:%s", url, resp.StatusCode, detail)
	}

	//todo:提取网页正文算法
	//page := documentloaders.NewHTML(bytes.NewBuffer(html_detail))

	article, err := readability.FromReader(bytes.NewBuffer(html_detail), req.URL)
	if err != nil {
		return make([]schema.Document, 0), fmt.Errorf("extract html content err:%w", err)
	}

	/**
	chunks, err := textsplitter.NewTokenSplitter().SplitText(article.TextContent)
	if err != nil {
		return make([]schema.Document, 0), fmt.Errorf("split text err:%w", err)
	}
	docs := make([]schema.Document, 0)
	for _, chunk := range chunks {
		doc := schema.Document{
			PageContent: chunk,
			Score:       0.0,
			Metadata: map[string]any{
				"url": url,
			},
		}
		docs = append(docs, doc)
	}
	*/
	docs := make([]schema.Document, 0)

	doc := schema.Document{
		PageContent: article.TextContent,
		Score:       0.0,
		Metadata: map[string]any{
			"url": url,
		},
	}
	docs = append(docs, doc)

	return docs, nil
}
