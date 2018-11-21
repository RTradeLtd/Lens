package images_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/RTradeLtd/Lens/analyzer/images"
)

const (
	testImg = "test.jpg"
)

func TestTendorize(t *testing.T) {
	opts := &images.ConfigOpts{
		ModelLocation: "models",
	}
	analyzer, err := images.NewAnalyzer(opts)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(testImg)
	if err != nil {
		t.Fatal(err)
	}

	guess, err := analyzer.ClassifyImage(b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(guess)
}
