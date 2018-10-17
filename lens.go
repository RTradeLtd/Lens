package lens

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Lens/analyzer/text"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/storage"
	"github.com/RTradeLtd/Lens/utils"
	"github.com/RTradeLtd/Lens/xtractor/planetary"
)

// Service contains the various components of Lens
type Service struct {
	TA *text.TextAnalyzer
	PX *planetary.Extractor
	SC *storage.Client
}

// NewService is used to generate our Lens service
func NewService(opts *ConfigOpts) (*Service, error) {
	ta := text.NewTextAnalyzer(opts.UseChainAlgorithm)
	px, err := planetary.NewPlanetaryExtractor()
	if err != nil {
		return nil, err
	}
	sc, err := storage.NewStorageClient()
	if err != nil {
		return nil, err
	}
	return &Service{
		TA: ta,
		PX: px,
		SC: sc,
	}, nil
}

// Magnify is used to examine a given content hash, determine if it's parsable
// and returned the summarized meta-data. Returned parameters are in the format of:
// content type, meta-data, error
func (s *Service) Magnify(contentHash string) (string, *MetaData, error) {
	contents, err := s.PX.ExtractContents(contentHash)
	if err != nil {
		return "", nil, nil
	}
	contentType := http.DetectContentType(contents)
	// it will be in the format of `<content-type>; charset=...`
	// we use strings.FieldsFunc to seperate the string, and to be able to exmaine the content type
	parsed := strings.FieldsFunc(contentType, func(r rune) bool {
		if r == ';' {
			return true
		}
		return false
	})
	switch parsed[0] {
	case "text/plain":
		break
	default:
		return "", nil, errors.New("unsupported content type for indexing")
	}
	// grab our meta data
	meta := s.TA.Summarize(string(contents), 0.025)
	// clear the stored text so we can parse new text later
	s.TA.Clear()
	metadata := &MetaData{
		Summary: utils.Unique(meta),
	}
	return parsed[0], metadata, nil
}

// Store is used to store our collected meta data in a formatted object
func (s *Service) Store(meta *MetaData, contentHash string, fullPath string) (string, error) {
	metaModel := models.MetaData{
		Summary: meta.Summary,
	}
	obj := models.Object{
		Type:          models.TypeTextDocument,
		AbsoluteName:  fullPath,
		ReferenceName: contentHash,
		Locations:     []string{"ipfs"},
		Paths:         []string{fullPath},
		MetaData:      metaModel,
	}
	marshaled, err := json.Marshal(&obj)
	if err != nil {
		return "", err
	}
	return s.SC.IPFS.Shell.DagPut(marshaled, "json", "cbor")
}
