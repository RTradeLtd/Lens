package searcher

import (
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
