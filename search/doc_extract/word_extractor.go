package docextract

import (
	"github.com/unidoc/unioffice/document"
)

type WordDoc struct{
	Path string
	Content []byte
}

func (doc *WordDoc) Extract() (meta map[string]any, pages []Page, err error){
	document.Open(doc.Path)
	return nil,nil,nil
}