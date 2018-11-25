package images_test

import (
	"io/ioutil"
	"testing"

	"github.com/RTradeLtd/Lens/analyzer/images"
)

const (
	testImg = "../../test/assets/image.jpg"
)

func TestTendorize(t *testing.T) {
	analyzer, err := images.NewAnalyzer(images.ConfigOpts{
		ModelLocation: "models",
	})
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(testImg)
	if err != nil {
		t.Fatal(err)
	}

	guess, err := analyzer.Classify(b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(guess)
}
