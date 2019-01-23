package engine

import (
	"strings"

	"github.com/RTradeLtd/Lens/models"
	"github.com/go-ego/riot/types"
)

func newResult(d *types.ScoredDoc) Result {
	var score float32
	if len(d.Scores) > 0 {
		score = d.Scores[0]
	}
	var md models.MetaDataV2
	if d.Fields != nil {
		fields, ok := d.Fields.(map[string]string)
		if ok {
			md.DisplayName = fields["display_name"]
			md.Category = fields["category"]
			md.MimeType = fields["mime_type"]
			md.Tags = strings.Split(fields["tags"], ",")
		}
	}

	return Result{
		Hash:  d.DocId,
		Score: score,
		MD:    md,
	}
}
