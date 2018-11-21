package ocr

import (
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func pdfToImage(content []byte) ([]byte, error) {
	imagick.Initialize()
	defer imagick.Terminate()
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.SetResolution(600, 600); err != nil {
		return nil, fmt.Errorf("failed to set resolution: %s", err.Error())
	}

	if err := mw.ReadImageBlob(content); err != nil {
		return nil, fmt.Errorf("failed to read content: %s", err.Error())
	}

	if err := mw.SetCompressionQuality(75); err != nil {
		return nil, fmt.Errorf("failed to set compression quality: %s", err.Error())
	}

	if err := mw.SetFormat("png"); err != nil {
		return nil, fmt.Errorf("failed to convert to image: %s", err.Error())
	}

	return mw.GetImageBlob(), nil
}
