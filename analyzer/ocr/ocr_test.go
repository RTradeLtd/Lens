package ocr

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/RTradeLtd/Lens/logs"

	"github.com/otiai10/gosseract"
)

func TestNewAnalyzer(t *testing.T) {
	var l, _ = logs.NewLogger("", true)
	var a = NewAnalyzer("", l)
	if a.Version() != gosseract.Version() {
		t.Errorf("expected version %s, got %s", gosseract.Version(), a.Version())
	}
}

func TestAnalyzer_Parse(t *testing.T) {
	type args struct {
		assetpath string
		filetype  string
	}
	tests := []struct {
		name         string
		args         args
		wantContents []string
		wantErr      bool
	}{
		{"nil asset", args{"", ""}, nil, true},
		{"not an image", args{"../../test/assets/text.pdf", "png"}, nil, true},
		{"text png asset", args{"../../test/assets/text.png", ""},
			[]string{"TECHNOLOGIES LTD"},
			false},
		{"pdf asset that uses to-text", args{"../../test/assets/text.pdf", "pdf"},
			[]string{"A Simple PDF File", "...continued from page 1"},
			false},
		{"pdf asset that uses OCR", args{"../../test/assets/scan.pdf", "pdf"},
			[]string{"Dear Pete", "Probably you have"},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.Open(tt.args.assetpath)
			if f != nil {
				defer f.Close()
			}

			var l, _ = logs.NewLogger("", true)
			var a = NewAnalyzer("", l)

			var start = time.Now()
			gotContents, err := a.Parse(f, tt.args.filetype)
			if (err != nil) != tt.wantErr {
				t.Errorf("Analyzer.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("time lapsed", time.Since(start))

			if tt.wantContents != nil {
				for _, c := range tt.wantContents {
					if !strings.Contains(gotContents, c) {
						t.Errorf("Analyzer.Parse() = '%v', want '%v'", gotContents, c)
					}
				}
			}
		})
	}
}
