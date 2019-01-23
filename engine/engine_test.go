package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/models"
)

func TestEngine_Index(t *testing.T) {
	type args struct {
		object *models.ObjectV2
	}
	tests := []struct {
		name        string
		args        args
		wantIndexed bool
	}{
		{"ok",
			args{&models.ObjectV2{
				Hash: "abcde",
				MD:   models.MetaDataV2{},
			}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l, out = logs.NewTestLogger()
			defer t.Logf("Logs: \n%v\n", out.All())
			e, err := New(l, Opts{"tmp", "tmp"})
			if err != nil {
				t.Error("failed to create engine: " + err.Error())
			}
			defer e.Close()
			go e.Run(time.Millisecond)
			e.Index(Document{tt.args.object, "", true})
			time.Sleep(time.Millisecond)
			if found := e.IsIndexed(tt.args.object.Hash); found != tt.wantIndexed {
				t.Errorf("wanted IsIndexed = '%v', got '%v'", tt.wantIndexed, found)
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
			MimeType: "text",
			Category: "",
			Tags:     []string{"test", "object"},
		},
	}
	type fields struct {
		content string
		object  models.ObjectV2
	}
	type args struct {
		q Query
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantDoc bool
	}{
		{"find test obj with text",
			fields{testContent, testObj},
			args{Query{
				Text: "Interplanetary File System",
			}},
			true},
		{"find test obj with mime type",
			fields{testContent, testObj},
			args{Query{
				MimeTypes: []string{testObj.MD.MimeType},
			}},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l, _ = logs.NewLogger("", true)
			// defer t.Logf("Logs: \n%v\n", out.All())
			e, err := New(l, Opts{"tmp", "tmp"})
			if err != nil {
				t.Error("failed to create engine: " + err.Error())
			}
			defer e.Close()
			go e.Run(time.Millisecond)

			e.Index(Document{&tt.fields.object, tt.fields.content, true})
			time.Sleep(time.Millisecond)

			got, err := e.Search(tt.args.q)
			if err != nil && tt.wantDoc {
				t.Error("got error: " + err.Error())
			}

			if tt.wantDoc {
				if len(got) < 1 {
					t.Error("got no results")
				}
				if got[0].Hash != tt.fields.object.Hash {
					t.Errorf("Engine.Search() = %s, want %s", got[0].Hash, tt.fields.object.Hash)
				}
				if !reflect.DeepEqual(got[0].MD, tt.fields.object.MD) {
					t.Errorf("Engine.Search() = %v, want %v", got[0].MD, tt.fields.object.MD)
				}
			}
		})
	}
}
