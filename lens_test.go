package lens_test

import (
	"fmt"
	"testing"

	lens "github.com/RTradeLtd/Lens"
)

const (
	testHash = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
)

func TestLens(t *testing.T) {
	cfg := &lens.ConfigOpts{UseChainAlgorithm: true}
	service, err := lens.NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	contentType, metadata, err := service.Magnify(testHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Content-Type ", contentType)
	fmt.Printf("metadata\n%+v\n", metadata)
}
