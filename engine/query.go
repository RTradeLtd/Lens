package engine

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/blevesearch/bleve"

	"github.com/blevesearch/bleve/search/query"
)

// Query denotes options for a search
type Query struct {
	Text     string
	Required []string

	// Query metadata
	Tags       []string
	Categories []string
	MimeTypes  []string

	// Hashes restricts what documents to include in query - this is only a
	// filtering option, so some other query fields must be provided as well
	Hashes []string
}

// Hash generates a checksum hash for the query
func (q *Query) Hash() string {
	bytes, _ := json.Marshal(q)
	var sum = md5.Sum(bytes)
	return hex.EncodeToString(sum[:])
}

func newBleveQuery(q *Query) query.Query {
	return query.NewConjunctionQuery(
		func() []query.Query {
			var qs = make([]query.Query, 0)

			// require phrase
			if q.Text != "" {
				var tq = query.NewMatchPhraseQuery(q.Text)
				tq.SetField(fieldContent)
				qs = append(qs, tq)
			}

			// require required words
			if len(q.Required) > 0 {
				var bq = newFieldTermsQuery(fieldContent, q.Required)
				bq.SetBoost(100)
				qs = append(qs, bq)
			}

			// require one of provided tags
			if len(q.Tags) > 0 {
				qs = append(qs, newFieldTermsQuery(fieldTags, q.Tags))
			}

			// require one of provided categories
			if len(q.Categories) > 0 {
				qs = append(qs, newFieldTermsQuery(fieldCategory, q.Categories))
			}

			// require one of provided mimetypes
			if len(q.MimeTypes) > 0 {
				qs = append(qs, newFieldTermsQuery(fieldMimeType, q.MimeTypes))
			}

			// require hashses
			if len(q.Hashes) > 0 {
				qs = append(qs, query.NewDocIDQuery(q.Hashes))
			}

			return qs
		}(),
	)
}

func stringSplitter(c rune) bool { return c == ' ' }

func newFieldTermsQuery(field string, should []string) *query.BooleanQuery {
	var bq = bleve.NewBooleanQuery()
	for _, s := range should {
		if parts := strings.FieldsFunc(s, stringSplitter); len(parts) > 1 {
			for _, p := range parts {
				if len(p) > 1 {
					var tq = query.NewTermQuery(strings.ToLower(p))
					tq.SetField(field)
					bq.AddShould(tq)
				}
			}
		} else {
			if stripped := strings.TrimSpace(s); len(stripped) > 1 {
				var tq = query.NewTermQuery(strings.ToLower(stripped))
				tq.SetField(field)
				bq.AddShould(tq)
			}
		}
	}
	return bq
}
