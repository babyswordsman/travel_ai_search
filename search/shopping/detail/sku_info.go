package detail

type SkuDetail struct {
	Title        string            //标题
	CategoryPath []string          //多级类目
	SkuProps     map[string]string //销售属性
	Describe     string            //描述
	SourceId     string            //原始item_id
}

type SkuDocument struct {
	Timestamp      int64   `json:"timestamp"`
	StoreName      string  `json:"store_name"`
	ProductMainPic string  `json:"product_main_pic"`
	ProductName    string  `json:"product_name"`
	Brand          string  `json:"brand"`
	FirstLevel     string  `json:"first_level"`
	SecondLevel    string  `json:"second_level"`
	ThirdLevel     string  `json:"third_level"`
	ProductPrice   float32 `json:"product_price"`
	//ExtendedProps  map[string]any `json:"extended_props"`
	//CommentSummary map[string]any `json:"comment_summary"`
	ExtendedProps  string `json:"extended_props"`
	CommentSummary string `json:"comment_summary"`
}
