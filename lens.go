package lens

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/Lens/analyzer/ocr"
	"github.com/RTradeLtd/rtfs"

	"github.com/RTradeLtd/Lens/analyzer/text"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/search"
	"github.com/RTradeLtd/Lens/utils"
	"github.com/RTradeLtd/Lens/xtractor/planetary"
	"github.com/RTradeLtd/config"
	"github.com/gofrs/uuid"
)

// Service contains the various components of Lens
type Service struct {
	ipfs   rtfs.Manager
	images images.TensorflowAnalyzer

	oc *ocr.Analyzer
	ta *text.Analyzer
	px *planetary.Extractor
	ss *search.Service
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

// IndexOperationResponse is the response from a successfuly lens indexing operation
type IndexOperationResponse struct {
	ContentHash string    `json:"lens_object_content_hash"`
	LensID      uuid.UUID `json:"lens_id"`
}

// NewService is used to generate our Lens service
func NewService(opts ConfigOpts, cfg config.TemporalConfig,
	rm rtfs.Manager,
	ia images.TensorflowAnalyzer,
	logger *zap.SugaredLogger) (*Service, error) {
	// instantiate utility classes
	px, err := planetary.NewPlanetaryExtractor(rm)
	if err != nil {
		return nil, err
	}

	// instantiate service
	ss, err := search.NewService(opts.DataStorePath)
	if err != nil {
		return nil, err
	}

	return &Service{
		ta:     text.NewTextAnalyzer(opts.UseChainAlgorithm),
		oc:     ocr.NewAnalyzer(opts.TesseractConfigPath, logger.Named("ocr")),
		images: ia,
		px:     px,
		ss:     ss,
	}, nil
}

// Close releases resources held by the service
func (s *Service) Close() error {
	return s.ss.Close()
}

// Magnify is used to examine a given content hash, determine if it's parsable
// and returned the summarized meta-data. Returned parameters are in the format of:
// content type, meta-data, error
func (s *Service) Magnify(hash string) (metadata *models.MetaData, err error) {
	if has, err := s.ss.Has(hash); err != nil {
		return nil, err
	} else if has {
		return nil, errors.New("this object has already been indexed")
	}

	// retrieve object and detect content type
	contents, err := s.px.ExtractContents(hash)
	if err != nil {
		return nil, err
	}
	contentType := http.DetectContentType(contents)
	if contentType == "" {
		return nil, fmt.Errorf("unknown content type for document '%s'", hash)
	}

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
		text, err := s.oc.Parse(contents, "pdf")
		if err != nil {
			return nil, err
		}
		meta = s.ta.Summarize(text, 0.25)
	default:
		var parsed2 = strings.FieldsFunc(contentType, func(r rune) bool { return (r == '/') })
		if parsed2 == nil || len(parsed2) == 0 {
			err = fmt.Errorf("invalid content type '%s'", contentType)
			return
		}
		switch parsed2[0] {
		case "text":
			category = "document"
			meta = s.ta.Summarize(string(contents), 0.25)
		case "image":
			category = "image"
			keyword, err := s.images.Classify(contents)
			if err != nil {
				return nil, err
			}
			meta = []string{keyword}
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

// Store is used to store our collected meta data in a formatted object
func (s *Service) Store(meta *models.MetaData, name string) (*IndexOperationResponse, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	// iterate over the meta data summary, and create keywords if they don't exist
	for _, keyword := range meta.Summary {
		if err := s.updateKeyword(keyword, id); err != nil {
			return nil, err
		}
	}

	// store the name (aka, content hash) of the object so we can avoid duplicate processing in the future
	if err = s.ss.Put(name, []byte(id.String())); err != nil {
		return nil, err
	}

	// store a "mapping" of the lens uuid to its corresponding lens object
	object, err := json.Marshal(&models.Object{
		LensID:   id,
		Name:     name,
		MetaData: *meta,
	})
	if err = s.ss.Put(id.String(), object); err != nil {
		return nil, err
	}

	// store the lens object in iPFS
	hash, err := s.ipfs.DagPut(object, "json", "cbor")
	if err != nil {
		return nil, err
	}

	return &IndexOperationResponse{
		ContentHash: hash,
		LensID:      id,
	}, nil
}

// SearchByKeyName is used to search for an object by key name
func (s *Service) SearchByKeyName(keyname string) ([]byte, error) {
	if has, err := s.ss.Has(keyname); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("keyname does not exist")
	}
	return s.ss.Get(keyname)
}

// KeywordSearch is used to search by keyword
func (s *Service) KeywordSearch(keywords []string) ([]models.Object, error) {
	return s.ss.KeywordSearch(keywords)
}

func (s *Service) updateKeyword(keyword string, objectID uuid.UUID) error {
	if has, err := s.ss.Has(keyword); err != nil {
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
		return s.ss.Put(keyword, kb)
	}

	// keyword exists, get the keyword object from the datastore
	kb, err := s.ss.Get(keyword)
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
	return s.ss.Put(keyword, kb)
}
