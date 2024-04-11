package conf

type SparkLLM struct {
	HostUrl      string `yaml:"host_url"`
	Appid        string `yaml:"appid"`
	Secret       string `yaml:"secret"`
	Key          string `yaml:"key"`
	ChatPrompt   string `yaml:"chat_prompt"`
	TravelPrompt string `yaml:"travel_prompt"`
	IsMock       bool   `yaml:"is_mock"`
}

type Config struct {
	//服务地址：0.0.0.0:8080
	ServerAddr string `yaml:"server_addr"`

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

	//模型服务地址
	EmbeddingModelHost   string `yaml:"embedding_model_host"`
	QueryEmbeddingPath   string `yaml:"query_embedding_path"`
	PassageEmbeddingPath string `yaml:"passage_embedding_path"`

	RerankerModelHost     string `yaml:"reranker_model_host"`
	PredictorRerankerPath string `yaml:"predictor_reranker_path"`

	PreRankingThreshold float32 `yaml:"preranking_threshold"`

	MaxCandidates int32 `yaml:"max_candidates"`

	SparkLLM SparkLLM `yaml:"spark_llm"`
}

var ErrHint = "这个问题，我不知道该怎么回答，我可能需要升级了..."
var EmptyHint = "抱歉！我的知识还不够丰富，我正在努力学习..."
var DETAIL_KEY_PREFIX string = "detail-"
var DETAIL_TITLE_FIELD string = "title"
var DETAIL_CONTENT_FIELD string = "content"

var GlobalConfig *Config
