package planetary

import (
	"io/ioutil"

	"github.com/RTradeLtd/RTFS"
	gocid "github.com/ipfs/go-cid"
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

// DecodeStringToCID is a wrapper used to convert a string to a cid object
func (e *Extractor) DecodeStringToCID(contentHash string) (gocid.Cid, error) {
	return gocid.Decode(contentHash)
}

// ExtractContents is used to extract the contents from the ipld object
func (e *Extractor) ExtractContents(contentHash string) ([]byte, error) {
	reader, err := e.Manager.Shell.Cat(contentHash)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}
