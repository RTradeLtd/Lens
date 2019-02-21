package engine

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/search/query"

	"github.com/blevesearch/bleve"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Lens/engine/queue"
	"github.com/RTradeLtd/Lens/models"
)

// Searcher exposes Engine's primary functions
type Searcher interface {
	Index(doc Document) error
	Search(query Query) ([]Result, error)

	IsIndexed(hash string) bool
	Remove(hash string)

	Close()
}

// Engine implements Lens V2's core search functionality
type Engine struct {
	l *zap.SugaredLogger

	index bleve.Index
	q     *queue.Queue

	stop chan bool
}

// Opts denotes options for the Lens engine
type Opts struct {
	DictPath  string
	StorePath string
	Queue     queue.Options
}

// New instantiates a new Engine
func New(l *zap.SugaredLogger, opts Opts) (*Engine, error) {
	var mdIndex = bleve.NewDocumentMapping()

	var pIndex = bleve.NewDocumentMapping()
	pIndex.AddFieldMappingsAt("indexed", bleve.NewDateTimeFieldMapping())

	var docIndex = bleve.NewDocumentMapping()
	docIndex.AddFieldMappingsAt("content", bleve.NewTextFieldMapping())
	docIndex.AddSubDocumentMapping("metadata", mdIndex)
	docIndex.AddSubDocumentMapping("properties", pIndex)

	var m = bleve.NewIndexMapping()
	m.AddDocumentMapping("objects", docIndex)

	index, err := bleve.New(opts.StorePath, m)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate index: %s", err.Error())
	}
	return &Engine{
		l: l,

		index: index,

		// TODO: reassess queue structure, make use of bleve.Batch
		// 	q: queue.New(l.Named("queue"), e.Flush, e.Close, opts.Queue),

		stop: make(chan bool),
	}, nil
}

// ClusterOpts denotes Lens database clustering options
type ClusterOpts struct {
	Port  string
	Peers []string
}

// Run starts any additional processes required to maintain the engine
func (e *Engine) Run(
//c *ClusterOpts,
) {
	for {
		select {
		case <-e.stop:
			e.l.Infow("exit signal received - closing")
			e.q.Close()
			return
		}
	}
}

// Document denotes a document to index
type Document struct {
	Object  *models.ObjectV2
	Content string
	Reindex bool
}

// Index stores the given object
func (e *Engine) Index(doc Document) error {
	if doc.Object == nil || doc.Object.Hash == "" {
		return errors.New("no object details provided")
	}
	var l = e.l.With("hash", doc.Object.Hash)

	// populate defaults if necessary
	if doc.Object.MD.MimeType == "" {
		l.Debug("defaulting to MimeTypeUnknown")
		doc.Object.MD.MimeType = models.MimeTypeUnknown
	}
	if doc.Object.MD.Category == "" {
		l.Debug("defaulting to category = 'unknown'")
		doc.Object.MD.Category = "unknown"
	}

	// prepare doc data
	var docData = DocData{
		Content:  doc.Content,
		Metadata: &doc.Object.MD,
		Properties: &DocProps{
			Indexed: time.Now().String(),
		},
	}

	// queue for index flush
	e.q.Queue(func() { e.index.Index(doc.Object.Hash, docData) })
	l.Infow("index requested",
		"size", len(doc.Content))

	return nil
}

// IsIndexed checks if the given content hash has already been indexed
func (e *Engine) IsIndexed(hash string) (found bool) {
	if hash == "" {
		return false
	}
	e.q.RLock()
	d, err := e.index.Document(hash)
	if err == nil && d.ID == hash {
		found = true
	}
	e.q.RUnlock()
	return
}

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

// Result denotes a found document
type Result struct {
	Hash string
	MD   models.MetaDataV2

	Score float64
}

// Search performs a query
func (e *Engine) Search(q Query) ([]Result, error) {
	var l = e.l.With("query", q)
	l.Debug("search requested")
	var request = bleve.SearchRequest{
		Query: query.NewBooleanQuery(
			func() []query.Query {
				var qs = make([]query.Query, 0)

				// require phrase
				if q.Text != "" {
					qs = append(qs, query.NewMatchPhraseQuery(q.Text))
				}

				// require required words
				if len(q.Required) > 0 {
					var terms = make([]query.Query, 0)
					// required must be lowercase, and cannot have spaces. we also want
					// to avoid required strings that are too short.
					var splitter = func(c rune) bool { return c == ' ' }
					for _, cur := range q.Required {
						if parts := strings.FieldsFunc(cur, splitter); len(parts) > 1 {
							for _, p := range parts {
								if len(p) > 1 {
									terms = append(terms, query.NewTermQuery(strings.ToLower(p)))
								}
							}
						} else {
							if stripped := strings.Replace(cur, " ", "", -1); len(stripped) > 1 {
								terms = append(terms, query.NewTermQuery(strings.ToLower(stripped)))
							}
						}
					}
					var bq = query.NewBooleanQuery(terms, nil, nil)
					bq.SetBoost(10)
					qs = append(qs, bq)
				}

				// require tags
				if len(q.Tags) > 0 {
					// TODO
				}

				// require categories
				if len(q.Categories) > 0 {
					// TODO
				}

				// require mimetypes
				if len(q.MimeTypes) > 0 {
					// TODO
				}

				// require hashses
				if len(q.Hashes) > 0 {
					qs = append(qs, query.NewDocIDQuery(q.Hashes))
				}

				return qs
			}(),
			nil,
			nil,
		),
		Size: 1000,
	}
	l.Debugw("search constructed", "request", request)

	// always log results of search
	var maxScore float64
	var results = make([]Result, 0)
	defer func() {
		l.Infow("search completed",
			"found", len(results),
			"max_score", maxScore)
	}()

	// execute request
	e.q.RLock()
	out, err := e.index.Search(&request)
	e.q.RUnlock()
	if err != nil {
		return nil, err
	}
	if out.Size() == 0 {
		return nil, errors.New("no results found")
	}

	// check returned docs
	l.Debugw("search returned", "docs", out.Hits)

	for _, d := range out.Hits {
		var r = newResult(d)
		if r.Score > maxScore {
			maxScore = r.Score
		}
		results = append(results, r)
	}

	return results, nil
}

// Remove deletes an indexed object from the engine
func (e *Engine) Remove(hash string) {
	e.q.Queue(func() { e.index.Delete(hash) })
}

// Close shuts down the engine
func (e *Engine) Close() { e.stop <- true }
