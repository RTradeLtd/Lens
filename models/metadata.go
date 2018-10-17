package models

import "github.com/gofrs/uuid"

// MetaData is a piece of meta data
type MetaData struct {
	Model
	ObjectIdentifiers    []uuid.UUID `gorm:"type:uuid[]" json:"object_identifiers" form:"object_identifiers"`
	Summary              []string    `gorm:"type:text[]" json:"summary" form:"summary"`
	ReferenceIdentifiers []uuid.UUID `gorm:"type:uuid[]" json:"reference_identifiers" form:"reference_identifiers"`
}
