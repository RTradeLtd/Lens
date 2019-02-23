package models

// MetaDataV2 is a piece of meta data from a given object after being lensed
type MetaDataV2 struct {
	DisplayName string   `json:"display_name"`
	MimeType    string   `json:"mime_type"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}
