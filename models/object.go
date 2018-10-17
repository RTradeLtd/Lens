package models

import (
	"github.com/gofrs/uuid"
	"github.com/lib/pq"
)

const (
	TypeTextDocument = "text-document"
)

var (
	SupportedTypes = []string{TypeTextDocument}
)

type Object struct {
	Model
	Type          string         `gorm:"type:varchar(255)" sql:"not null" json:"type" form:"type"`
	AbsoluteName  string         `gorm:"type:text" sql:"not null" json:"absolute_name" form:"absolute_name"`
	ReferenceName string         `gorm:"type:text" json:"reference_name" form:"absolute_name"`
	MetaDataID    uuid.UUID      `gorm:"type:uuid"  json:"meta_data_id" form:"meta_data_id"`
	Locations     pq.StringArray `gorm:"type:text[]" json:"locations" json:"locations" form:"locations"`
	Paths         pq.StringArray `gorm:"type:text[]" json:"paths" form:"paths"`
}
