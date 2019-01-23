package lens

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/models"
)

// MagnifyOpts declares configuration for magnification
type MagnifyOpts struct {
	Reindex bool
	Tags    []string
}

// MagnifyV2 is used to examine a given content hash, determine if it's parsable
// and returned the summarized meta-data. Returned parameters are in the format of:
// content type, meta-data, error
// TODO implement
func (s *Service) MagnifyV2(hash string, opts MagnifyOpts) (content string, metadata *models.MetaDataV2, err error) {
	if has, err := s.search.Has(hash); err != nil {
		return "", nil, err
	} else if has && !opts.Reindex {
		return "", nil, errors.New("this object has already been indexed")
	}

	var l = logs.NewProcessLogger(s.l, "magnify",
		"hash", hash)

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

	var (
		category string
	)

	// it will be in the format of `<content-type>; charset=...`
	// we use strings.FieldsFunc to seperate the string, and to be able to exmaine the content type
	var parsed = strings.FieldsFunc(contentType, func(r rune) bool { return (r == ';') })
	if parsed == nil || len(parsed) == 0 {
		err = fmt.Errorf("invalid content type '%s'", contentType)
		return
	}
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
		default:
			return "", nil, errors.New("unsupported content type for indexing")
		}
	}

	// clear the stored text so we can parse new text later
	s.ta.Clear()

	return "", &models.MetaDataV2{
		MimeType: contentType,
		Category: category,
	}, nil
}
