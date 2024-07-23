package conf

import (
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type SparkLLM struct {
	HostUrl string `yaml:"host_url"`
	Appid   string `yaml:"appid"`
	Secret  string `yaml:"secret"`
	Key     string `yaml:"key"`
	IsMock  bool   `yaml:"is_mock"`
}

type PromptTemplate struct {
	ChatPrompt              string `yaml:"chat_prompt"`
	TravelPrompt            string `yaml:"travel_prompt"`
	QueryRewritingPrompt    string `yaml:"query_rewriting_prompt"`
	QueryRoute              string `yaml:"query_route"`
	AdditionalInfo          string `yaml:"additional_information"`
	SearchCondition         string `yaml:"search_condition"`
	SkuRecommend            string `yaml:"sku_recommend"`
	WalmartAdditionalInfo   string `yaml:"walmart_additional_information"`
	WalmartShoppingIntent   string `yaml:"walmart_shopping_intent"`
	WalmartSkuRecommend     string `yaml:"walmart_sku_recommend"`
	AgentRouting            string `yaml:"agent_routing"`
	WalmartExtractProductId string `yaml:"walmart_extract_product_id"`
	WalmartChat             string `yaml:"walmart_chat"`
}

type DashScopeLLM struct {
	Key       string `yaml:"key"`
	HostUrl   string `yaml:"host_url"`
	OpenaiUrl string `yaml:"openai_url"`
	Model     string `yaml:"model"`
}

type AgentInfo struct {
	Name   string `yaml:"name"`
	Desc   string `yaml:"desc"`
	Param  string `yaml:"param"`
	Output string `yaml:"output"`
}

type GoogleCustomSearch struct {
	Key     string `yaml:"key"`
	Appid   string `yaml:"cx"`
	Url     string `yaml:"url"`
	Hl      string `yaml:"hl"`
	Lr      string `yaml:"lr"`
	Cr      string `yaml:"cr"`
	IsProxy bool   `yaml:"is_proxy"`
}

type OpenSerpSearch struct {
	Url     string   `yaml:"url"`
	Engines []string `yaml:"engines"`
}

type Config struct {
	//服务地址：0.0.0.0:8080
	ServerAddr     string `yaml:"server_addr"`
	TlsServerAddr  string `yaml:"tls_server_addr"`
	TlsCertPath    string `yaml:"tls_cert_path"`
	TlsCertKeyPath string `yaml:"tls_cert_key_path"`
	//前端websocket地址
	ChatAddr string `yaml:"chat_addr"`

	CookieCodeKey string `yaml:"cookie_code_key"`

	CookieUser string `yaml:"cookie_user"`

	CookieSession string `yaml:"cookie_session"`

	//爬虫数据目录
	CrawlerDataPath string `yaml:"crawler_data_path"`

	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`

	QdrantAddr string `yaml:"qdrant_addr"`

	QuickwitUrl string `yaml:"quickwit_url"`

	ESUrl []string `yaml:"es_url"`

	//模型服务地址
	EmbeddingModelHost   string `yaml:"embedding_model_host"`
	QueryEmbeddingPath   string `yaml:"query_embedding_path"`
	PassageEmbeddingPath string `yaml:"passage_embedding_path"`

	RerankerModelHost     string `yaml:"reranker_model_host"`
	PredictorRerankerPath string `yaml:"predictor_reranker_path"`

	PreRankingThreshold float32 `yaml:"preranking_threshold"`

	MaxCandidates int32 `yaml:"max_candidates"`

	LogLevel string `yaml:"log_level"`

	SparkLLM       SparkLLM       `yaml:"spark_llm"`
	DashScopeLLM   DashScopeLLM   `yaml:"dash_scope_llm"`
	PromptTemplate PromptTemplate `yaml:"prompt_template"`

	Agents map[string][]AgentInfo `yaml:"agents"`

	GoogleCustomSearch GoogleCustomSearch `yaml:"google_custom_search"`
	OpenSerpSearch     OpenSerpSearch     `yaml:"openserp_search"`
}

var ErrHint = "这个问题，我不知道该怎么回答，我可能需要升级了..."
var EmptyHint = "抱歉！我的知识还不够丰富，我正在努力学习..."
var DETAIL_KEY_PREFIX string = "detail-"
var SKU_KEY_PREFIX string = "sku-"
var DOC_KEY_PREFIX string = "doc-"
var CHUNK_KEY_PREFIX string = "chunk-"
var DETAIL_TITLE_FIELD string = "title"
var DETAIL_CONTENT_FIELD string = "content"
var DETAIL_CONTENT_CHUNK_FIELD = "chunkid"
var DETAIL_CONTENT_DOC_FIELD = "docid"
var UPLOAD_FILE_MODE os.FileMode = 0666
var LLM_HISTORY_TOKEN_LEN = 3096
var LLM_PROMPT_TOKEN_LEN = 3096
var EMB_VEC_SIZE = 768

var GlobalConfig *Config

func ParseConfig(configPath string) (*Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {

		return nil, fmt.Errorf("read file:%s err:%s", configPath, err)
	}
	config := &Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, fmt.Errorf("parse yaml:%s err:%s", configPath, err)
	}
	return config, nil
}

func AgentTemplate(conf *Config, flowName string) string {
	flowAgents, ok := conf.Agents[flowName]
	if !ok {
		return ""
	}
	var buf strings.Builder
	for i, agent := range flowAgents {
		if i > 0 {
			buf.WriteString("\r\n")
		}
		info := fmt.Sprintf("%d:\n    agent name:%s;\n    description:%s;\n    parameters:%s。\n",
			i+1, agent.Name, agent.Desc, agent.Param)
		buf.WriteString(info)
	}
	return buf.String()
}

