// These are high-level integration tests for the entire Lens service.
package lens_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/config"
)

const (
	testHash         = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
	testHashPdf      = "QmTbvUMmniE7wUP1ucbtC9s4ree7s8mSiQBt1c4odzKnY4"
	testHashMarkdown = "QmS5yadpmuu5hPz884XoRFnTTTKaTS4GmdJddd7maysznm"
	testHashJpg      = "QmNWaM9vM4LUs8ZUHThAqC3hCHeQF8fYdJhLjJMwzJmzYS"
	defaultConfig    = "test/config.json"
)

func TestContentTypeDetect(t *testing.T) {
	if os.Getenv("TEST") != "integration" {
		t.Skip("skipping integration test", t.Name())
	}

	// set up client and lens
	cfg, err := config.LoadConfig(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	service, err := lens.NewService(&lens.ConfigOpts{
		UseChainAlgorithm: true, DataStorePath: "tmp/badgerds-lens",
	}, cfg)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		contentHash string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantType string
	}{
		{"pdf", args{testHashPdf}, false, "pdf"},
		{"markdown", args{testHashMarkdown}, false, "markdown"},
		{"jpg", args{testHashJpg}, false, "jpg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("retrieving %s", tt.args.contentHash)
			contents, err := service.PX.ExtractContents(tt.args.contentHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// check content type
			contentType := http.DetectContentType(contents)
			t.Logf("content type: %s", contentType)
			if contentType != tt.wantType {
				t.Errorf("wanted %s, got %s", tt.wantType, contentType)
			}
		})
	}
}

func TestLens(t *testing.T) {
	if os.Getenv("TEST") != "integration" {
		t.Skip("skipping integration test", t.Name())
	}

	cfg, err := config.LoadConfig(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	opts := &lens.ConfigOpts{UseChainAlgorithm: true, DataStorePath: "/tmp/badgerds-lens"}
	service, err := lens.NewService(opts, cfg)
	if err != nil {
		t.Fatal(err)
	}

	// test hash examination
	contentType, metadata, err := service.Magnify(testHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Content-Type", contentType)
	t.Log("meta data", metadata)
	resp, err := service.Store(metadata, testHash)
	if err != nil {
		t.Fatal(err)
	}
	keywordBytes, err := service.SS.Get(metadata.Summary[0])
	if err != nil {
		t.Fatal(err)
	}
	keyword := models.Keyword{}
	if err = json.Unmarshal(keywordBytes, &keyword); err != nil {
		t.Fatal(err)
	}
	match, err := service.SearchByKeyName("protocols")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("match found", string(match))
	t.Log("hash of indexed object ", resp)

	var out models.Object
	if err = service.PX.Manager.Shell.DagGet(resp.ContentHash, &out); err != nil {
		t.Fatal(err)
	}
	t.Log("showing ipld lens object")
	t.Logf("%+v\n", out)
	t.Log("retrieving content that was indexed")
	reader, err := service.PX.Manager.Shell.Cat(out.Name)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	contentBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(contentBytes))
	contentType, metadata, err = service.Magnify(testHashPdf)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = service.Store(metadata, testHashPdf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("pdf processing response")
	t.Logf("%+v\n", resp)
}
