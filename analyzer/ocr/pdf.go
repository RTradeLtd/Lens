package ocr

import "gopkg.in/gographics/imagick.v2/imagick"

func pdfToImage(content []byte) ([]byte, error) {
	imagick.Initialize()
	defer imagick.Terminate()
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.SetResolution(600, 600); err != nil {
		return nil, err
	}

	if err := mw.ReadImageBlob(content); err != nil {
		return nil, err
	}

	if err := mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_FLATTEN); err != nil {
		return nil, err
	}

	if err := mw.SetCompressionQuality(75); err != nil {
		return nil, err
	}

	if err := mw.SetFormat("jpg"); err != nil {
		return nil, err
	}

	return mw.GetImageBlob(), nil
}
