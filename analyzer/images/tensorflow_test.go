package images_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Lens/analyzer/images"
)

const (
	testImg = "../../test/assets/image.jpg"
)

func TestTendorize(t *testing.T) {
	opts := &images.ConfigOpts{
		ModelLocation: "models",
	}
	analyzer, err := images.NewAnalyzer(opts)
	if err != nil {
		t.Fatal(err)
	}
	guess, err := analyzer.ClassifyImage(testImg)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(guess)
}
