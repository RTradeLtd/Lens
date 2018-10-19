package lens_test

import (
	"encoding/json"
	"fmt"
	"testing"

	lens "github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/models"
)

const (
	testHash = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
)

func TestLens(t *testing.T) {
	cfg := &lens.ConfigOpts{UseChainAlgorithm: true, DataStorePath: "/tmp/badgerds-lens"}
	service, err := lens.NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	contentType, metadata, err := service.Magnify(testHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Content-Type ", contentType)
	fmt.Println("meta data ", metadata)
	resp, err := service.Store(metadata, testHash)
	if err != nil {
		t.Fatal(err)
	}
	keywordBytes, err := service.SS.Get(metadata.Summary[0])
	if err != nil {
		t.Fatal(err)
	}
	keyword := models.Keyword{}
	if err = json.Unmarshal(keywordBytes, &keyword); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", keyword)
	fmt.Println("Meta data collection IFPS hash ", resp)
}
