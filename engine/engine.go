package engine

import (
	"context"
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
	Search(ctx context.Context, query Query) ([]Result, error)

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
	StorePath string
	Queue     queue.Options
}

// New instantiates a new Engine
func New(l *zap.SugaredLogger, opts Opts) (*Engine, error) {
	index, err := bleve.New(opts.StorePath, newLensIndex())
	if err != nil {
		if err == bleve.ErrorIndexPathExists {
			index, err = bleve.Open(opts.StorePath)
			if err != nil {
				return nil, fmt.Errorf("failed to instantiate index: %s", err.Error())
			}
		} else {
			return nil, fmt.Errorf("failed to instantiate index: %s", err.Error())
		}
	}

	var queueLogger = l.Named("queue")
	return &Engine{
		l: l,

		index: index,

		q: queue.New(queueLogger,
			func(items []*queue.Item) error {
				var b = index.NewBatch()
				for _, item := range items {
					if item != nil {
						if item.Val != nil {
							if err := b.Index(item.Key, item.Val); err != nil {
								queueLogger.Errorw("failed to add document to batch",
									"error", err, "key", item.Key)
							}
						} else {
							b.Delete(item.Key)
						}
					}
				}
				return index.Batch(b)
			},
			index.Close,
			opts.Queue),

		stop: make(chan bool, 1),
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
	e.q.Run()
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
	if e.q.IsStopped() {
		l.Warnw("queue stopped - waiting and trying again")
		time.Sleep(time.Second)
	}
	if err := e.q.Queue(&queue.Item{Key: doc.Object.Hash, Val: docData}); err != nil {
		return fmt.Errorf("could not index object: %s", err.Error())
	}
	l.Infow("index requested",
		"size", len(doc.Content))

	return nil
}

// IsIndexed checks if the given content hash has already been indexed
func (e *Engine) IsIndexed(hash string) bool {
	if hash == "" {
		return false
	}
	d, err := e.index.Document(hash)
	if err == nil && d != nil && d.ID == hash {
		return true
	}
	return false
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
func (e *Engine) Search(ctx context.Context, q Query) ([]Result, error) {
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
	timeout, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	out, err := e.index.SearchInContext(timeout, &request)
	cancel()
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
	e.q.Queue(&queue.Item{Key: hash, Val: nil})
}

// Close shuts down the engine
func (e *Engine) Close() { e.stop <- true }
