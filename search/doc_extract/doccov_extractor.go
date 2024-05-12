package docextract

import (
	docconv "code.sajari.com/docconv/v2"
)

type DocconvExtractor struct {
	Path string
}

func (doc *DocconvExtractor) Extract() (meta map[string]any, pages []Page, err error) {
	var res = &docconv.Response{}
	res, err = docconv.ConvertPath(doc.Path)
	meta = make(map[string]any)
	pages = make([]Page, 0)

	if err != nil {
		return
	}
	pages = append(pages, Page{
		Content: res.Body,
		PageNo:  1,
		Type:    Text_ContentType,
	})

	for k, v := range res.Meta {
		meta[k] = v
	}

	return
}
