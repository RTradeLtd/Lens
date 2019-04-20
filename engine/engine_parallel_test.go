package engine

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/RTradeLtd/Lens/v2/engine/queue"
	"github.com/RTradeLtd/Lens/v2/models"
)

func TestEngine_parallel(t *testing.T) {
	var l = zaptest.NewLogger(t).Sugar()
	e, err := New(l, Opts{
		StorePath: filepath.Join("tmp", t.Name()),
		Queue: queue.Options{
			Rate:      500 * time.Millisecond,
			BatchSize: 1,
		}})
	if err != nil {
		t.Error("failed to create engine: " + err.Error())
	}
	defer e.Close()
	go e.Run()

	// each case must be able to successfully index the given args.object and content
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
	t.Run("tests", func(t *testing.T) {
		for _, tt := range tests {
			var tcase = tt // copy case
			t.Run("object "+tcase.args.object.Hash, func(t *testing.T) {
				t.Parallel()

				// request index
				if err := e.Index(Document{tcase.args.object, tcase.args.content, true}); err != nil {
					t.Errorf("wanted Index error = false, got %v", err)
				}

				// wait for flush
				time.Sleep(time.Second)

				// we'll be referring to this hash a few times
				var objHash = tcase.args.object.Hash

				// make sure object can be found
				if !e.IsIndexed(objHash) {
					t.Errorf("wanted IsIndexed = true, got false")
				}

				// attempt search
				if res, err := e.Search(context.Background(), Query{
					Text:   tcase.args.content,
					Hashes: []string{objHash},
				}); err == nil && len(res) > 0 {
					if res[0].Hash != objHash {
						t.Errorf("wanted Search to find '%s', but failed (found '%s')",
							objHash, res[0].Hash)
					}
				} else {
					t.Errorf("wanted Search to find '%s', but failed (got err = %v and res = %v",
						objHash, err, res)
				}

				// remove document
				e.Remove(objHash)

				// wait for flush
				time.Sleep(time.Second)

				// make sure object can't be found
				if e.IsIndexed(objHash) {
					t.Errorf("wanted IsIndexed = false, got true")
				}
			})
		}
	})
}
