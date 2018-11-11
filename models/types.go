package models

import "github.com/gofrs/uuid"

// MetaData is a piece of meta data from a given object after being lensed
type MetaData struct {
	Summary  []string `json:"summary"`
	MimeType string   `json:"mime_type"`
	Category string   `json:"category"`
}

// Category is a particular search category, such as document, pdf, etc..
type Category struct {
	Name string `json:"name"`
	// LensIdentifiers are id's of indexed lens object which match this category
	LensIdentifiers []uuid.UUID `json:"object_identifiers"`
}

// MetaDataOld is an old metadata object
type MetaDataOld struct {
	Summary []string `json:"summary"`
}
