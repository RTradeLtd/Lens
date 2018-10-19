package searcher_test

import (
	"testing"

	"github.com/RTradeLtd/Lens/searcher"
)

const (
	testPath = "testds"
)

func TestService(t *testing.T) {
	s, err := searcher.NewService(testPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Put("testKey", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Get("testKey"); err != nil {
		t.Fatal(err)
	}
}
