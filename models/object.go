package models

import "github.com/gofrs/uuid"

// Object is a distributed web object (ie, ipld) that has been indexed by lens
type Object struct {
	LensID   uuid.UUID `json:"lens_id"`
	Keywords []string  `json:"keywords"`
}
