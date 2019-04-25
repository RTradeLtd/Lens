package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/blevesearch/bleve"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Lens/v2/engine/queue"
	"github.com/RTradeLtd/Lens/v2/models"
)

// Searcher exposes Engine's primary functions
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../mocks/engine.mock.go github.com/RTradeLtd/Lens/v2/engine.Searcher
type Searcher interface {
	Index(doc Document) error
	Search(ctx context.Context, query Query) ([]Result, error)

	IsIndexed(hash string) bool
	Remove(hash string) error

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
			l.Infow("opening existing index",
				"path", opts.StorePath)
			index, err = bleve.Open(opts.StorePath)
			if err != nil {
				return nil, fmt.Errorf("failed to open existing index at %s: %s",
					opts.StorePath, err.Error())
			}
		} else {
			return nil, fmt.Errorf("failed to instantiate index: %s", err.Error())
		}
	} else {
		l.Infow("successfully created new bleve index",
			"path", opts.StorePath)
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
	go e.q.Run()
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
	if e.IsIndexed(doc.Object.Hash) && !doc.Reindex {
		return fmt.Errorf("document with hash '%s' already exists", doc.Object.Hash)
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

	// queue for index flush
	if e.q.IsStopped() {
		l.Warnw("queue stopped - waiting and trying again")
		time.Sleep(3 * time.Second)
	}
	if err := e.q.Queue(&queue.Item{Key: doc.Object.Hash, Val: DocData{
		Content:  doc.Content,
		Metadata: &doc.Object.MD,
		Properties: &DocProps{
			Indexed: time.Now().String(),
		},
	}}); err != nil {
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

// Search performs a query
func (e *Engine) Search(ctx context.Context, q Query) ([]Result, error) {
	var l = e.l.With("query_id", q.Hash())
	var start = time.Now()
	var request = bleve.SearchRequest{
		Query:  newBleveQuery(&q),
		Fields: allMetaFields,
		Size:   1000,
	}
	l.Debugw("search constructed",
		"query", q,
		"request", request)

	// always log results of search
	var out = &bleve.SearchResult{}
	var results = make([]Result, 0)
	defer func() {
		l.Infow("search ended",
			"found", len(results),
			"max_score", out.MaxScore,
			"duration.search", out.Took,
			"duration.total", time.Since(start))
	}()

	// execute request
	timeout, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	var err error
	out, err = e.index.SearchInContext(timeout, &request)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %s", err.Error())
	}
	if out.Size() == 0 {
		return nil, errors.New("no results found")
	}

	// check returned docs
	l.Debugw("search returned", "hits", out.Hits)
	for _, d := range out.Hits {
		results = append(results, newResult(d))
	}

	return results, nil
}

// Remove deletes an indexed object from the engine
func (e *Engine) Remove(hash string) error {
	if !e.IsIndexed(hash) {
		return fmt.Errorf("no document '%s' in index", hash)
	}
	if e.q.IsStopped() {
		e.l.Warnw("queue stopped - waiting and trying again",
			"hash", hash)
		time.Sleep(3 * time.Second)
	}
	return e.q.Queue(&queue.Item{Key: hash, Val: nil})
}

// Close shuts down the engine
func (e *Engine) Close() { e.stop <- true }
