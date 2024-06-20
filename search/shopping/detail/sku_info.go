package detail

type SkuDetail struct {
	Title        string            //标题
	CategoryPath []string          //多级类目
	SkuProps     map[string]string //销售属性
	Describe     string            //描述
	SourceId     string            //原始item_id
}
