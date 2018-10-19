package models

import "github.com/gofrs/uuid"

// Keyword is a keyword object stored in badgerds
// It contains a list of LensIdentifiers, which are the identifiers to a object (ie, ipld) that has been indexed by lens
type Keyword struct {
	Name string `json:"name"`
	// LensIdentifiers are the uuids of objects indexed by lens with this particular keyword
	LensIdentifiers []uuid.UUID `json:"lens_identifiers"`
}
