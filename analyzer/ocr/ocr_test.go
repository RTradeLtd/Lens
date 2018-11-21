package ocr

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/otiai10/gosseract"
)

func TestNewAnalyzer(t *testing.T) {
	a := NewAnalyzer("")
	if a.Version() != gosseract.Version() {
		t.Errorf("expected version %s, got %s", gosseract.Version(), a.Version())
	}
}

func TestAnalyzer_Parse(t *testing.T) {
	png, err := os.Open("../../test/assets/text.png")
	if err != nil {
		t.Fatal(err)
	}

	// disable for now
	pdf, err := os.Open("../../test/assets/sample.pdf")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		asset    io.Reader
		filetype string
	}
	tests := []struct {
		name         string
		args         args
		wantContents string
		wantErr      bool
	}{
		{"nil asset", args{nil, ""}, "", true},
		{"text png asset", args{png, ""}, "TECHNOLOGIES LTD", false},
		{"pdf asset", args{pdf, "pdf"}, "TECHNOLOGIES LTD", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnalyzer("")

			gotContents, err := a.Parse(tt.args.asset, tt.args.filetype)
			if (err != nil) != tt.wantErr {
				t.Errorf("Analyzer.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !strings.Contains(gotContents, tt.wantContents) {
				t.Errorf("Analyzer.Parse() = %v, want %v", gotContents, tt.wantContents)
			}
		})
	}
}
