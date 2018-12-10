package server

import (
	"context"
	"testing"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/mocks"
	"github.com/RTradeLtd/config"
	pbreq "github.com/RTradeLtd/grpc/lens/request"
	"github.com/gofrs/uuid"
)

func TestAPIServer_Index(t *testing.T) {
	// generate uuid for testing
	var id, _ = uuid.NewV4()

	type returns struct {
		searchHas    bool
		catAssetPath string
		knownID      *uuid.UUID
	}
	type args struct {
		req *pbreq.Index
	}
	tests := []struct {
		name    string
		args    args
		returns returns
		wantErr bool
	}{
		{"invalid type",
			args{&pbreq.Index{Type: "not_ipld"}},
			returns{},
			true},
		{"existing object without reindex",
			args{&pbreq.Index{
				Type:    "ipld",
				Reindex: false,
			}},
			returns{
				searchHas: true,
			},
			true},
		{"reindex with unknown object",
			args{&pbreq.Index{
				Type:       "ipld",
				Identifier: "test_unknown_obj_reindex",
				Reindex:    true,
			}},
			returns{
				catAssetPath: "../README.md",
			},
			true},
		{"attempting to update without forced reindex",
			args{&pbreq.Index{
				Type:       "ipld",
				Identifier: "test_known_obj_update_without_reindex",
				Reindex:    false,
			}},
			returns{
				knownID:      &id,
				catAssetPath: "../README.md",
			},
			true},
		{"ok (text, no reindex)",
			args{&pbreq.Index{
				Type:       "ipld",
				Identifier: "test_store",
				Reindex:    false,
			}},
			returns{
				catAssetPath: "../README.md",
			},
			false},
		{"ok (text, with reindex)",
			args{&pbreq.Index{
				Type:       "ipld",
				Identifier: "test_reindex",
				Reindex:    true,
			}},
			returns{
				knownID:      &id,
				catAssetPath: "../README.md",
			},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up dependencies
			var ipfs = &mocks.FakeManager{}
			var tensor = &mocks.FakeTensorflowAnalyzer{}
			var searcher = &mocks.FakeSearcher{}
			var l, _ = logs.NewLogger("", false)
			ls, err := lens.NewService(lens.ConfigOpts{}, config.TemporalConfig{},
				ipfs,
				tensor,
				searcher,
				l)
			if err != nil {
				t.Error(err)
				return
			}
			defer ls.Close()
			as := &APIServer{
				lens: ls,
				l:    l,
			}

			// set up mock returns
			if tt.returns.searchHas {
				searcher.HasReturns(true, nil)
			}
			if tt.returns.catAssetPath != "" {
				ipfs.CatStub = mocks.StubIpfsCat(tt.returns.catAssetPath)
			}
			if tt.returns.knownID != nil {
				searcher.HasStub = func(key string) (bool, error) {
					return tt.args.req.GetIdentifier() == key, nil
				}
				searcher.GetReturns(tt.returns.knownID.Bytes(), nil)
			}

			// run test
			_, err = as.Index(context.Background(), tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("APIServer.Index() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
