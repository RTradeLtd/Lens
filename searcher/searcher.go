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
func (s *Service) KeywordSearch(keywords []string) ([]string, error) {
	ids := []uuid.UUID{}
	usedIDs := make(map[uuid.UUID]bool)
	fmt.Println("searching through keywords")
	for _, v := range keywords {
		has, err := s.Has(v)
		if err != nil {
			return nil, err
		}
		if !has {
			// this keyword does not exist lets skip it
			continue
			// return nil, errors.New("keyword does not exist")
		}
		fmt.Println("valid keyword")
		resp, err := s.Get(v)
		if err != nil {
			return nil, err
		}
		k := models.Keyword{}
		if err = json.Unmarshal(resp, &k); err != nil {
			return nil, err
		}
		fmt.Printf("%+v\n", k)
		for _, id := range k.LensIdentifiers {
			if id != uuid.Nil && !usedIDs[id] {
				ids = append(ids, id)
				usedIDs[id] = true
			}
		}
	}
	hashes := []string{}
	usedHashes := make(map[string]bool)
	for _, v := range ids {
		hashBytes, err := s.Get(v.String())
		if err != nil {
			return nil, err
		}
		hash := string(hashBytes)
		if hash != "" && !usedHashes[hash] {
			hashes = append(hashes, hash)
			usedHashes[hash] = true
		}
	}
	if len(hashes) == 0 {
		return nil, errors.New("no hashes found")
	}
	return hashes, nil
}
