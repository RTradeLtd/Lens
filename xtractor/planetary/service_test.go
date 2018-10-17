package planetary_test

import (
	"testing"

	"github.com/RTradeLtd/Lens/xtractor/planetary"
)

func TestPlanetaryExtractor(t *testing.T) {
	if _, err := planetary.NewPlanetaryExtractor(); err != nil {
		t.Fatal(err)
	}
}
