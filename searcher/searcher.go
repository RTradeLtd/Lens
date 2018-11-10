package searcher

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RTradeLtd/Lens/models"
	"github.com/gofrs/uuid"
	ds "github.com/ipfs/go-datastore"
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
