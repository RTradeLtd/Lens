package lens

// MimeType are categories recognized by Lens
type MimeType string

const (
	// MimeTypePDF is a pdf document
	MimeTypePDF = "pdf"
	// MimeTypeDocument is a plain-text document
	MimeTypeDocument = "document"
	// MimeTypeImage is an image asset
	MimeTypeImage = "image"
)
