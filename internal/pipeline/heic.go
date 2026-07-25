package pipeline

import (
	"bytes"
	"fmt"
	"image/jpeg"

	"github.com/jdeng/goheif"
)

func IsHEIC(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	if string(data[4:8]) != "ftyp" {
		return false
	}
	brand := string(data[8:12])
	switch brand {
	case "heic", "heix", "hevc", "hevx", "heim", "heis", "mif1":
		return true
	}
	return false
}

func ConvertHEICtoJPEG(data []byte) ([]byte, error) {
	img, err := goheif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode heic: %w", err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}
