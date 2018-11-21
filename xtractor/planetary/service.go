package planetary

import (
	"github.com/RTradeLtd/rtfs"
)

// Extractor is how we grab data from ipld objects
type Extractor struct {
	im rtfs.Manager
}

// NewPlanetaryExtractor is used to generate our IPLD object extractor
func NewPlanetaryExtractor(ipfsManager rtfs.Manager) (*Extractor, error) {
	return &Extractor{
		im: ipfsManager,
	}, nil
}

// ExtractObject is used to extract an IPLD object from a content hash
func (e *Extractor) ExtractObject(contentHash string, out interface{}) error {
	return e.im.DagGet(contentHash, out)
}

// ExtractContents is used to extract the contents from the ipld object
func (e *Extractor) ExtractContents(contentHash string) ([]byte, error) {
	return e.im.Cat(contentHash)
}
