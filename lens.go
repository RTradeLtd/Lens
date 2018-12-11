package lens

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RTradeLtd/Lens/logs"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/Lens/analyzer/ocr"
	"github.com/RTradeLtd/rtfs"

	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/search"
	"github.com/RTradeLtd/Lens/text"
	"github.com/RTradeLtd/Lens/utils"
	"github.com/RTradeLtd/Lens/xtractor/planetary"
	"github.com/RTradeLtd/config"
	"github.com/gofrs/uuid"
)

// Service contains the various components of Lens
type Service struct {
	ipfs   rtfs.Manager
	images images.TensorflowAnalyzer
	search search.Searcher

	oc *ocr.Analyzer
	ta *text.Analyzer
	px *planetary.Extractor

	l *zap.SugaredLogger
}

// ConfigOpts are options used to configure the lens service
type ConfigOpts struct {
	UseChainAlgorithm   bool
	DataStorePath       string
	ModelsPath          string
	TesseractConfigPath string
	API                 APIOpts
}

// APIOpts defines options for the lens API
type APIOpts struct {
	IP   string
	Port string
}

// NewService is used to generate our Lens service
func NewService(opts ConfigOpts, cfg config.TemporalConfig,
	rm rtfs.Manager,
	ia images.TensorflowAnalyzer,
	ss search.Searcher,
	logger *zap.SugaredLogger) (*Service, error) {
	// instantiate utility classes
	px, err := planetary.NewPlanetaryExtractor(rm)
	if err != nil {
		return nil, err
	}

	return &Service{
		ipfs:   rm,
		images: ia,
		search: ss,

		px: px,
		ta: text.NewTextAnalyzer(opts.UseChainAlgorithm),
		oc: ocr.NewAnalyzer(opts.TesseractConfigPath, logger.Named("ocr")),

		l: logger.Named("service"),
	}, nil
}

// Magnify is used to examine a given content hash, determine if it's parsable
// and returned the summarized meta-data. Returned parameters are in the format of:
// content type, meta-data, error
func (s *Service) Magnify(hash string, reindex bool) (metadata *models.MetaData, err error) {
	if has, err := s.search.Has(hash); err != nil {
		return nil, err
	} else if has && !reindex {
		return nil, errors.New("this object has already been indexed")
	}

	var l = logs.NewProcessLogger(s.l, "magnify",
		"hash", hash)

	var start = time.Now()
	defer func() { l.Infow("magnification ended", "duration", time.Since(start)) }()

	// retrieve object and detect content type
	contents, err := s.px.ExtractContents(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to find content for hash '%s'", hash)
	}
	contentType := http.DetectContentType(contents)
	if contentType == "" {
		return nil, fmt.Errorf("unknown content type for document '%s'", hash)
	}
	l.Infow("object retrieved and content type detected",
		"content_type", contentType)

	var (
		meta     []string
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
			return nil, err
		}
		meta = s.ta.Summarize(text, 0.25)
	default:
		var parsed2 = strings.FieldsFunc(contentType, func(r rune) bool { return (r == '/') })
		if parsed2 == nil || len(parsed2) == 0 {
			return nil, fmt.Errorf("invalid content type '%s'", contentType)
		}
		switch parsed2[0] {
		case "text":
			category = "document"
			meta = s.ta.Summarize(string(contents), 0.25)
		case "image":
			category = "image"

			// categorize
			keyword, err := s.images.Analyze(hash, contents)
			if err != nil {
				l.Warnw("failed to categorize image", "error", err)
				return nil, errors.New("failed to categorize image")
			}

			// grab any text in image
			text, err := s.oc.Analyze(hash, contents, "image")
			if err != nil {
				l.Warnw("failed to OCR image", "error", err)
				meta = []string{keyword}
			} else {
				meta = append(s.ta.Summarize(text, 0.1), keyword)
			}
		default:
			return nil, errors.New("unsupported content type for indexing")
		}
	}

	// clear the stored text so we can parse new text later
	s.ta.Clear()

	return &models.MetaData{
		Summary:  utils.Unique(meta),
		MimeType: contentType,
		Category: category,
	}, nil
}

// Object is the response from a successfuly lens indexing operation
type Object struct {
	ContentHash string    `json:"lens_object_content_hash"`
	LensID      uuid.UUID `json:"lens_id"`
}

// Store is used to store our collected meta data in a formatted object
func (s *Service) Store(name string, meta *models.MetaData) (*Object, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	// store the name (aka, content hash) of the object so we can avoid duplicate
	// processing in the future
	if err := s.search.Put(name, id.Bytes()); err != nil {
		return nil, err
	}

	return s.Update(id, name, meta)
}

// Update is used to update an object
func (s *Service) Update(id uuid.UUID, name string, meta *models.MetaData) (*Object, error) {
	if meta == nil || len(id.String()) < 1 || name == "" {
		return nil, errors.New("invalid input")
	}

	// iterate over the meta data summary, and create keywords if they don't exist
	for _, keyword := range meta.Summary {
		if err := s.updateKeyword(keyword, id); err != nil {
			return nil, fmt.Errorf("failed to update keyword '%s' for '%s': %s",
				keyword, id, err.Error())
		}
	}

	// store a "mapping" of the lens uuid to its corresponding lens object
	object, err := json.Marshal(&models.Object{
		LensID:   id,
		Name:     name,
		MetaData: *meta,
	})
	if err = s.search.Put(id.String(), object); err != nil {
		return nil, fmt.Errorf("failed to store '%s': '%s'", id.String(), err.Error())
	}

	// store the lens object in iPFS
	hash, err := s.ipfs.DagPut(object, "json", "cbor")
	if err != nil {
		return nil, fmt.Errorf("failed to store '%s': %s", id.String(), err.Error())
	}

	return &Object{
		ContentHash: hash,
		LensID:      id,
	}, nil
}

// Get is used to search for an object identifier by key name
func (s *Service) Get(keyname string) ([]byte, error) {
	if has, err := s.search.Has(keyname); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("keyname does not exist")
	}
	return s.search.Get(keyname)
}

// KeywordSearch is used to search by keyword
func (s *Service) KeywordSearch(keywords []string) ([]models.Object, error) {
	return s.search.KeywordSearch(keywords)
}

func (s *Service) updateKeyword(keyword string, objectID uuid.UUID) error {
	if has, err := s.search.Has(keyword); err != nil {
		return err
	} else if !has {
		var key = models.Keyword{
			Name:            keyword,
			LensIdentifiers: []uuid.UUID{objectID},
		}
		kb, err := json.Marshal(&key)
		if err != nil {
			return err
		}
		return s.search.Put(keyword, kb)
	}

	// keyword exists, get the keyword object from the datastore
	kb, err := s.search.Get(keyword)
	if err != nil {
		return err
	}
	var k = models.Keyword{}
	if err = json.Unmarshal(kb, &k); err != nil {
		return err
	}

	// ensure keyword does not already know about the identifier
	for _, id := range k.LensIdentifiers {
		if id == objectID {
			continue
		}
	}

	// update the lens identifiers in the keyword object
	k.LensIdentifiers = append(k.LensIdentifiers, objectID)
	// TODO: add field to model of content hashes that are mapped in the keyword obj
	kb, err = json.Marshal(k)
	if err != nil {
		return err
	}
	return s.search.Put(keyword, kb)
}
