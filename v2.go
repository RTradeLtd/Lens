package lens

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/Lens/analyzer/ocr"
	"github.com/RTradeLtd/Lens/engine"
	"github.com/RTradeLtd/Lens/xtractor/planetary"
	"github.com/RTradeLtd/grpc/lensv2"
	"github.com/RTradeLtd/rtfs"

	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/models"
)

// V2 is the new Lens API, and implements the LensV2 gRPC interface directly.
type V2 struct {
	se   engine.Searcher
	ipfs rtfs.Manager

	// Analysis classes
	oc *ocr.Analyzer
	px *planetary.Extractor
	tf images.TensorflowAnalyzer

	l *zap.SugaredLogger
}

// V2Options denotes options for the V2 Lens API
type V2Options struct {
	TesseractConfigPath string

	Engine engine.Opts
}

// NewV2 instantiates a new V2 API
func NewV2(
	opts V2Options,
	ipfs rtfs.Manager,
	ia images.TensorflowAnalyzer,
	logger *zap.SugaredLogger,
) (*V2, error) {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	// create new engine
	se, err := engine.New(logger.Named("engine"), opts.Engine)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate search engine: %s", err.Error())
	}
	go se.Run(nil)

	return &V2{
		se:   se,
		ipfs: ipfs,

		tf: ia,
		px: planetary.NewPlanetaryExtractor(ipfs),
		oc: ocr.NewAnalyzer(opts.TesseractConfigPath, logger.Named("ocr")),
		l:  logger.Named("service.v2"),
	}, nil
}

// NewV2WithEngine instantiates a Lens V2 service with the given engine
func NewV2WithEngine(
	opts V2Options,
	ipfs rtfs.Manager,
	ia images.TensorflowAnalyzer,
	se engine.Searcher,
	logger *zap.SugaredLogger,
) *V2 {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	return &V2{
		se:   se,
		ipfs: ipfs,

		tf: ia,
		px: planetary.NewPlanetaryExtractor(ipfs),
		oc: ocr.NewAnalyzer(opts.TesseractConfigPath, logger.Named("ocr")),
		l:  logger.Named("service.v2"),
	}
}

// Close releases Lens resources
func (v *V2) Close() { v.se.Close() }

// Index analyzes and stores the given object
func (v *V2) Index(ctx context.Context, req *lensv2.IndexReq) (*lensv2.IndexResp, error) {
	var l = v.l.With("request", req)
	switch req.GetType() {
	case lensv2.IndexReq_IPLD:
		break
	default:
		return nil, status.Errorf(codes.InvalidArgument,
			"invalid data type '%s' provided", req.GetType())
	}

	var hash = req.GetHash()
	var reindex = req.GetOptions().GetReindex()
	content, md, err := v.magnify(hash, MagnifyOpts{
		DisplayName: req.GetDisplayName(),
		Tags:        req.GetTags(),
		Reindex:     reindex,
	})
	if err != nil {
		l.Errorw("failed to magnify document", "error", err)
		if strings.Contains(err.Error(), "failed to find content") {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.FailedPrecondition,
			"failed to perform magnification for '%s': %s", hash, err.Error())
	}

	if !reindex {
		if err = v.store(hash, content, md); err != nil {
			l.Errorw("failed to store document", "error", err)
			return nil, status.Errorf(codes.Internal,
				"failed to store requested document: %s", err.Error())
		}
	} else {
		if err = v.update(hash, content, md); err != nil {
			l.Errorw("failed to update document", "error", err)
			return nil, status.Errorf(codes.Internal,
				"failed to update requested document: %s", err.Error())
		}
	}

	l.Info("document indexed")

	return &lensv2.IndexResp{
		Doc: &lensv2.Document{
			Hash:        hash,
			DisplayName: md.DisplayName,
			MimeType:    md.MimeType,
			Category:    md.Category,
			Tags:        md.Tags,
		},
	}, nil
}

// Search executes a query against the Lens index
func (v *V2) Search(ctx context.Context, req *lensv2.SearchReq) (*lensv2.SearchResp, error) {
	var (
		err     error
		results []engine.Result
		opts    = req.GetOptions()
	)

	if opts == nil {
		results, err = v.se.Search(engine.Query{Text: req.GetQuery()})
	} else {
		results, err = v.se.Search(engine.Query{
			Text:       req.GetQuery(),
			Required:   opts.GetRequired(),
			Tags:       opts.GetTags(),
			Categories: opts.GetCategories(),
			MimeTypes:  opts.GetMimeTypes(),
			Hashes:     opts.GetHashes(),
		})
	}
	if err != nil {
		v.l.Errorw("error occured on query execution",
			"error", err, "query", req)
		return nil, status.Errorf(codes.Internal,
			"error occured on query execution: %s", err.Error())
	}

	v.l.Debugw("query completed",
		"query", req, "results", len(results))
	return &lensv2.SearchResp{
		Results: func() []*lensv2.SearchResp_Result {
			var formatted = make([]*lensv2.SearchResp_Result, len(results))
			for i := 0; i < len(results); i++ {
				var r = results[i]
				formatted[i] = &lensv2.SearchResp_Result{
					Score: r.Score,
					Doc: &lensv2.Document{
						Hash:        r.Hash,
						DisplayName: r.MD.DisplayName,
						MimeType:    r.MD.MimeType,
						Category:    r.MD.Category,
						Tags:        r.MD.Tags,
					},
				}
			}
			return formatted
		}(),
	}, nil
}

