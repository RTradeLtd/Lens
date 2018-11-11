package searcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/RTFS"
	"github.com/gofrs/uuid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/ipfs/go-ds-badger"
)

// Service is our searcher service
type Service struct {
	DS *badger.Datastore
}

// NewService is used to generate our new searcher service
func NewService(dsPath string) (*Service, error) {
	ds, err := badger.NewDatastore(dsPath, &badger.DefaultOptions)
	if err != nil {
		return nil, err
	}
	return &Service{
		DS: ds,
	}, nil
}

// GetEntries is used to get all known entries
func (s *Service) GetEntries() ([]query.Entry, error) {
	resp, err := s.DS.Query(query.Query{})
	if err != nil {
		log.Fatal(err)
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
		_, err := uuid.FromString(split[1])
		if err != nil {
			// make sure we don't process ipfs hashes
			if split[1][0] == 'Q' {
				continue
			}
			keywordBytes, err := s.Get(split[1])
			if err != nil {
				continue
			}
			if err = json.Unmarshal(keywordBytes, &keyword); err != nil {
				continue
			}
			for _, v := range keyword.LensIdentifiers {
				// we have to use split[1] as the `name` field was improperly set to an ipfs hash
				keywords[split[1]] = append(keywords[split[1]], v)
				ids = append(ids, v)
			}
			continue
		}
		//ids = append(ids, id)
	}
	fmt.Println(keywords)
	// rebuild our keywords and properly migrate them
	for name, identifiers := range keywords {
		key := models.Keyword{
			Name:            name,
			LensIdentifiers: identifiers,
		}
		// marshal the keyword
		keyBytes, err := json.Marshal(&key)
		if err != nil {
			// dont hard fail
			continue
		}
		if err = s.Put(name, keyBytes); err != nil {
			continue
		}
		for _, id := range identifiers {
			// get the data it refers to
			bytes, err := s.Get(id.String())
			if err != nil {
				// dont hard fail
				continue
			}
			// if we can't unmarshal into new object, then it's al "old" object and refers to the content hash
			var obj models.Object
			if err = json.Unmarshal(bytes, &obj); err == nil {
				// get the new object
				bytes, err = s.Get(id.String())
				if err != nil {
					// dont hard fail
					continue
				}
				if err = json.Unmarshal(bytes, &obj); err != nil {
					// dont hard fail
					continue
				}
				obj.MetaData.Summary = append(obj.MetaData.Summary, name)
				// marshal the object
				bytes, err = json.Marshal(&obj)
				if err != nil {
					// dont hard fail
					continue
				}
				// update the new object
				if err = s.Put(id.String(), bytes); err != nil {
					// dont hard fail
					fmt.Println("error migrating lens object ", err)
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
			// marshal the new object
			bytes, err = json.Marshal(&obj)
			if err != nil {
				// don't hard fail
				fmt.Println("error marshaling new object ", err)
				continue
			}
			// store the new object
			if err = s.Put(id.String(), bytes); err != nil {
				fmt.Println("failed to store new object ", err)
				continue
			}
		}
	}
	if migrateContent && im != nil {
		processedIds := make(map[string]bool)
		// update the category
		for _, id := range ids {
			if processedIds[id.String()] {
				fmt.Println("id already processed ", id)
				continue
			}
			processedIds[id.String()] = true
			// getthe data from IPFs so we can sniff its mime-type
			// grab the object from the database
			objBytes, err := s.Get(id.String())
			if err != nil {
				// don't hard fail
				fmt.Println("failed to get lens object ", err)
				continue
			}
			var obj models.Object
			if err = json.Unmarshal(objBytes, &obj); err != nil {
				fmt.Println("failed to unmarshal lens object ", err)
				// don't hard fail
				continue
			}
			reader, err := im.Shell.Cat(obj.Name)
			if err != nil {
				fmt.Println("failed to get content from ipfs ", err)
				continue
			}
			dataBytes, err := ioutil.ReadAll(reader)
			if err != nil {
				fmt.Println("failed to get bytes from reader ", err)
				continue
			}
			contentType := http.DetectContentType(dataBytes)
			split := strings.Split(contentType, ";")
			// update the content type
			obj.MetaData.MimeType = split[0]
			if contentType == "application/pdf" {
				obj.MetaData.Category = "pdf"
			} else {
				split := strings.Split(contentType, "/")
				obj.MetaData.Category = split[0]
			}
			// marshal the newly update object
			objBytes, err = json.Marshal(&obj)
			if err != nil {
				// dont hard fail
				fmt.Println("failed to marshal updated object ", err)
				continue
			}
			fmt.Printf("%+v\n", obj)
			if err = s.Put(id.String(), objBytes); err != nil {
				fmt.Println("failed to store updated object ", err)
				continue
			}
		}
	}
	return nil
}

// Put is used to store something in badgerds
func (s *Service) Put(keyName string, data []byte) error {
	k := ds.NewKey(keyName)
	return s.DS.Put(k, data)
}

// Get is used to retrieve something from badgerds by key name
func (s *Service) Get(keyName string) ([]byte, error) {
	k := ds.NewKey(keyName)
	return s.DS.Get(k)
}

// Has is used to check if our database has a key matching this name
func (s *Service) Has(keyName string) (bool, error) {
	k := ds.NewKey(keyName)
	return s.DS.Has(k)
}

// KeywordSearch retrieves a slice of content hashes that were indexed with these keywords
func (s *Service) KeywordSearch(keywords []string) (*[]models.Object, error) {
	// ids are a list of id's for which this keyword matched
	ids := []uuid.UUID{}
	// usedIDs represetn lens identifiers we've already searched
	usedIDs := make(map[uuid.UUID]bool)
	fmt.Println("searching through keywords")
	// search through all keywords
	for _, v := range keywords {
		// check if we have content for this keyword
		has, err := s.Has(v)
		if err != nil {
			return nil, err
		}
		// if we don't skip it
		if !has {
			continue
		}
		fmt.Println("valid keyword")
		// get the keyword object from the datastore
		resp, err := s.Get(v)
		if err != nil {
			return nil, err
		}
		// prepare  messsage to unmarshal the data into
		k := models.Keyword{}
		// unmarshal into keyword object
		if err = json.Unmarshal(resp, &k); err != nil {
			return nil, err
		}
		fmt.Printf("%+v\n", k)
		// search through all the identifiers
		for _, id := range k.LensIdentifiers {
			// if the id is not nil, and we haven't seen this id already
			if id != uuid.Nil && !usedIDs[id] {
				// add it to the list of IDs to process
				ids = append(ids, id)
				usedIDs[id] = true
			}
		}
	}
	var (
		objects []models.Object
		object  models.Object
	)
	fmt.Println("searching for ids")
	for _, v := range ids {
		has, err := s.Has(v.String())
		if err != nil {
			return nil, err
		}
		if !has {
			return nil, errors.New("no entry with lens id found")
		}
		objectBytes, err := s.Get(v.String())
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(objectBytes, &object); err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	return &objects, nil
}

// FilterOpts is used to configure an advanced search
type FilterOpts struct {
	Categories  []string
	Keywords    []string
	SearchDepth int
}

// AdvancedSearch is used to perform an advanced search against the lens index
func (s *Service) AdvancedSearch(opts *FilterOpts) (*[]models.Object, error) {
	var (
		idsToSearch []string
		category    models.Category
		obj         models.Object
	)

	// search through categories, and get a list of IDs to search for
	for _, v := range opts.Categories {
		has, err := s.Has(v)
		if err != nil {
			// dont hard fail
			continue
		}
		if !has {
			// dont hard fail
			continue
		}
		categoryBytes, err := s.Get(v)
		if err != nil {
			// dont hard fail
			continue
		}
		if err = json.Unmarshal(categoryBytes, &category); err != nil {
			// dont hard fail
			continue
		}
		idsToSearch = append(idsToSearch, category.ObjectIdentifiers...)
	}
	var (
		matchedObjects []models.Object
		matched        = make(map[uuid.UUID]bool)
	)
	for _, v := range idsToSearch {
		id, err := uuid.FromString(v)
		if err != nil {
			// dont hard fail
			continue
		}
		has, err := s.Has(id.String())
		if err != nil {
			// dont hard fail
			continue
		}
		if !has {
			// dont hard fail
			continue
		}
		objBytes, err := s.Get(id.String())
		if err != nil {
			// dont hard fail
			continue
		}
		if err = json.Unmarshal(objBytes, &obj); err != nil {
			// dont hard fail
			continue
		}
		for _, keyword := range opts.Keywords {
			for i := 0; i < opts.SearchDepth; i++ {
				if keyword == obj.MetaData.Summary[i] {
					if matched[id] {
						continue
					}
					matchedObjects = append(matchedObjects, obj)
				}
			}
		}
	}
	return nil, nil
}
