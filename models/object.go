package models

import "github.com/gofrs/uuid"

// ObjectV0 is an old type of lens object
type ObjectV0 struct {
	// LensID is the id of this particular object within the lens system
	LensID uuid.UUID `json:"lens_id"`
	// Name is how you identify the object on it's network. For IPFS/ipld objects, it is the content hash
	Name string `json:"name"`
	// Keywords are they words that when search will reveal this content hash
	Keywords []string `json:"keywords"`
}

// ObjectV1 is a distributed web object (ie, ipld) that has been indexed by lens
type ObjectV1 struct {
	// LensID is the id of this particular object within the lens system
	LensID uuid.UUID `json:"lens_id"`
	// Name is how you identify the object on it's network. For IPFS/ipld objects, it is the content hash
	Name     string     `json:"name"`
	MetaData MetaDataV1 `json:"meta_data"`
}

// ObjectV2 is a distributed web object (ie, ipld)
type ObjectV2 struct {
	// Hash is how you identify the object on its network, ie content hash
	Hash string `json:"content_hash"`

	// MD is metadata associated with the object
	MD MetaDataV2 `json:"meta"`
}
