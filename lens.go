package lens

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Lens/analyzer/text"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/searcher"
	"github.com/RTradeLtd/Lens/utils"
	"github.com/RTradeLtd/Lens/xtractor/planetary"
	"github.com/RTradeLtd/config"
	"github.com/gofrs/uuid"
	"github.com/ledongthuc/pdf"
)

// Service contains the various components of Lens
type Service struct {
	TA *text.TextAnalyzer
	PX *planetary.Extractor
	SS *searcher.Service
}

// NewService is used to generate our Lens service
func NewService(opts *ConfigOpts, cfg *config.TemporalConfig) (*Service, error) {
	ta := text.NewTextAnalyzer(opts.UseChainAlgorithm)
	px, err := planetary.NewPlanetaryExtractor(cfg)
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
	meta := []string{}
	switch parsed[0] {
	case "text/plain":
		meta = s.TA.Summarize(string(contents), 0.025)
	case "application/pdf":
		if err = ioutil.WriteFile("/tmp/"+contentHash, contents, 0642); err != nil {
			return "", nil, err
		}
		file, reader, err := pdf.Open("/tmp/" + contentHash)
		if err != nil {
			return "", nil, err
		}
		defer file.Close()
		var buf bytes.Buffer
		b, err := reader.GetPlainText()
		if err != nil {
			return "", nil, err
		}
		if _, err := buf.ReadFrom(b); err != nil {
			return "", nil, err
		}
		contentsString := buf.String()
		meta = s.TA.Summarize(contentsString, 0.05)
	default:
		return "", nil, errors.New("unsupported content type for indexing")
	}
	// clear the stored text so we can parse new text later
	s.TA.Clear()
	metadata := &MetaData{
		Summary: utils.Unique(meta),
	}
	return parsed[0], metadata, nil
}

// Store is used to store our collected meta data in a formatted object
func (s *Service) Store(meta *MetaData, name string) (*IndexOperationResponse, error) {
	// generate a uuid for the lens object
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	// create the lens object
	obj := models.Object{
		LensID:   id,
		Name:     name,
		Keywords: meta.Summary,
	}
	// mrshal the lens object
	marshaled, err := json.Marshal(&obj)
	if err != nil {
		return nil, err
	}
	// iterate over the meta data summary
	for _, v := range meta.Summary {
		// check to see if a keyword with this name already exists
		has, err := s.SS.Has(v)
		if err != nil {
			return nil, err
		}
		// if the keyword does not exist, create the keyword object
		if !has {
			keyObj := models.Keyword{
				Name:            v,
				LensIdentifiers: []uuid.UUID{id},
			}
			keyObjMarshaled, err := json.Marshal(&keyObj)
			if err != nil {
				return nil, err
			}
			if err = s.SS.Put(v, keyObjMarshaled); err != nil {
				return nil, err
			}
			continue
		}
		// keyword exists, get the keyword object from the database
		keywordBytes, err := s.SS.Get(v)
		if err != nil {
			return nil, err
		}
		// create a keyword object
		keyword := models.Keyword{}
		// unmarshal into the keyword object
		if err = json.Unmarshal(keywordBytes, &keyword); err != nil {
			return nil, err
		}
		detected := false
		// TODO: this is false logic, we should never see it as its a new uuid
		// instead, we should store a mapping of content_hash -> uuid
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
		// update the lens identifiers in the keyword object
		keyword.LensIdentifiers = append(keyword.LensIdentifiers, id)
		// TODO: add update to "content hashes" that are mapped in the keyword obj
		keywordMarshaled, err := json.Marshal(keyword)
		if err != nil {
			return nil, err
		}
		// put (aka, update) the keyword object
		if err = s.SS.Put(v, keywordMarshaled); err != nil {
			return nil, err
		}
	}
	// store update badgerds with the uuid -> content hash mapping
	if err = s.SS.Put(id.String(), []byte(name)); err != nil {
		return nil, err
	}
	// store the lens object in iPFS
	hash, err := s.PX.Manager.Shell.DagPut(marshaled, "json", "cbor")
	if err != nil {
		return nil, err
	}
	resp := &IndexOperationResponse{
		ContentHash: hash,
		LensID:      id,
	}
	return resp, nil
}

// SearchByKeyName is used to search for an object by key name
func (s *Service) SearchByKeyName(keyname string) ([]byte, error) {
	has, err := s.SS.Has(keyname)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("keyname does not exist")
	}
	return s.SS.Get(keyname)
}
