package search_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/RTradeLtd/Lens/search"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/rtfs"
)

const (
	testPath    = "../test/ds"
	testCfgPath = "../test/config.json"
)

func TestService(t *testing.T) {
	cfg, err := config.LoadConfig(testCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	im, err := rtfs.NewManager(
		fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port),
		1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	s, err := search.NewService(testPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
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
		t.Fatal("expected 'testKey' to be present")
	}

	//TODO: re-enable after updating testds
	t.Skip()
	keywords := []string{"storage"}
	objects, err := s.KeywordSearch(keywords)
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) == 0 {
		t.Fatal("no hashes recovered")
	}
	t.Log("hashes recovered")
	t.Log(objects)
}
