package engine

import (
	"fmt"

	"github.com/blevesearch/bleve/search"

	"github.com/RTradeLtd/Lens/models"
)

// Result denotes a found document
type Result struct {
	Hash string
	MD   models.MetaDataV2

	Score float64
}

func newResult(d *search.DocumentMatch) Result {
	var md models.MetaDataV2
	if d.Fields != nil {
		var fields = d.Fields
		md.DisplayName, _ = fields[fieldDisplayName].(string)
		md.Category, _ = fields[fieldCategory].(string)
		md.MimeType, _ = fields[fieldMimeType].(string)
		rawTags, _ := fields[fieldTags].([]interface{})
		if len(rawTags) > 0 {
			md.Tags = make([]string, len(rawTags))
			for i, v := range rawTags {
				md.Tags[i] = fmt.Sprint(v)
			}
		}
	}

	return Result{
		Hash:  d.ID,
		Score: d.Score,
		MD:    md,
	}
}
