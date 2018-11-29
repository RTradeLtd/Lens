package ocr

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"

	"go.uber.org/zap"

	fitz "github.com/gen2brain/go-fitz"
	"github.com/otiai10/gosseract"
)

// Analyzer is the OCR analysis class
type Analyzer struct {
	configPath string

	l *zap.SugaredLogger
}

// NewAnalyzer creates a new OCR analyzer
func NewAnalyzer(configPath string, logger *zap.SugaredLogger) *Analyzer {
	return &Analyzer{configPath, logger}
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

	switch assetType {
	case "pdf":
		return a.pdfToText(b, 10)
	default:
		return a.imageToText(b)
	}
}

func (a *Analyzer) pdfToText(content []byte, threshold int) (string, error) {
	doc, err := fitz.NewFromMemory(content)
	if err != nil {
		return "", err
	}
	defer doc.Close()

	var text string
	for i := 0; i < doc.NumPage(); i++ {
		// try pulling text
		if page, _ := doc.Text(i); len(page) > threshold {
			text += " " + page
			continue
		}

		// if text is unsatisfactory, perform OCR on image
		if image, _ := doc.Image(i); image != nil {
			var img = new(bytes.Buffer)
			jpeg.Encode(img, image, &jpeg.Options{Quality: 50})
			if img.Bytes() == nil || len(img.Bytes()) == 0 {
				continue
			}
			if page, _ := a.imageToText(img.Bytes()); page != "" {
				text += " " + page
			}
		}
	}

	return text, nil
}

func (a *Analyzer) imageToText(asset []byte) (contents string, err error) {
	t, err := a.newTesseractClient()
	if err != nil {
		return "", err
	}
	defer t.Close()

	if err = t.SetImageFromBytes(asset); err != nil {
		return "", fmt.Errorf("failed to set image: %s", err.Error())
	}

	if contents, err = t.Text(); err != nil {
		err = fmt.Errorf("failed to convert asset to text: %s", err.Error())
	}
	return
}

func (a *Analyzer) newTesseractClient() (*gosseract.Client, error) {
	t := gosseract.NewClient()
	if a.configPath != "" {
		if err := t.SetConfigFile(a.configPath); err != nil {
			return nil, fmt.Errorf("failed to set tesseract configuration file at '%s': %s",
				a.configPath, err.Error())
		}
	}
	return t, nil
}
