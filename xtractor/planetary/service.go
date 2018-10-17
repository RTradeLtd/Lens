package planetary

import (
	"github.com/RTradeLtd/Temporal/rtfs"
)

// Extractor is how we grab data from ipld objects
type Extractor struct {
	Manager *rtfs.IpfsManager
}

// NewPlanetaryExtractor is used to generate our IPLD object extractor
func NewPlanetaryExtractor() (*Extractor, error) {
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		return nil, err
	}
	return &Extractor{
		Manager: manager,
	}, nil
}

// ExtractObject is used to extract an IPLD object from a content hash
func (e *Extractor) ExtractObject(contentHash string, out interface{}) error {
	return e.Manager.Shell.DagGet(contentHash, out)
}
