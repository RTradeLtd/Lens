package engine

import (
	"errors"
	"strings"
	"time"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/net/com"
	cluster "github.com/go-ego/riot/net/grpc"
	"github.com/go-ego/riot/types"

	"go.uber.org/zap"

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
	e *riot.Engine
	l *zap.SugaredLogger

	stop chan bool

	engineOpts *types.EngineOpts
}

// Opts denotes options for the Lens engine
type Opts struct {
	DictPath  string
	StorePath string
}

// New instantiates a new Engine
func New(l *zap.SugaredLogger, opts Opts) (*Engine, error) {
	var e = &riot.Engine{}
	var r = types.EngineOpts{
		GseDict: opts.DictPath,
	}

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

		stop: make(chan bool),

		engineOpts: &r,
	}, nil
}

// ClusterOpts denotes Lens database clustering options
type ClusterOpts struct {
	Port  string
	Peers []string
}

// Run starts any additional processes required to maintain the engine
func (e *Engine) Run(c *ClusterOpts) {
	if c != nil {
		// TODO: implement
		// https://github.com/go-ego/riot/issues/62
		e.l.Fatal("cluster support is incomplete - do not use")
		e.l.Infow("setting up Riot cluster", "opts", c)
		cluster.InitEngine(com.Config{
			Engine: com.Engine{
				StoreEngine: e.engineOpts.StoreEngine,
				StoreFolder: e.engineOpts.StoreFolder,
			},
			Rpc: com.Rpc{
				GrpcPort: []string{}, // ??
				DistPort: []string{}, // ??
				Port:     c.Port,
			},
		})
		go cluster.InitGrpc(c.Port)
	}

	for {
		select {
		case <-e.stop:
			e.l.Infow("exit signal received")
			var now = time.Now()
			e.e.Close()
			e.l.Infow("index flushed and engine closed",
				"duration", time.Since(now),
				"documents", e.e.NumIndexed())
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
	var docData = types.DocData{
		Content: doc.Content,
		Labels: append(doc.Object.MD.Tags,
			mimeType(doc.Object.MD.MimeType),
			category(doc.Object.MD.Category)),
		Fields: map[string]string{
			"indexed":      time.Now().String(),
			"display_name": doc.Object.MD.DisplayName,
			"category":     doc.Object.MD.Category,
			"mime_type":    doc.Object.MD.MimeType,
			"tags":         strings.Join(doc.Object.MD.Tags, ","),
		},
	}

	// execute index and flush
	l.Debugw("requesting index",
		"labels", docData.Labels,
		"fields", docData.Fields)
	e.e.Index(doc.Object.Hash, docData, doc.Reindex)
	l.Infow("index requested", "size", len(doc.Content))
	var now = time.Now()
	e.e.Flush()
	l.Infow("index flushed",
		"duration", time.Since(now),
		"documents", e.e.NumIndexed())

	return nil
}

// IsIndexed checks if the given content hash has already been indexed
func (e *Engine) IsIndexed(hash string) bool {
	if hash == "" {
		return false
	}
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

	// Hashes restricts what documents to include in query - this is only a
	// filtering option, so some other query fields must be provided as well
	Hashes []string
}

// Result denotes a found document
type Result struct {
	Hash string
	MD   models.MetaDataV2

	Score float32
}

// Search performs a query
func (e *Engine) Search(q Query) ([]Result, error) {
	var l = e.l.With("query", q)
	l.Debug("search requested")

	// construct search request
	var request = types.SearchReq{
		// search for provided plain-text query
		Text: q.Text,

		// query for specific tags, categories, or mimetypes
		Labels: func() (labels []string) {
			if len(q.Tags) > 0 {
				labels = q.Tags
			} else if len(q.Categories) > 0 || len(q.MimeTypes) > 0 {
				labels = make([]string, 0)
			} else {
				return
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
				Must: func() (required []string) {
					if len(q.Required) < 1 {
						return
					}
					// required must be lowercase, and cannot have spaces. we also want
					// to avoid required strings that are too short.
					var splitter = func(c rune) bool { return c == ' ' }
					required = make([]string, 0)
					for _, cur := range q.Required {
						if parts := strings.FieldsFunc(cur, splitter); len(parts) > 1 {
							for _, p := range parts {
								if len(p) > 1 {
									required = append(required, strings.ToLower(p))
								}
							}
						} else {
							if stripped := strings.Replace(cur, " ", "", -1); len(stripped) > 1 {
								required = append(required, strings.ToLower(stripped))
							}
						}
					}
					return
				}(),
			},
		},

		// prevent query from blowing up
		Timeout: 10000, // 10 seconds
		RankOpts: &types.RankOpts{
			MaxOutputs: 1000, // max 1000 documents
		},
	}
	l.Debugw("search constructed", "request", request)

	// always log results of search
	var maxScore float32
	var results = make([]Result, 0)
	defer func() {
		l.Infow("search completed",
			"found", len(results),
			"max_score", maxScore)
	}()

	// execute request
	var out = e.e.Search(request)
	if out.NumDocs == 0 {
		return nil, errors.New("no results found")
	}

	// check returned docs
	l.Debugw("search returned", "docs", out.Docs)
	docs, ok := out.Docs.(types.ScoredDocs)
	if !ok {
		return nil, errors.New("failed to read search result")
	}

	for _, d := range docs {
		var r = newResult(&d)
		if r.Score > maxScore {
			maxScore = r.Score
		}
		results = append(results, r)
	}

	return results, nil
}

// Remove deletes an indexed object from the engine
func (e *Engine) Remove(hash string) {
	e.e.RemoveDoc(hash, true)
	e.e.Flush()
}

// Close shuts down the engine, but not the goroutine started by Run - cancel
// the provided context to do that.
func (e *Engine) Close() {
	e.stop <- true
}
