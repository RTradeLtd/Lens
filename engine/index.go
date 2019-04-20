package engine

import (
	"github.com/RTradeLtd/Lens/v2/models"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

// Check generated fields using `bleve fields ./engine/tmp/TestEngine_Search`
const (
	fieldContent     = "content"
	fieldDisplayName = "metadata.display_name"
	fieldMimeType    = "metadata.mime_type"
	fieldCategory    = "metadata.category"
	fieldTags        = "metadata.tags"
	fieldIndexed     = "properties.indexed"
)

// allMetaFields includes all fields except 'content'
var allMetaFields = []string{
	fieldDisplayName,
	fieldMimeType,
	fieldCategory,
	fieldTags,
	fieldIndexed,
}

// DocData defines the structure of indexed objects
type DocData struct {
	Content    string             `json:"content"`
	Metadata   *models.MetaDataV2 `json:"metadata"`
	Properties *DocProps          `json:"properties"`
}

// DocProps denotes additional information about a document
type DocProps struct {
	Indexed string `json:"indexed"` // date indexed
}

func newLensIndex() mapping.IndexMapping {
	var docData = bleve.NewDocumentMapping()

	// DocData::Content
	docData.AddFieldMappingsAt("content", bleve.NewTextFieldMapping())

	// DocData::Metadata
	var mdIndex = bleve.NewDocumentMapping()
	docData.AddSubDocumentMapping("metadata", mdIndex)

	// DocData::Properties
	var pIndex = bleve.NewDocumentMapping()
	pIndex.AddFieldMappingsAt("indexed", bleve.NewDateTimeFieldMapping())
	docData.AddSubDocumentMapping("properties", pIndex)

	// construct overall index
	var m = bleve.NewIndexMapping()
	m.AddDocumentMapping("objects", docData)
	m.DefaultField = "content"
	return m
}
