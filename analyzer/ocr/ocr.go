package ocr

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/otiai10/gosseract"
)

// Analyzer is the OCR analysis class
type Analyzer struct {
	config string
}

// NewAnalyzer creates a new OCR analyzer
func NewAnalyzer(config string) *Analyzer {
	return &Analyzer{config}
}

// Version reports the version of Tesseract
func (a *Analyzer) Version() string { return gosseract.Version() }

// Parse executes OCR on text
func (a *Analyzer) Parse(asset io.Reader, assetType string) (contents string, err error) {
	if asset == nil {
		return "", errors.New("invalid asset provided")
	}

	b, err := ioutil.ReadAll(asset)
	if err != nil {
		return "", fmt.Errorf("failed to read asset: %s", err.Error())
	}

	// preprocessing
	switch assetType {
	case "pdf":
		if b, err = pdfToImage(b); err != nil {
			return "", fmt.Errorf("failed to convert PDF: %s", err.Error())
		}
	default:
	}

	t, err := a.newTesseractClient()
	if err != nil {
		return "", err
	}
	defer t.Close()

	if err = t.SetImageFromBytes(b); err != nil {
		return "", fmt.Errorf("failed to set image: %s", err.Error())
	}

	if contents, err = t.Text(); err != nil {
		err = fmt.Errorf("failed to convert asset to text: %s", err.Error())
	}
	return
}

func (a *Analyzer) newTesseractClient() (*gosseract.Client, error) {
	t := gosseract.NewClient()
	if a.config != "" {
		if err := t.SetConfigFile(a.config); err != nil {
			return nil, fmt.Errorf("failed to set tesseract configuration file at '%s': %s",
				a.config, err.Error())
		}
	}
	return t, nil
}
