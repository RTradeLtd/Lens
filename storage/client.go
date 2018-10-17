package storage

import rtfs "github.com/RTradeLtd/RTFS"

// Client is our single client for our storage interface
type Client struct {
	IPFS *rtfs.IpfsManager
}

// NewStorageClient is used to generate our storage client
func NewStorageClient() (*Client, error) {
	ipfsManager, err := rtfs.Initialize("", "")
	if err != nil {
		return nil, err
	}
	return &Client{
		IPFS: ipfsManager,
	}, nil
}
