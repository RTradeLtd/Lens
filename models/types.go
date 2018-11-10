package models

// MetaData is a piece of meta data from a given object after being lensed
type MetaData struct {
	Summary  []string `json:"summary"`
	MimeType string   `json:"mime_type"`
	Category string   `json:"category"`
}

// Category is a particular search category, such as document, pdf, etc..
type Category struct {
	Name string `json:"name"`
	// ObjectIdentifiers are distributed web object identifiers such as IPFS content hashes
	ObjectIdentifiers []string `json:"object_identifiers"`
}
