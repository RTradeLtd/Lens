package models

import "github.com/gofrs/uuid"

type MetaData struct {
	Model
	ObjectIdentifiers    []uuid.UUID `gorm:"type:uuid[]" json:"object_identifiers" form:"object_identifiers"`
	Summary              string      `gorm:"type:text" json:"summary" form:"summary"`
	Description          string      `gorm:"type:text" json:"description" form:"description"`
	ReferenceIdentifiers []uuid.UUID `gorm:"type:uuid[]" json:"reference_identifiers" form:"reference_identifiers"`
}
