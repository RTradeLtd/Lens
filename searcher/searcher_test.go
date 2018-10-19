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
}
