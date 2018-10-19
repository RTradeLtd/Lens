package storage_test

import (
	"testing"

	"github.com/RTradeLtd/Lens/storage"
)

func TestStorageClient(t *testing.T) {
	client, err := storage.NewStorageClient()
	if err != nil {
		t.Fatal(err)
	}
	if client.IPFS.KeystoreEnabled {
		t.Fatal("keystore should not be enabled")
	}
}
