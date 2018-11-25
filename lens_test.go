// These are high-level integration tests for the entire Lens service.
package lens

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/Lens/mocks"

	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/rtfs"
)

const (
	testHash         = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
	testHashPdf      = "QmTbvUMmniE7wUP1ucbtC9s4ree7s8mSiQBt1c4odzKnY4"
	testHashMarkdown = "QmS5yadpmuu5hPz884XoRFnTTTKaTS4GmdJddd7maysznm"
	testHashJpg      = "QmNWaM9vM4LUs8ZUHThAqC3hCHeQF8fYdJhLjJMwzJmzYS"
	defaultConfig    = "test/config.json"
)

func TestContentTypeDetect_Integration(t *testing.T) {
	if os.Getenv("TEST") != "integration" {
		t.Skip("skipping integration test", t.Name())
	}

	// set up client and lens
	cfg, err := config.LoadConfig(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	ipfsAPI := fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port)
	manager, err := rtfs.NewManager(ipfsAPI, nil, 1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ia, err := images.NewAnalyzer(images.ConfigOpts{
		ModelLocation: "tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(ConfigOpts{
		UseChainAlgorithm: true, DataStorePath: "tmp/badgerds-lens",
	}, *cfg, manager, ia)
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
			contents, err := service.px.ExtractContents(tt.args.contentHash)
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

func TestLens_Integration(t *testing.T) {
	if os.Getenv("TEST") != "integration" {
		t.Skip("skipping integration test", t.Name())
	}

	cfg, err := config.LoadConfig(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}

	ipfsAPI := fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port)
	manager, err := rtfs.NewManager(ipfsAPI, nil, 1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ia, err := images.NewAnalyzer(images.ConfigOpts{
		ModelLocation: "tmp",
	})
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(ConfigOpts{
		UseChainAlgorithm: true, DataStorePath: "/tmp/badgerds-lens",
	}, *cfg, manager, ia)
	if err != nil {
		t.Fatal(err)
	}

	// test hash examination
	metadata, err := service.Magnify(testHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Content-Type", metadata.MimeType)
	t.Log("meta data", metadata)
	resp, err := service.Store(metadata, testHash)
	if err != nil {
		t.Fatal(err)
	}
	keywordBytes, err := service.ss.Get(metadata.Summary[0])
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
	if err = service.ipfs.DagGet(resp.ContentHash, &out); err != nil {
		t.Fatal(err)
	}
	t.Log("showing ipld lens object")
	t.Logf("%+v\n", out)
	t.Log("retrieving content that was indexed")
	content, err := service.ipfs.Cat(out.Name)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(content))
	metadata, err = service.Magnify(testHashPdf)
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

func TestService_Magnify(t *testing.T) {
	type args struct {
		contentHash string
	}
	type returns struct {
		catAssetPath string
		tensorErr    bool
	}
	tests := []struct {
		name     string
		args     args
		returns  returns
		wantMeta *models.MetaData
		wantErr  bool
	}{
		{"bad hash", args{""}, returns{"", false}, nil, true},
		{"already indexed", args{"existing"}, returns{"", false}, nil, true},
		{"tensor failure", args{"test"}, returns{"", true}, nil, true},
		{"ok: pdf",
			args{"test"},
			returns{"test/assets/text.pdf", false},
			&models.MetaData{
				MimeType: "application/pdf",
				Category: "pdf",
				Summary:  []string{"page", "simple"},
			},
			false,
		},
		{"ok: text",
			args{"test"},
			returns{"README.md", false},
			&models.MetaData{
				MimeType: "text/plain; charset=utf-8",
				Category: "document",
				Summary:  []string{"search", "service"},
			},
			false,
		},
		{"ok: image",
			args{"test"},
			returns{"test/assets/image.jpg", false},
			&models.MetaData{
				MimeType: "image/jpeg",
				Category: "image",
				Summary:  []string{"test"}, // mock output
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ipfs = &mocks.FakeManager{}
			var tensor = &mocks.FakeTensorflowAnalyzer{}
			s, err := NewService(ConfigOpts{
				DataStorePath: "tmp",
			}, config.TemporalConfig{},
				ipfs, tensor)
			if err != nil {
				t.Error(err)
				return
			}
			defer s.Close()

			// set up mocks
			s.ss.Put("existing", []byte("asdfasdf"))
			ipfs.CatStub = mocks.StubIpfsCat(tt.returns.catAssetPath)
			if tt.returns.tensorErr {
				tensor.ClassifyReturns("", errors.New("oh no"))
			} else {
				tensor.ClassifyReturns("test", nil)
			}

			// test
			meta, err := s.Magnify(tt.args.contentHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Magnify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// check output categories
			if tt.wantMeta != nil {
				if meta.MimeType != tt.wantMeta.MimeType {
					t.Errorf("expected mime type '%s', got '%s'", tt.wantMeta.MimeType, meta.MimeType)
				}
				if meta.Category != tt.wantMeta.Category {
					t.Errorf("expected category '%s', got '%s'", tt.wantMeta.Category, meta.Category)
				}

				t.Logf("identified summary: %v\n", meta.Summary)
				for _, s := range tt.wantMeta.Summary {
					var found = false
					for _, f := range meta.Summary {
						if s == f {
							found = true
							continue
						}
					}
					if !found {
						t.Errorf("expected '%s' in summary", s)
					}
				}
			}
		})
	}
}
