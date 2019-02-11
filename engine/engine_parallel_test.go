package engine

import (
	"path/filepath"
	"testing"

	"github.com/RTradeLtd/Lens/models"
	"go.uber.org/zap/zaptest"
)

func TestEngine_parallel(t *testing.T) {
	var l = zaptest.NewLogger(t).Sugar()
	e, err := New(l, Opts{"", filepath.Join("tmp", t.Name())})
	if err != nil {
		t.Error("failed to create engine: " + err.Error())
	}
	defer e.Close()
	go e.Run(nil)

	type args struct {
		object  *models.ObjectV2
		content string
	}
	tests := []struct {
		args args
	}{
		{args{&models.ObjectV2{
			Hash: "abcde",
			MD:   models.MetaDataV2{},
		}, "quick brown fox"}},
		{args{&models.ObjectV2{
			Hash: "asdf",
			MD:   models.MetaDataV2{},
		}, "slow white fox"}},
		{args{&models.ObjectV2{
			Hash: "qwewqr",
			MD:   models.MetaDataV2{},
		}, "hungry grey fox"}},
		{args{&models.ObjectV2{
			Hash: "oiuysa",
			MD: models.MetaDataV2{
				DisplayName: "launch pad",
				Category:    "clubs",
				Tags:        []string{"ubc"},
			},
		}, "ubc launch pad"}},
		{args{&models.ObjectV2{
			Hash: "oishii",
			MD: models.MetaDataV2{
				DisplayName: "rtrade",
				Category:    "startup",
				Tags:        []string{"ipfs", "decentralized"},
			},
		}, "rtrade technologies"}},
	}
	for _, tt := range tests {
		tt := tt // copy case
		t.Run("index "+tt.args.object.Hash, func(t *testing.T) {
			t.Parallel()

			// request index
			if err = e.Index(Document{tt.args.object, "", true}); err != nil {
				t.Errorf("wanted Index error = false, got %v", err)
			}

			// make sure object can be found
			if !e.IsIndexed(tt.args.object.Hash) {
				t.Errorf("wanted IsIndexed = true, got false")
			}

			// attempt search

		})
	}
}
