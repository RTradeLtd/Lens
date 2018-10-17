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
