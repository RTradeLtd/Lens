package models

// ObjectV2 is a distributed web object (ie, ipld)
type ObjectV2 struct {
	// Hash is how you identify the object on its network, ie content hash
	Hash string `json:"content_hash"`

	// MD is metadata associated with the object
	MD MetaDataV2 `json:"meta"`
}
