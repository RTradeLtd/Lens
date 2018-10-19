package lens_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	match, err := service.SearchByKeyName("protocols")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("match found ", string(match))
	fmt.Println("hash of indexed object ", resp)
	var out models.Object
	if err = service.PX.Manager.Shell.DagGet(resp.ContentHash, &out); err != nil {
		t.Fatal(err)
	}
	fmt.Println("showing ipld lens object")
	fmt.Printf("%+v\n", out)
	fmt.Println("retrieving content that was indexed")
	reader, err := service.PX.Manager.Shell.Cat(out.Name)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	contentBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(contentBytes))
}
