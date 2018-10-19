package lens

import "github.com/gofrs/uuid"

// ConfigOpts are options used to configure the lens service
type ConfigOpts struct {
	UseChainAlgorithm bool
	DataStorePath     string
	API               struct {
		IP   string
		Port string
	}
}

// MetaData is a piece of meta data from a given object after being lensed
type MetaData struct {
	Summary []string `json:"summary"`
}

// IndexOperationResponse is the response from a successfuly lens indexing operation
type IndexOperationResponse struct {
	ContentHash string    `json:"lens_object_content_hash"`
	LensID      uuid.UUID `json:"lens_id"`
}
