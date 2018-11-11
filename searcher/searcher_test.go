package searcher_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Lens/searcher"
	rtfs "github.com/RTradeLtd/RTFS"
	"github.com/RTradeLtd/config"
)

const (
	testPath    = "testds"
	testCfgPath = "../test/config.json"
)

func TestService(t *testing.T) {
	// load the config object
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	im, err := rtfs.Initialize("", fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port))
	if err != nil {
		t.Fatal(err)
	}
	s, err := searcher.NewService(testPath)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := s.GetEntries()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("failed to find entries")
	}
	if err = s.MigrateEntries(entries, im, true); err != nil {
		t.Fatal(err)
	}
	if err = s.Put("testKey", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	contents, err := s.Get("testKey")
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "hello" {
		t.Fatal("failed to get correct object")
	}
	has, err := s.Has("testKey")
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal("has was false but expected it to be true")
	}
	//TODO: re-enable after updating testds
	t.Skip()
	keywords := []string{"storage"}
	objects, err := s.KeywordSearch(keywords)
	if err != nil {
		t.Fatal(err)
	}
	if len(*objects) == 0 {
		t.Fatal("no hashes recovered")
	}
	fmt.Println("hashes recovered")
	fmt.Println(objects)
}