// Remove unindexes and deletes the requested object
func (v *V2) Remove(ctx context.Context, req *lensv2.RemoveReq) (*lensv2.RemoveResp, error) {
	if err := v.remove(req.GetHash()); err != nil {
		return nil, status.Errorf(codes.NotFound,
			"failed to remove requested hash: %s", err.Error())
	}
	return &lensv2.RemoveResp{}, nil
}

// MagnifyOpts declares configuration for magnification
type MagnifyOpts struct {
	DisplayName string
	Reindex     bool
	Tags        []string
}

func (v *V2) magnify(hash string, opts MagnifyOpts) (content string, metadata *models.MetaDataV2, err error) {
	if v.se.IsIndexed(hash) && !opts.Reindex {
		return "", nil, fmt.Errorf("object '%s' has already been indexed", hash)
	}

	// set up args
	if opts.Tags == nil {
		opts.Tags = make([]string, 0)
	}

	// set up logger
	var l = logs.NewProcessLogger(v.l, "magnify", "hash", hash)
	var start = time.Now()
	defer func() { l.Infow("magnification ended", "duration", time.Since(start)) }()

	// retrieve object and detect content type
	contents, err := v.px.ExtractContents(hash)
	if err != nil {
		return "", nil, fmt.Errorf("failed to find content for hash '%s'", hash)
	}
	contentType := http.DetectContentType(contents)
	if contentType == "" {
		return "", nil, fmt.Errorf("unknown content type for document '%s'", hash)
	}
	l.Infow("object retrieved and content type detected",
		"content_type", contentType)

	// contentType will be in the format of `<content-type>; charset=...`
	// we use strings.FieldsFunc to seperate the string, and to be able to exmaine
	// the content type
	var parsed = strings.FieldsFunc(contentType, func(r rune) bool { return (r == ';') })
	if parsed == nil || len(parsed) == 0 {
		return "", nil, fmt.Errorf("invalid content type '%s'", contentType)
	}

	// scrape for content based on content-type
	var category string
	switch parsed[0] {
	case "application/pdf":
		category = "pdf"
		text, err := v.oc.Analyze(hash, contents, "pdf")
		if err != nil {
			return "", nil, err
		}
		content = text
	default:
		var parsed2 = strings.FieldsFunc(contentType, func(r rune) bool { return (r == '/') })
		if parsed2 == nil || len(parsed2) == 0 {
			return "", nil, fmt.Errorf("invalid content type '%s'", contentType)
		}
		switch parsed2[0] {
		case "text":
			category = "document"
			content = string(contents)
		case "image":
			category = "image"
			keyword, err := v.tf.Analyze(hash, contents)
			if err != nil {
				l.Warnw("failed to categorize image", "error", err)
				return "", nil, errors.New("failed to categorize image")
			}

			// grab any text in image
			text, err := v.oc.Analyze(hash, contents, "image")
			if err != nil {
				l.Warnw("failed to OCR image", "error", err)
				content = keyword
			} else {
				content = text
			}
			opts.Tags = append(opts.Tags, keyword)
		default:
			return "", nil, errors.New("unsupported content type for indexing")
		}
	}

	return content, &models.MetaDataV2{
		DisplayName: opts.DisplayName,
		MimeType:    contentType,
		Category:    category,
		Tags:        opts.Tags,
	}, nil
}

// Store is used to store our collected meta data in a formatted object
func (v *V2) store(hash string, content string, md *models.MetaDataV2) error {
	return v.se.Index(engine.Document{
		Object: &models.ObjectV2{
			Hash: hash,
			MD:   *md,
		},
		Content: content,
	})
}

// Update is used to update an indexed object
func (v *V2) update(hash string, content string, md *models.MetaDataV2) error {
	if err := v.remove(hash); err != nil {
		return err
	}
	return v.store(hash, content, md)
}

// Remove is used to remove an indexed object
func (v *V2) remove(hash string) error {
	if v.se.IsIndexed(hash) {
		return fmt.Errorf("object '%s' does not exist", hash)
	}
	v.se.Remove(hash)
	return nil
}
