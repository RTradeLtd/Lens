package images_test

import (
	"io/ioutil"
	"testing"

	"github.com/RTradeLtd/Lens/v2/analyzer/images"
	"go.uber.org/zap/zaptest"
)

const (
	testImg = "../../test/assets/image.jpg"
)

func TestTendorize(t *testing.T) {
	var l = zaptest.NewLogger(t)
	analyzer, err := images.NewAnalyzer(images.ConfigOpts{
		ModelLocation: "models",
	}, l.Sugar())
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
