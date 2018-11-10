package models

import "github.com/gofrs/uuid"

// Object is a distributed web object (ie, ipld) that has been indexed by lens
type Object struct {
	// LensID is the id of this particular object within the lens system
	LensID uuid.UUID `json:"lens_id"`
	// Name is how you identify the object on it's network. For IPFS/ipld objects, it is the content hash
	Name     string   `json:"name"`
	MetaData MetaData `json:"meta_data"`
}
