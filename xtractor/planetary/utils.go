package planetary

import gocid "github.com/ipfs/go-cid"

// DecodeStringToCID is a wrapper used to convert a string to a cid object
func DecodeStringToCID(contentHash string) (gocid.Cid, error) {
	return gocid.Decode(contentHash)
}
