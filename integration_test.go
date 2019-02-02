package lens

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/search"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/rtfs"
)

const (
	testHash         = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
	testHashPdf      = "QmTbvUMmniE7wUP1ucbtC9s4ree7s8mSiQBt1c4odzKnY4"
	testHashMarkdown = "QmS5yadpmuu5hPz884XoRFnTTTKaTS4GmdJddd7maysznm"
	testHashJpg      = "QmNWaM9vM4LUs8ZUHThAqC3hCHeQF8fYdJhLjJMwzJmzYS"
	defaultConfig    = "test/config.json"
)

func TestLens_Integration(t *testing.T) {
	if os.Getenv("TEST") != "integration" {
		t.Skip("skipping integration test", t.Name())
	}

	// load config, instantiate everything
	cfg, err := config.LoadConfig(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	ipfsAPI := fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port)
	manager, err := rtfs.NewManager(ipfsAPI, "", 1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	var l, _ = logs.NewLogger("", false)
	ia, err := images.NewAnalyzer(images.ConfigOpts{
		ModelLocation: "tmp",
	}, l)
	if err != nil {
		t.Fatal(err)
	}
	searcher, err := search.NewService("/tmp/integration_test/badgerds-lens")
	if err != nil {
		t.Fatal(err)
	}
	defer searcher.Close()
	defer os.RemoveAll("/tmp/integration_test/badgerds-lens")
	service, err := NewServiceV1(ConfigOpts{
		UseChainAlgorithm: true,
	}, *cfg, manager, ia, searcher, l)
	if err != nil {
		t.Fatal(err)
	}

	///////////////////////
	// INTEGRATION TESTS //
	///////////////////////

	// magnify and store object
	metadata, err := service.Magnify(testHash, false)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := service.Store(testHash, metadata)
	if err != nil {
		t.Fatal(err)
	}
	keywordBytes, err := service.search.Get(metadata.Summary[0])
	if err != nil {
		t.Fatal(err)
	}
	keyword := models.Keyword{}
	if err = json.Unmarshal(keywordBytes, &keyword); err != nil {
		t.Fatal(err)
	}

	// query for identifier
	_, err = service.Get("protocols")
	if err != nil {
		t.Fatal(err)
	}

	// retrieve stored object
	var out models.ObjectV1
	if err = service.ipfs.DagGet(resp.ContentHash, &out); err != nil {
		t.Fatal(err)
	}
	if out.LensID != resp.LensID {
		t.Errorf("expected id '%s', found '%s'", resp.LensID, out.LensID)
	}
	_, err = service.ipfs.Cat(out.Name)
	if err != nil {
		t.Fatal(err)
	}

	// test pdf index and storage
	metadata, err = service.Magnify(testHashPdf, false)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = service.Store(testHashPdf, metadata)
	if err != nil {
		t.Fatal(err)
	}

	// try to update existing object and retrieve it from IPFS
	metadata.Category = "bobheadxi"
	o, err := service.Update(resp.LensID, testHashPdf, metadata)
	if err != nil {
		t.Fatalf("failed to update object: %s", err.Error())
	}
	var updated models.ObjectV1
	b, err := service.Get(o.LensID.String())
	if err != nil {
		t.Fatalf("failed to retrieve object: %s", err.Error())
	}
	json.Unmarshal(b, &updated)
	if updated.MetaData.Category != metadata.Category {
		t.Errorf("expected category '%s', found '%s'", metadata.Category, updated.MetaData.Category)
	}
	if updated.LensID != o.LensID {
		t.Errorf("expected id '%s', got '%s'", o.LensID, updated.LensID)
	}
}
