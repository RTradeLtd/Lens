package images_test

import (
	"io/ioutil"
	"testing"

	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/Lens/logs"
)

const (
	testImg = "../../test/assets/image.jpg"
)

func TestTendorize(t *testing.T) {
	var l, _ = logs.NewLogger("", true)
	analyzer, err := images.NewAnalyzer(images.ConfigOpts{
		ModelLocation: "models",
	}, l)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(testImg)
	if err != nil {
		t.Fatal(err)
	}

	guess, err := analyzer.Analyze("test", b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(guess)
}
