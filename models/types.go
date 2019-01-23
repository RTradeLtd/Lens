package models

import "github.com/gofrs/uuid"

// MetaDataV0 is an old metadata object
type MetaDataV0 struct {
	Summary []string `json:"summary"`
}

// MetaDataV1 is a piece of meta data from a given object after being lensed
type MetaDataV1 struct {
	Summary  []string `json:"summary"`
	MimeType string   `json:"mime_type"`
	Category string   `json:"category"`
}

// Category is a particular search category, such as document, pdf, etc..
type Category struct {
	Name string `json:"name"`
	// Identifiers are id's of indexed lens object which match this category
	Identifiers []uuid.UUID `json:"object_identifiers"`
}

// MetaDataV2 is a piece of meta data from a given object after being lensed
type MetaDataV2 struct {
	DisplayName string   `json:"display_name"`
	MimeType    string   `json:"mime_type"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}
