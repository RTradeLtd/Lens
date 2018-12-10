// These are high-level integration tests for the entire Lens service.
package lens

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/mocks"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/search"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/rtfs"
	"github.com/gofrs/uuid"
)

const (
	testHash         = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
	testHashPdf      = "QmTbvUMmniE7wUP1ucbtC9s4ree7s8mSiQBt1c4odzKnY4"
	testHashMarkdown = "QmS5yadpmuu5hPz884XoRFnTTTKaTS4GmdJddd7maysznm"
	testHashJpg      = "QmNWaM9vM4LUs8ZUHThAqC3hCHeQF8fYdJhLjJMwzJmzYS"
	defaultConfig    = "test/config.json"
)

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

	var l, _ = logs.NewLogger("", false)
	ia, err := images.NewAnalyzer(images.ConfigOpts{
		ModelLocation: "tmp",
	}, l)
	if err != nil {
		t.Fatal(err)
	}

	searcher, err := search.NewService("/tmp/badgerds-lens")
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(ConfigOpts{
		UseChainAlgorithm: true,
	}, *cfg, manager, ia, searcher, l)
	if err != nil {
		t.Fatal(err)
	}

	// test hash examination
	metadata, err := service.Magnify(testHash, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Content-Type", metadata.MimeType)
	t.Log("meta data", metadata)
	resp, err := service.Store(testHash, metadata)
	if err != nil {
		t.Fatal(err)
	}
	keywordBytes, err := service.search.Get(metadata.Summary[0])
	if err != nil {
		t.Fatal(err)
	}
	keyword := models.Keyword{}
	if err = json.Unmarshal(keywordBytes, &keyword); err != nil {
		t.Fatal(err)
	}
	match, err := service.Get("protocols")
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
	metadata, err = service.Magnify(testHashPdf, false)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = service.Store(testHashPdf, metadata)
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
			// setup
			var ipfs = &mocks.FakeManager{}
			var tensor = &mocks.FakeTensorflowAnalyzer{}
			var searcher = &mocks.FakeSearcher{}
			var l, _ = logs.NewLogger("", false)
			s, err := NewService(ConfigOpts{}, config.TemporalConfig{},
				ipfs,
				tensor,
				searcher,
				l)
			if err != nil {
				t.Error(err)
				return
			}
			defer s.Close()

			// set up mocks
			s.search.Put("existing", []byte("asdfasdf"))
			ipfs.CatStub = mocks.StubIpfsCat(tt.returns.catAssetPath)
			if tt.returns.tensorErr {
				tensor.AnalyzeReturns("", errors.New("oh no"))
			} else {
				tensor.AnalyzeReturns("test", nil)
			}

			// test
			meta, err := s.Magnify(tt.args.contentHash, false)
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

func TestService_Update(t *testing.T) {
	// generate uuid for testing
	var id, _ = uuid.NewV4()

	type args struct {
		meta *models.MetaData
		id   uuid.UUID
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *Object
		wantErr bool
	}{
		{"bad input",
			args{nil, uuid.UUID{}, ""},
			nil,
			true},
		{"should attempt to update",
			args{&models.MetaData{
				Summary:  []string{"test"},
				MimeType: "blah",
				Category: "blah",
			}, id, "test_object"},
			nil,
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			var ipfs = &mocks.FakeManager{}
			var tensor = &mocks.FakeTensorflowAnalyzer{}
			var searcher = &mocks.FakeSearcher{}
			var l, _ = logs.NewLogger("", false)
			s, err := NewService(ConfigOpts{}, config.TemporalConfig{},
				ipfs,
				tensor,
				searcher,
				l)
			if err != nil {
				t.Error(err)
				return
			}
			defer s.Close()

			_, err = s.Update(tt.args.id, tt.args.name, tt.args.meta)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// check if correct data was attempted to be stored. skip if this is an
			// error case
			if tt.wantErr {
				return
			}
			var o models.Object
			var call = 1
			key, val := searcher.PutArgsForCall(call)
			t.Logf("input for call %d: '%s', '%s'", call, key, string(val))
			if err := json.Unmarshal(val, &o); err != nil {
				t.Errorf("unexpected input for searcher.Put call %d: %s", call, err.Error())
				return
			}
			if !reflect.DeepEqual(&o.MetaData, tt.args.meta) {
				t.Errorf("got = %v, wanted %v", o, tt.args.meta)
				return
			}
			if tt.args.id != o.LensID {
				t.Errorf("got = %v, wanted %v", o.LensID, tt.args.id)
				return
			}
		})
	}
}
