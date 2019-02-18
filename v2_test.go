package lens

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/status"

	"github.com/RTradeLtd/Lens/engine"
	"github.com/RTradeLtd/Lens/mocks"
	"github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/grpc/lensv2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

func TestNewV2(t *testing.T) {
	var ipfs = &mocks.FakeRTFSManager{}
	var ia = &mocks.FakeTensorflowAnalyzer{}
	if _, err := NewV2(V2Options{}, ipfs, ia, nil); err != nil {
		t.Errorf("NewV2() error = %v", err)
		return
	}
	if v := NewV2WithEngine(V2Options{}, ipfs, ia, &mocks.FakeEngineSearcher{}, nil); v == nil {
		t.Error("NewV2WithEngine() = nil")
		return
	}
}

func TestV2_Index(t *testing.T) {
	type args struct {
		req *lensv2.IndexReq
	}
	type returns struct {
		catAssetPath string
		tensorErr    bool
		indexErr     bool
		isIndexed    bool
	}
	tests := []struct {
		name        string
		args        args
		returns     returns
		wantType    models.MimeType
		wantErrCode codes.Code
	}{
		{"nil request",
			args{nil},
			returns{"", false, false, false},
			"",
			codes.InvalidArgument},
		{"bad type",
			args{&lensv2.IndexReq{
				Type: lensv2.IndexReq_UNKNOWN,
			}},
			returns{"", false, false, false},
			"",
			codes.InvalidArgument},
		{"no content for hash found",
			args{&lensv2.IndexReq{
				Type: lensv2.IndexReq_IPLD,
				Hash: "asdf",
			}},
			returns{"", false, false, false},
			"",
			codes.NotFound},
		{"already indexed",
			args{&lensv2.IndexReq{
				Type: lensv2.IndexReq_IPLD,
				Hash: "asdf",
			}},
			returns{"README.md", false, false, true},
			"",
			codes.FailedPrecondition},
		{"tensor failure",
			args{&lensv2.IndexReq{
				Type: lensv2.IndexReq_IPLD,
				Hash: "asdf",
			}},
			returns{"test/assets/image.jpg", true, false, false},
			"",
			codes.FailedPrecondition}, // TODO: might not be the best code to return
		{"ok: image",
			args{&lensv2.IndexReq{
				Type: lensv2.IndexReq_IPLD,
				Hash: "asdf",
			}},
			returns{"test/assets/image.jpg", false, false, false},
			models.MimeTypeImage,
			codes.OK},
		{"ok: pdf",
			args{&lensv2.IndexReq{
				Type: lensv2.IndexReq_IPLD,
				Hash: "asdf",
			}},
			returns{"test/assets/text.pdf", false, false, false},
			models.MimeTypePDF,
			codes.OK},
		{"ok: document",
			args{&lensv2.IndexReq{
				Type: lensv2.IndexReq_IPLD,
				Hash: "asdf",
			}},
			returns{"README.md", false, false, false},
			models.MimeTypeDocument,
			codes.OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ipfs = &mocks.FakeRTFSManager{}
			var tensor = &mocks.FakeTensorflowAnalyzer{}
			var se = &mocks.FakeEngineSearcher{}
			var v = NewV2WithEngine(V2Options{},
				ipfs,
				tensor,
				se,
				zap.NewNop().Sugar())

			// set up mocks
			ipfs.CatStub = mocks.StubIpfsCat(tt.returns.catAssetPath)
			if tt.returns.tensorErr {
				tensor.AnalyzeReturns("", errors.New("oh no"))
			} else {
				tensor.AnalyzeReturns("test", nil)
			}
			if tt.returns.indexErr {
				se.IndexReturns(errors.New("oh no"))
			}
			se.IsIndexedReturns(tt.returns.isIndexed)

			// execute tests
			got, err := v.Index(context.Background(), tt.args.req)
			if (err != nil) != (tt.wantErrCode != 0) {
				t.Errorf("V2.Index() error = %v, wantErr %v", err, (tt.wantErrCode != 0))
				return
			}

			t.Logf("got response '%+v'", got)
			if tt.wantErrCode == 0 {
				// if an error was not expected, check the returned document
				if got.GetDoc().GetHash() != tt.args.req.GetHash() {
					t.Errorf("got hash %s, want %s",
						got.GetDoc().GetHash(), tt.args.req.GetHash())
				}
				if got.GetDoc().GetDisplayName() != tt.args.req.GetDisplayName() {
					t.Errorf("got display name %s, want %s",
						got.GetDoc().GetDisplayName(), tt.args.req.GetDisplayName())
				}
				if got.GetDoc().GetCategory() != string(tt.wantType) {
					t.Errorf("got category %s, want %s",
						got.GetDoc().GetCategory(), tt.wantType)
				}
			} else {
				// otherwise, check codes are the same
				var s = status.Convert(err)
				t.Logf("got error message '%s'", s.Message())
				if s.Code() != tt.wantErrCode {
					t.Errorf("V2.Index() err code = %s, want %s",
						s.Code().String(), tt.wantErrCode.String())
				}
			}
		})
	}
}

