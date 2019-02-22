package engine

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

func newLensIndex() mapping.IndexMapping {
	var mdIndex = bleve.NewDocumentMapping()

	var pIndex = bleve.NewDocumentMapping()
	pIndex.AddFieldMappingsAt("indexed", bleve.NewDateTimeFieldMapping())

	var docIndex = bleve.NewDocumentMapping()
	docIndex.AddFieldMappingsAt("content", bleve.NewTextFieldMapping())
	docIndex.AddSubDocumentMapping("metadata", mdIndex)
	docIndex.AddSubDocumentMapping("properties", pIndex)

	var m = bleve.NewIndexMapping()
	m.AddDocumentMapping("objects", docIndex)

	return m
}
