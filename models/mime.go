package models

// MimeType are categories recognized by Lens
type MimeType string

const (
	// MimeTypeUnknown is a catch-all for unspecified object type
	MimeTypeUnknown = "unknown"

	// MimeTypePDF is a pdf document
	MimeTypePDF = "pdf"
	// MimeTypeDocument is a plain-text document
	MimeTypeDocument = "document"
	// MimeTypeImage is an image asset
	MimeTypeImage = "image"
)
