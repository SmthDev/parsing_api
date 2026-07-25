package ocr

import (
	"fmt"
	"image"
	_ "image/jpeg" // register JPEG decoder
	"image/png"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/otiai10/gosseract/v2"
)

func preprocess(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w < 1000 {
		scale := 2000.0 / float64(w)
		src = imaging.Resize(src, int(float64(w)*scale), int(float64(h)*scale), imaging.Lanczos)
	}

	src = imaging.Grayscale(src)
	src = imaging.AdjustContrast(src, 30)
	src = imaging.Sharpen(src, 1.5)

	return src
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func saveTempPNG(img image.Image) (string, error) {
	tmp, err := os.CreateTemp("", "ocr_preprocessed_*.png")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if err := png.Encode(tmp, img); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}

func ExtractText(imagePath string) (string, error) {
	img, err := loadImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("load image: %w", err)
	}

	processed := preprocess(img)

	tmpPath, err := saveTempPNG(processed)
	if err != nil {
		return "", fmt.Errorf("save temp image: %w", err)
	}
	defer os.Remove(tmpPath)

	client := gosseract.NewClient()
	defer func() { _ = client.Close() }()

	if err := client.SetImage(tmpPath); err != nil {
		return "", err
	}
	if err := client.SetLanguage("rus+eng"); err != nil {
		return "", err
	}
	if err := client.SetPageSegMode(gosseract.PSM_SINGLE_BLOCK); err != nil {
		return "", err
	}

	text, err := client.Text()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
