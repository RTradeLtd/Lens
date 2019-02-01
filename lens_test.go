// These are high-level integration tests for the entire Lens service.
package lens

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/mocks"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/search"
	"github.com/RTradeLtd/config"
	"github.com/gofrs/uuid"
)

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
		wantMeta *models.MetaDataV1
		wantErr  bool
	}{
		{"bad hash", args{""}, returns{"", false}, nil, true},
		{"already indexed", args{"existing"}, returns{"", false}, nil, true},
		{"tensor failure", args{"test"}, returns{"", true}, nil, true},
		{"ok: pdf",
			args{"test"},
			returns{"test/assets/text.pdf", false},
			&models.MetaDataV1{
				MimeType: "application/pdf",
				Category: "pdf",
				Summary:  []string{"page", "simple"},
			},
			false,
		},
		{"ok: text",
			args{"test"},
			returns{"README.md", false},
			&models.MetaDataV1{
				MimeType: "text/plain; charset=utf-8",
				Category: "document",
				Summary:  []string{"search", "service"},
			},
			false,
		},
		{"ok: image",
			args{"test"},
			returns{"test/assets/image.jpg", false},
			&models.MetaDataV1{
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
			var l, _ = logs.NewLogger("", false)
			searcher, err := search.NewService("/tmp/badgerds-lens")
			if err != nil {
				t.Fatal(err)
			}
			defer searcher.Close()
			s, err := NewServiceV1(ConfigOpts{}, config.TemporalConfig{},
				ipfs,
				tensor,
				searcher,
				l)
			if err != nil {
				t.Error(err)
				return
			}

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
		meta *models.MetaDataV1
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
			args{&models.MetaDataV1{
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
			s, err := NewServiceV1(ConfigOpts{}, config.TemporalConfig{},
				ipfs,
				tensor,
				searcher,
				l)
			if err != nil {
				t.Error(err)
				return
			}

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
			var o models.ObjectV1
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
