package engine

import (
	"strings"

	"github.com/blevesearch/bleve/search"

	"github.com/RTradeLtd/Lens/models"
)

func newResult(d *search.DocumentMatch) Result {
	var md models.MetaDataV2
	if d.Fields != nil {
		var fields = d.Fields
		md.DisplayName = fields["display_name"].(string)
		md.Category = fields["category"].(string)
		md.MimeType = fields["mime_type"].(string)
		md.Tags = strings.Split(fields["tags"].(string), ",")
	}

	return Result{
		Hash:  d.ID,
		Score: d.Score,
		MD:    md,
	}
}
