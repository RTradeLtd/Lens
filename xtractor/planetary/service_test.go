package planetary_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Lens/xtractor/planetary"
)

const (
	testHash = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
)

func TestPlanetaryExtractor(t *testing.T) {
	px, err := planetary.NewPlanetaryExtractor()
	if err != nil {
		t.Fatal(err)
	}
	var out interface{}
	if err = px.ExtractObject(testHash, &out); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", out)
}
