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

// MagnifyOpts declares configuration for magnification
type MagnifyOpts struct {
	DisplayName string
	Reindex     bool
	Tags        []string
}

// MagnifyV2 is used to examine a given content hash, determine if it's parsable,
// and return the object's content and metadata
func (s *Service) MagnifyV2(hash string, opts MagnifyOpts) (content string, metadata *models.MetaDataV2, err error) {
	if s.se.IsIndexed(hash) && !opts.Reindex {
		return "", nil, fmt.Errorf("object '%s' has already been indexed", hash)
	}

	// set up args
	if opts.Tags == nil {
		opts.Tags = make([]string, 0)
	}

	// set up logger
	var l = logs.NewProcessLogger(s.l, "magnify", "hash", hash)
	var start = time.Now()
	defer func() { l.Infow("magnification ended", "duration", time.Since(start)) }()

	// retrieve object and detect content type
	contents, err := s.px.ExtractContents(hash)
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
		text, err := s.oc.Analyze(hash, contents, "pdf")
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
			keyword, err := s.images.Analyze(hash, contents)
			if err != nil {
				l.Warnw("failed to categorize image", "error", err)
				return "", nil, errors.New("failed to categorize image")
			}

			// grab any text in image
			text, err := s.oc.Analyze(hash, contents, "image")
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

// StoreV2 is used to store our collected meta data in a formatted object
func (s *Service) StoreV2(hash string, content string, md *models.MetaDataV2) error {
	s.se.Index(engine.Document{
		Object: &models.ObjectV2{
			Hash: hash,
			MD:   *md,
		},
		Content: content,
		Reindex: true,
	})
	return nil
}

// UpdateV2 is used to update an indexed object
func (s *Service) UpdateV2(hash string, content string, md *models.MetaDataV2) error {
	if err := s.RemoveV2(hash); err != nil {
		return err
	}
	s.StoreV2(hash, content, md)
	return nil
}

// RemoveV2 is used to remove an indexed object
func (s *Service) RemoveV2(hash string) error {
	if s.se.IsIndexed(hash) {
		return fmt.Errorf("object '%s' does not exist", hash)
	}
	return nil
}
