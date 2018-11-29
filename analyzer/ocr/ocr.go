package ocr

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"time"

	"github.com/RTradeLtd/Lens/logs"

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

// Analyze executes OCR on text
func (a *Analyzer) Analyze(jobID string, content []byte, assetType string) (contents string, err error) {
	if content == nil {
		return "", errors.New("invalid asset provided")
	}

	switch assetType {
	case "pdf":
		return a.pdfToText(jobID, content, 10)
	default:
		return a.imageToText(jobID, content)
	}
}

func (a *Analyzer) pdfToText(jobID string, content []byte, threshold int) (string, error) {
	var l = logs.NewProcessLogger(a.l, "pdf_to_text",
		"job_id", jobID,
		"threshold", threshold)
	var start = time.Now()

	doc, err := fitz.NewFromMemory(content)
	if err != nil {
		l.Warn("failed to create fitz document in memory from content",
			"error", err)
		return "", errors.New("failed to analyze PDF")
	}
	defer doc.Close()

	var text string
	var ocrPages int
	var textPages int
	for i := 0; i < doc.NumPage(); i++ {
		// try pulling text
		if page, err := doc.Text(i); err != nil {
			l.Warnw("failed to convert document page to text",
				"error", i, "error", err)
		} else if len(page) > threshold {
			textPages++
			text += " " + page
			continue
		}

		// if text is unsatisfactory, perform OCR on image
		if image, _ := doc.Image(i); image != nil {
			ocrPages++
			var img = new(bytes.Buffer)
			if err := jpeg.Encode(img, image, &jpeg.Options{Quality: 50}); err != nil {
				l.Warnw("failed to convert document page to image",
					"page", i, "error", err)
				return "", fmt.Errorf("failed to analyze page %d of document", i)
			}
			if img.Bytes() == nil || len(img.Bytes()) == 0 {
				continue
			}
			if page, err := a.imageToText(jobID, img.Bytes()); err != nil {
				l.Warnw("failed to OCR document page",
					"page", i, "error", err)
				return "", fmt.Errorf("failed to analyze page %d of document", i)
			} else if page != "" {
				text += " " + page
			}
		}
	}

	l.Infow("PDF to text conversion complete",
		"duration", time.Since(start),
		"converted.length", len(text),
		"converted.pages.text_extract", textPages,
		"converted.pages.ocr", ocrPages)

	return text, nil
}

func (a *Analyzer) imageToText(jobID string, asset []byte) (contents string, err error) {
	var l = logs.NewProcessLogger(a.l, "image_to_text",
		"job_id", jobID)
	var start = time.Now()

	t, err := a.newTesseractClient()
	if err != nil {
		l.Errorw("failed to init tesseract client", "error", err)
		return "", errors.New("failed to start analysis engine")
	}
	defer t.Close()

	if err = t.SetImageFromBytes(asset); err != nil {
		l.Warnw("failed to set image from asset", "error", err)
		return "", errors.New("failed to analyze image")
	}

	if contents, err = t.Text(); err != nil {
		l.Warnw("failed to convert image to text", "error", err)
		return "", errors.New("failed to convert image to text")
	}

	l.Infow("converted image to text",
		"duration", time.Since(start),
		"converted.length", len(contents))

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
