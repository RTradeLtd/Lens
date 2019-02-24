package lens

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RTradeLtd/Lens/engine"
	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/models"
)

// magnifyOpts declares configuration for magnification
type magnifyOpts struct {
	DisplayName string
	Reindex     bool
	Tags        []string
}

func (v *V2) magnify(hash string, opts magnifyOpts) (content string, metadata *models.MetaDataV2, err error) {
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
	var category models.MimeType
	switch parsed[0] {
	case "application/pdf":
		category = models.MimeTypePDF
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
			category = models.MimeTypeDocument
			content = string(contents)
		case "image":
			category = models.MimeTypeImage
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
		Category:    string(category),
		Tags:        opts.Tags,
	}, nil
}

// Store is used to store our collected meta data in a formatted object
func (v *V2) store(hash, content string, md *models.MetaDataV2, reindex bool) error {
	return v.se.Index(engine.Document{
		Object: &models.ObjectV2{
			Hash: hash,
			MD:   *md,
		},
		Content: content,
		Reindex: reindex,
	})
}

// Remove is used to remove an indexed object
func (v *V2) remove(hash string) error {
	if !v.se.IsIndexed(hash) {
		return fmt.Errorf("object '%s' does not exist", hash)
	}
	v.se.Remove(hash)
	return nil
}
