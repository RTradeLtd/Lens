package lens_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	lens "github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/config"
)

const (
	testHash         = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
	testHashPdf      = "QmTbvUMmniE7wUP1ucbtC9s4ree7s8mSiQBt1c4odzKnY4"
	testHashMarkdown = "QmS5yadpmuu5hPz884XoRFnTTTKaTS4GmdJddd7maysznm"
	testHashJpg      = "QmXQBRL6JJQGEqL3L2wsBdZS2MLfZ2GFa6rrorSXtqF7GM"
	defaultConfig    = "test/config.json"
)

func TestContentTypeDetect(t *testing.T) {
	//	t.Skip()
	cfg, err := config.LoadConfig(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	opts := &lens.ConfigOpts{UseChainAlgorithm: true, DataStorePath: "/tmp/badgerds-lens"}
	service, err := lens.NewService(opts, cfg)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := service.PX.ExtractContents(testHashPdf)
	if err != nil {
		t.Fatal(err)
	}
	contentType := http.DetectContentType(contents)
	fmt.Println("content type of pdf")
	fmt.Println(contentType)
	contents, err = service.PX.ExtractContents(testHashMarkdown)
	if err != nil {
		t.Fatal(err)
	}
	contentType = http.DetectContentType(contents)
	fmt.Println("content type of markdown")
	fmt.Println(contentType)
	contents, err = service.PX.ExtractContents(testHashJpg)
	if err != nil {
		t.Fatal(err)
	}
	contentType = http.DetectContentType(contents)
	fmt.Println("content type of jpg")
	fmt.Println(contentType)
}

func TestLens(t *testing.T) {
	cfg, err := config.LoadConfig(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	opts := &lens.ConfigOpts{UseChainAlgorithm: true, DataStorePath: "/tmp/badgerds-lens"}
	service, err := lens.NewService(opts, cfg)
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
	contentType, metadata, err = service.Magnify(testHashPdf)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = service.Store(metadata, testHashPdf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("pdf processing response")
	fmt.Printf("%+v\n", resp)
}
