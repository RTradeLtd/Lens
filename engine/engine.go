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

	e.stop = make(chan bool)
	for {
		select {
		case <-e.stop:
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
	e.e.Index(doc.Object.Hash, types.DocData{
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
	}, doc.Reindex)
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

// Result denotes a found document
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

// Remove deletes an indexed object from the engine
func (e *Engine) Remove(hash string) {
	e.e.RemoveDoc(hash, true)
	e.e.Flush()
}

// Close shuts down the engine, but not the goroutine started by Run - cancel
// the provided context to do that.
func (e *Engine) Close() {
	e.stop <- true
	close(e.stop)
	e.e.Close()
}
