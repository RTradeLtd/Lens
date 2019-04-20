package engine

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/RTradeLtd/Lens/v2/engine/queue"
	"github.com/RTradeLtd/Lens/v2/models"
)

func TestEngine_Index(t *testing.T) {
	type args struct {
		object  *models.ObjectV2
		reindex bool
	}
	tests := []struct {
		name        string
		args        args
		wantIndexed bool
	}{
		{"no hash",
			args{&models.ObjectV2{
				MD: models.MetaDataV2{},
			}, false},
			false,
		},
		{"ok",
			args{&models.ObjectV2{
				Hash: "abcde",
				MD:   models.MetaDataV2{},
			}, false},
			true,
		},
		{"with tags and some metadata",
			args{&models.ObjectV2{
				Hash: "oishii",
				MD: models.MetaDataV2{
					DisplayName: "rtrade",
					Category:    "startup",
					Tags:        []string{"ipfs", "decentralized"},
				},
			}, true},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l = zaptest.NewLogger(t).Sugar()
			e, err := New(l, Opts{
				StorePath: filepath.Join("tmp", t.Name()),
				Queue: queue.Options{
					Rate:      500 * time.Millisecond,
					BatchSize: 1,
				}})
			if err != nil {
				t.Error("failed to create engine: " + err.Error())
				return
			}
			defer func() {
				e.Close()
				os.RemoveAll("tmp")
			}()
			go e.Run()

			// insert bogus doc if testing reindex
			if tt.args.reindex {
				var bogus = *tt.args.object
				bogus.MD.Tags = []string{"i", "am", "fake"}
				if err = e.Index(Document{&bogus, "", true}); (err == nil) != tt.wantIndexed {
					t.Errorf("wanted Index error = %v, got %v", !tt.wantIndexed, err)
					return
				}
				time.Sleep(time.Second)
			}

			// request index
			if err = e.Index(Document{tt.args.object, "", tt.args.reindex}); (err == nil) != tt.wantIndexed {
				t.Errorf("wanted Index error = %v, got %v", !tt.wantIndexed, err)
			}
			t.Logf("object index requested, got error = %v", err)
			time.Sleep(time.Second)

			// make sure object can be found (or can't)
			var found bool
			if found = e.IsIndexed(tt.args.object.Hash); found != tt.wantIndexed {
				t.Errorf("wanted IsIndexed = '%v', got '%v'", tt.wantIndexed, found)
			}

			// run additional check on actual stored object if wantIndexed
			if tt.wantIndexed {
				r, err := e.Search(context.Background(), Query{Hashes: []string{tt.args.object.Hash}})
				if err != nil {
					t.Errorf("wanted Search err = nil, got '%v'", err)
					return
				}
				if len(r) < 1 {
					t.Errorf("could not find object '%s' in search", tt.args.object.Hash)
					return
				}
				if !reflect.DeepEqual(r[0].MD, tt.args.object.MD) {
					t.Errorf("Engine.Search() = %v, want %v", r[0].MD, tt.args.object.MD)
				}
			}
		})
	}
}

func TestEngine_Search(t *testing.T) {
	var testContent = `You are currently using an enterprise storage solution powered by
			Temporal, an API built for the Interplanetary File System. This platform
			showcases the outstanding features that decentralized storage technologies
			can offer you.`
	var testObj = models.ObjectV2{
		Hash: "abcde",
		MD: models.MetaDataV2{
			DisplayName: "my test object!",
			MimeType:    "text",
			Category:    "amazing startup",
			Tags:        []string{"test", "object"},
		},
	}

	// not testing indexing capabilities, so we can share an instance
	var l = zaptest.NewLogger(t).Sugar()
	e, err := New(l, Opts{
		StorePath: filepath.Join("tmp", t.Name()),
		Queue: queue.Options{
			Rate:      500 * time.Millisecond,
			BatchSize: 1,
		}})
	if err != nil {
		t.Error("failed to create engine: " + err.Error())
		return
	}
	defer func() {
		e.Close()
		os.RemoveAll("tmp")
	}()
	go e.Run()

	// store test object in engine
	e.Index(Document{&testObj, testContent, true})
	time.Sleep(time.Second)

	type args struct {
		q Query
	}
	tests := []struct {
		name    string
		args    args
		wantDoc bool
	}{
		{"ok: find test obj with hash",
			args{Query{
				// Needs text - hashes are only provided as a filtering option
				Text:   "Interplanetary File System",
				Hashes: []string{testObj.Hash},
			}},
			true},
		{"fail: do NOT find test obj with wrong hash filter",
			args{Query{
				Text:   "Interplanetary File System",
				Hashes: []string{"not_my_hash"},
			}},
			false},
		{"ok: find test obj with subtext",
			args{Query{
				Text: "Interplanetary File System",
			}},
			true},
		{"ok: find test obj with exact text",
			args{Query{
				Text: testContent,
			}},
			true},
		{"fail: do NOT find test obj with wrong text",
			args{Query{
				Text: "robert is the best!",
			}},
			false},
		{"ok: find test obj with required text",
			args{Query{
				Required: []string{"Interplanetary"},
			}},
			true},
		{"ok: find test obj with required text separated",
			args{Query{
				Required: []string{" API   ", "Interplanetary    File   System", "outstanding features", "   "},
			}},
			true},
		{"fail: do NOT find test obj without required text",
			args{Query{
				Required: []string{"ubc launch pad"},
			}},
			false},
		{"ok: find test obj with mime type",
			args{Query{
				MimeTypes: []string{testObj.MD.MimeType},
			}},
			true},
		{"fail: do NOT find test obj without mime type",
			args{Query{
				MimeTypes: []string{models.MimeTypeUnknown},
			}},
			false},
		{"ok: find test obj with category",
			args{Query{
				Categories: []string{testObj.MD.Category},
			}},
			true},
		{"fail: do NOT find test obj without category",
			args{Query{
				Categories: []string{"amazing"},
			}},
			false},
		{"ok: find test obj with tag",
			args{Query{
				Tags: []string{testObj.MD.Tags[0]},
			}},
			true},
		{"fail: do NOT find test obj without tag",
			args{Query{
				Tags: []string{"kfc"},
			}},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set new logger for each test for cleanliness
			e.l = zaptest.NewLogger(t).Sugar()

			// attempt to search for object
			got, err := e.Search(context.Background(), tt.args.q)
			if err != nil && tt.wantDoc {
				t.Error("got error: " + err.Error())
				return
			}

			// check for document
			if tt.wantDoc {
				if len(got) < 1 {
					t.Error("got no results")
					return
				}
				if got[0].Hash != testObj.Hash {
					t.Errorf("Engine.Search() = %s, want %s", got[0].Hash, testObj)
				}
				if !reflect.DeepEqual(got[0].MD, testObj.MD) {
					t.Errorf("Engine.Search() = %v, want %v", got[0].MD, testObj.MD)
				}
			}
		})
	}
}
