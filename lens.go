package lens

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Lens/analyzer/text"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/searcher"
	"github.com/RTradeLtd/Lens/storage"
	"github.com/RTradeLtd/Lens/utils"
	"github.com/RTradeLtd/Lens/xtractor/planetary"
	"github.com/gofrs/uuid"
)

// Service contains the various components of Lens
type Service struct {
	TA *text.TextAnalyzer
	PX *planetary.Extractor
	SC *storage.Client
	SS *searcher.Service
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
	ss, err := searcher.NewService(opts.DataStorePath)
	if err != nil {
		return nil, err
	}
	return &Service{
		TA: ta,
		PX: px,
		SC: sc,
		SS: ss,
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
func (s *Service) Store(meta *MetaData, name string) (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	obj := models.Object{
		LensID:   id,
		Name:     name,
		Keywords: meta.Summary,
	}
	marshaled, err := json.Marshal(&obj)
	if err != nil {
		return "", err
	}
	for _, v := range meta.Summary {
		has, err := s.SS.Has(v)
		if err != nil {
			return "", err
		}
		if !has {
			if err = s.SS.Put(v, marshaled); err != nil {
				return "", err
			}
			continue
		}
		keywordBytes, err := s.SS.Get(v)
		if err != nil {
			return "", err
		}
		keyword := models.Keyword{}
		if err = json.Unmarshal(keywordBytes, &keyword); err != nil {
			return "", err
		}
		detected := false
		for _, v := range keyword.LensIdentifiers {
			if v == id {
				detected = true
				break
			}
		}
		if detected {
			// this object has already  been indexed for the particular keyword, so we can skip
			continue
		}
		keyword.LensIdentifiers = append(keyword.LensIdentifiers, id)
		keywordMarshaled, err := json.Marshal(keyword)
		if err != nil {
			return "", err
		}
		if err = s.SS.Put(v, keywordMarshaled); err != nil {
			return "", err
		}
	}
	resp, err := s.SC.IPFS.Shell.DagPut(marshaled, "json", "cbor")
	if err != nil {
		return "", err
	}
	return resp, nil
}
