package search

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/rtfs"
	"github.com/gofrs/uuid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	badger "github.com/ipfs/go-ds-badger"
)

// Service is our searcher service
type Service struct {
	ds *badger.Datastore
}

// NewService is used to generate our new searcher service
func NewService(dsPath string) (*Service, error) {
	ds, err := badger.NewDatastore(dsPath, &badger.DefaultOptions)
	if err != nil {
		return nil, err
	}
	return &Service{
		ds: ds,
	}, nil
}

// Close shuts down the service datastore
func (s *Service) Close() error { return s.ds.Close() }

// GetEntries is used to get all known entries
func (s *Service) GetEntries() ([]query.Entry, error) {
	resp, err := s.ds.Query(query.Query{})
	if err != nil {
		return nil, err
	}
	return resp.Rest()
}

// MigrateEntries is used to migrate entries to newer object types
func (s *Service) MigrateEntries(entries []query.Entry, im *rtfs.IpfsManager, migrateContent bool) error {
	var (
		ids      []uuid.UUID
		keywords = make(map[string][]uuid.UUID)
		keyword  models.Keyword
	)
	// extract id's to mgirate
	for _, v := range entries {
		split := strings.Split(v.Key, "/")
		if _, err := uuid.FromString(split[1]); err != nil {
			if split == nil || len(split) < 2 || len(split[1]) == 0 {
				continue
			}

			if split[1][0] == 'Q' {
				continue
			}
			b, err := s.Get(split[1])
			if err != nil {
				continue
			}
			if err = json.Unmarshal(b, &keyword); err != nil {
				continue
			}
			for _, v := range keyword.LensIdentifiers {
				// we have to use split[1] as the `name` field was improperly set to an
				// ipfs hash
				keywords[split[1]] = append(keywords[split[1]], v)
				ids = append(ids, v)
			}
			continue
		}
	}
	// rebuild our keywords and properly migrate them
	for name, identifiers := range keywords {
		key := models.Keyword{
			Name:            name,
			LensIdentifiers: identifiers,
		}
		b, err := json.Marshal(&key)
		if err != nil {
			continue
		}
		if err = s.Put(name, b); err != nil {
			continue
		}

		for _, id := range identifiers {
			bytes, err := s.Get(id.String())
			if err != nil {
				continue
			}
			// if we can't unmarshal into new object, then it's an "old" object and
			// refers to the content hash
			var obj models.Object
			if err = json.Unmarshal(bytes, &obj); err == nil {
				bytes, err = s.Get(id.String())
				if err != nil {
					continue
				}
				if err = json.Unmarshal(bytes, &obj); err != nil {
					continue
				}
				obj.MetaData.Summary = append(obj.MetaData.Summary, name)
				bytes, err = json.Marshal(&obj)
				if err != nil {
					continue
				}
				// update the new object
				if err = s.Put(id.String(), bytes); err != nil {
					continue
				}
				continue
			}

			// format the meta-data
			meta := models.MetaData{
				Summary: []string{name},
			}
			obj.LensID = id
			obj.MetaData = meta
			obj.Name = string(bytes)

			// store the new object
			bytes, err = json.Marshal(&obj)
			if err != nil {
				continue
			}
			if err = s.Put(id.String(), bytes); err != nil {
				continue
			}
		}
	}

	if migrateContent && im != nil {
		processedIds := make(map[string]bool)
		// update the category
		for _, id := range ids {
			if processedIds[id.String()] {
				continue
			}
			processedIds[id.String()] = true
			// get the data from IPFs so we can sniff its mime-type
			// grab the object from the database
			b, err := s.Get(id.String())
			if err != nil {
				continue
			}
			var obj models.Object
			if err = json.Unmarshal(b, &obj); err != nil {
				continue
			}
			bytes, err := im.Cat(obj.Name)
			if err != nil {
				continue
			}

			// update the content type
			contentType := http.DetectContentType(bytes)
			if split := strings.Split(contentType, ";"); split != nil && len(split) > 0 {
				obj.MetaData.MimeType = split[0]
				obj.MetaData.Category = split[0]
			}
			if contentType == "application/pdf" {
				obj.MetaData.Category = "pdf"
			}

			// update object
			b, err = json.Marshal(&obj)
			if err != nil {
				continue
			}
			if err = s.Put(id.String(), b); err != nil {
				continue
			}
		}
	}
	return nil
}

// Put is used to store something in badgerds
func (s *Service) Put(keyName string, data []byte) error {
	return s.ds.Put(ds.NewKey(keyName), data)
}

// Get is used to retrieve something from badgerds by key name
func (s *Service) Get(keyName string) ([]byte, error) {
	return s.ds.Get(ds.NewKey(keyName))
}

// Has is used to check if our database has a key matching this name
func (s *Service) Has(keyName string) (bool, error) {
	return s.ds.Has(ds.NewKey(keyName))
}

// KeywordSearch retrieves a slice of content hashes that were indexed with these keywords
func (s *Service) KeywordSearch(keywords []string) ([]models.Object, error) {
	var (
		matches = make([]uuid.UUID, 0)
		visited = make(map[uuid.UUID]bool)
	)

	for _, v := range keywords {
		if has, err := s.Has(v); err != nil {
			return nil, err
		} else if !has {
			continue
		}

		// get the keyword object from the datastore
		resp, err := s.Get(v)
		if err != nil {
			return nil, err
		}
		var k = models.Keyword{}
		if err = json.Unmarshal(resp, &k); err != nil {
			return nil, err
		}

		// search through all the identifiers for matches
		for _, id := range k.LensIdentifiers {
			if id != uuid.Nil && !visited[id] {
				matches = append(matches, id)
				visited[id] = true
			}
		}
	}

	var (
		objects []models.Object
		object  models.Object
	)
	for _, v := range matches {
		if has, err := s.Has(v.String()); err != nil {
			return nil, err
		} else if !has {
			return nil, errors.New("no entry with lens id found")
		}

		// retrieve object
		b, err := s.Get(v.String())
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(b, &object); err != nil {
			return nil, err
		}

		objects = append(objects, object)
	}

	return objects, nil
}

// FilterOpts is used to configure an advanced search
type FilterOpts struct {
	Categories  []string
	Keywords    []string
	SearchDepth int
}

// AdvancedSearch is used to perform an advanced search against the lens index
func (s *Service) AdvancedSearch(opts *FilterOpts) ([]models.Object, error) {
	// search through categories, and get a list of IDs to search for
	var (
		query    []string
		category models.Category
		obj      models.Object
	)
	for _, v := range opts.Categories {
		if has, err := s.Has(v); err != nil || !has {
			continue
		}

		// retrieve category
		b, err := s.Get(v)
		if err != nil {
			continue
		}
		if err = json.Unmarshal(b, &category); err != nil {
			continue
		}

		for _, id := range category.Identifiers {
			query = append(query, id.String())
		}
	}

	var (
		matched = make([]models.Object, 0)
		visited = make(map[uuid.UUID]bool)
	)
	for _, v := range query {
		id, err := uuid.FromString(v)
		if err != nil {
			continue
		}
		if has, err := s.Has(id.String()); err != nil || !has {
			continue
		}

		// retrieve object matching query
		b, err := s.Get(id.String())
		if err != nil {
			continue
		}
		if err = json.Unmarshal(b, &obj); err != nil {
			continue
		}

		// check if object contains keyword
		for _, keyword := range opts.Keywords {
			for i := 0; i < opts.SearchDepth; i++ {
				if keyword == obj.MetaData.Summary[i] {
					if visited[id] {
						continue
					}
					matched = append(matched, obj)
				}
			}
		}
	}

	return matched, nil
}