func TestV2_Search(t *testing.T) {
	type args struct {
		req *lensv2.SearchReq
	}
	type returns struct {
		searchReturns []engine.Result
		searchError   error
	}
	tests := []struct {
		name        string
		args        args
		returns     returns
		wantErrCode codes.Code
	}{
		{"nil request",
			args{nil},
			returns{[]engine.Result{}, nil},
			codes.InvalidArgument},
		{"no query, no options",
			args{&lensv2.SearchReq{}},
			returns{[]engine.Result{}, nil},
			codes.InvalidArgument},
		{"search error",
			args{&lensv2.SearchReq{
				Query: "cats",
			}},
			returns{nil, errors.New("oh no")},
			codes.Internal},
		{"ok: no results",
			args{&lensv2.SearchReq{
				Query: "cats",
			}},
			returns{[]engine.Result{}, nil},
			0},
		{"ok: nil results",
			args{&lensv2.SearchReq{
				Query: "cats",
			}},
			returns{nil, nil},
			0},
		{"ok: with results",
			args{&lensv2.SearchReq{
				Query: "cats",
			}},
			returns{[]engine.Result{{Hash: "asdf"}}, nil},
			0},
		{"ok: with options",
			args{&lensv2.SearchReq{
				Query: "cats",
				Options: &lensv2.SearchReq_Options{
					Hashes: []string{"asdf"},
				}}},
			returns{[]engine.Result{{Hash: "asdf"}}, nil},
			0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ipfs = &mocks.FakeRTFSManager{}
			var tensor = &mocks.FakeTensorflowAnalyzer{}
			var se = &mocks.FakeEngineSearcher{}
			var v = NewV2WithEngine(V2Options{},
				ipfs,
				tensor,
				se,
				zap.NewNop().Sugar())

			// set up mocks
			se.SearchReturns(tt.returns.searchReturns, tt.returns.searchError)

			// execute tests
			got, err := v.Search(context.Background(), tt.args.req)
			if (err != nil) != (tt.wantErrCode != 0) {
				t.Errorf("V2.Search() error = %v, wantErr %v", err, (tt.wantErrCode != 0))
				return
			}

			t.Logf("got response '%+v'", got)
			if tt.wantErrCode == 0 {
				// if an error was not expected, check the returned document
				if got.GetResults() == nil {
					t.Errorf("V2.Search() docs = nil, want not nil")
				}
			} else {
				// otherwise, check codes are the same
				var s = status.Convert(err)
				t.Logf("got error message '%s'", s.Message())
				if s.Code() != tt.wantErrCode {
					t.Errorf("V2.Search() err code = %s, want %s",
						s.Code().String(), tt.wantErrCode.String())
				}
			}
		})
	}
}

func TestV2_Remove(t *testing.T) {
	type args struct {
		req *lensv2.RemoveReq
	}
	type returns struct {
		exists bool
	}
	tests := []struct {
		name        string
		args        args
		returns     returns
		wantErrCode codes.Code
	}{
		{"nil request",
			args{nil},
			returns{true},
			codes.InvalidArgument},
		{"no hash",
			args{&lensv2.RemoveReq{}},
			returns{true},
			codes.InvalidArgument},
		{"not indexed",
			args{&lensv2.RemoveReq{
				Hash: "asdf",
			}},
			returns{false},
			codes.NotFound},
		{"ok: indexed",
			args{&lensv2.RemoveReq{
				Hash: "asdf",
			}},
			returns{true},
			0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ipfs = &mocks.FakeRTFSManager{}
			var tensor = &mocks.FakeTensorflowAnalyzer{}
			var se = &mocks.FakeEngineSearcher{}
			var v = NewV2WithEngine(V2Options{},
				ipfs,
				tensor,
				se,
				zap.NewNop().Sugar())

			// set up mocks
			se.IsIndexedReturns(tt.returns.exists)

			// execute tests
			got, err := v.Remove(context.Background(), tt.args.req)
			if (err != nil) != (tt.wantErrCode != 0) {
				t.Errorf("V2.Remove() error = %v, wantErr %v", err, (tt.wantErrCode != 0))
				return
			}

			t.Logf("got response '%+v'", got)
			if tt.wantErrCode == 0 {
				// if an error was not expected, check the returned document
				if got == nil {
					t.Errorf("V2.Remove() = nil, want not nil")
				}
			} else {
				// otherwise, check codes are the same
				var s = status.Convert(err)
				t.Logf("got error message '%s'", s.Message())
				if s.Code() != tt.wantErrCode {
					t.Errorf("V2.Remove() err code = %s, want %s",
						s.Code().String(), tt.wantErrCode.String())
				}
			}
		})
	}
}
