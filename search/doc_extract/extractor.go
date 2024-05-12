package docextract

type ContentType int

const (
	Text_ContentType  ContentType = 1
	Table_ContentType ContentType = 2
	Image_ContentType ContentType = 2
)

type Page struct {
	Type    ContentType
	Content string
	PageNo  int
}

type Extractor interface {
	Extract() (meta map[string]any, pages []Page, err error)
}
