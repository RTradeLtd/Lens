package engine

import (
	"errors"
	"strings"
	"time"

	"github.com/RTradeLtd/Lens/models"
	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
	"go.uber.org/zap"
)

// Searcher exposes Engine's primary functions
type Searcher interface {
	Index(object *models.ObjectV2, content string, force bool)
	IsIndexed(hash string) bool
	Search(query Query) ([]Result, error)
	Remove(hash string)
}

// Engine implements Lens V2's core search functionality
type Engine struct {
	e *riot.Engine
	l *zap.SugaredLogger

	stop chan bool
}

// Opts denotes options for the Lens engine
type Opts struct {
	DictPath  string
	StorePath string
}

// New instantiates a new Engine
func New(l *zap.SugaredLogger, opts Opts) (*Engine, error) {
	var e = &riot.Engine{}
	var r = types.EngineOpts{}

	// database persistence settings
	if opts.StorePath != "" {
		r.UseStore = true
		r.StoreEngine = "bg" // use badger for persistence
		r.StoreFolder = opts.StorePath
	}

	l.Infow("starting up search engine", "config", r)
	e.Init(r)
	return &Engine{
		e: e,
		l: l,
	}, nil
}

// Run initiates period index flushes
func (e *Engine) Run(indexInterval time.Duration) {
	var ticker = time.NewTicker(indexInterval)
	e.stop = make(chan bool)
	for {
		select {
		case <-ticker.C:
			var now = time.Now()
			e.e.Flush()
			e.l.Infow("index flushed",
				"duration", time.Since(now),
				"documents", e.e.NumIndexed())

		case <-e.stop:
			ticker.Stop()
			e.l.Infow("exit signal received")
			var now = time.Now()
			e.e.Flush()
			e.l.Infow("index flushed",
				"duration", time.Since(now),
				"documents", e.e.NumIndexed())
			return
		}
	}
}

// Index stores the given object
func (e *Engine) Index(object *models.ObjectV2, content string, force bool) {
	e.e.Index(object.Hash, types.DocData{
		Content: content,
		Labels: append(object.MD.Tags,
			mimeType(object.MD.MimeType),
			category(object.MD.Category)),
		Fields: map[string]string{
			"indexed":      time.Now().String(),
			"display_name": object.MD.DisplayName,
			"category":     object.MD.Category,
			"mime_type":    object.MD.MimeType,
			"tags":         strings.Join(object.MD.Tags, ","),
		},
	}, force)
	e.l.Debugw("indexed", "hash", object.Hash)
}

// IsIndexed checks if the given content hash has already been indexed
func (e *Engine) IsIndexed(hash string) bool {
	return e.e.HasDoc(hash)
}

// Query denotes options for a search
type Query struct {
	Text     string
	Required []string

	// Query metadata
	Tags       []string
	Categories []string
	MimeTypes  []string

	// restrict what documents to include in query
	Hashes []string
}

type Result struct {
	Hash string
	MD   models.MetaDataV2

	Score float32
}

// Search performs a query
func (e *Engine) Search(q Query) ([]Result, error) {
	e.l.Debugw("search requested", "query", q)
	var out = e.e.Search(types.SearchReq{
		// search for provided plain-text query
		Text: q.Text,

		// query for specific tags, categories, or mimetypes
		Labels: func() (labels []string) {
			if len(q.Categories) > 0 || len(q.MimeTypes) > 0 {
				if len(q.Tags) > 0 {
					labels = q.Tags
				} else {
					labels = make([]string, 0)
				}
			}
			for _, c := range q.Categories {
				labels = append(labels, category(c))
			}
			for _, m := range q.MimeTypes {
				labels = append(labels, mimeType(m))
			}
			return
		}(),

		// filter results by specific documents, if requested
		DocIds: func() (ids map[string]bool) {
			if len(q.Hashes) > 0 {
				ids = make(map[string]bool)
				for _, h := range q.Hashes {
					ids[h] = true
				}
			}
			return ids
		}(),

		// require certain words
		Logic: types.Logic{
			Expr: types.Expr{
				Must: q.Required,
			},
		},

		// prevent query from blowing up
		Timeout: 10000, // 10 seconds
		RankOpts: &types.RankOpts{
			MaxOutputs: 1000, // max 1000 documents
		},
	})
	if out.NumDocs == 0 {
		return nil, errors.New("no results found")
	}

	e.l.Debugw("search returned", "docs", out.Docs)
	docs, ok := out.Docs.(types.ScoredDocs)
	if !ok {
		return nil, errors.New("failed to read search result")
	}
	var results = make([]Result, 0)
	for _, d := range docs {
		results = append(results, newResult(&d))
	}
	return results, nil
}

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

// Remove deletes an indexed object from the engine
func (e *Engine) Remove(hash string) {
	e.e.RemoveDoc(hash, true)
}

// Close shuts down the engine, but not the goroutine started by Run - cancel
// the provided context to do that.
func (e *Engine) Close() {
	e.stop <- true
	close(e.stop)
	e.e.Close()
}
