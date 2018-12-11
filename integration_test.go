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

	cfg, err := config.LoadConfig(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}

	ipfsAPI := fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port)
	manager, err := rtfs.NewManager(ipfsAPI, nil, 1*time.Minute)
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

	searcher, err := search.NewService("/tmp/badgerds-lens")
	if err != nil {
		t.Fatal(err)
	}
	defer searcher.Close()

	service, err := NewService(ConfigOpts{
		UseChainAlgorithm: true,
	}, *cfg, manager, ia, searcher, l)
	if err != nil {
		t.Fatal(err)
	}

	// test hash examination
	metadata, err := service.Magnify(testHash, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Content-Type", metadata.MimeType)
	t.Log("meta data", metadata)
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
	match, err := service.Get("protocols")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("match found", string(match))
	t.Log("hash of indexed object ", resp)

	var out models.Object
	if err = service.ipfs.DagGet(resp.ContentHash, &out); err != nil {
		t.Fatal(err)
	}
	t.Log("showing ipld lens object")
	t.Logf("%+v\n", out)
	t.Log("retrieving content that was indexed")
	content, err := service.ipfs.Cat(out.Name)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(content))
	metadata, err = service.Magnify(testHashPdf, false)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = service.Store(testHashPdf, metadata)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("pdf processing response")
	t.Logf("%+v\n", resp)
}
