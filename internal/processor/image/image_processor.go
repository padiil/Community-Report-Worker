package image

import (
	"bytes"
	"image"
	"io"

	_ "image/jpeg"
	_ "image/png"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

func ProcessImage(input io.Reader) (*bytes.Buffer, error) {
	img, _, err := image.Decode(input)
	if err != nil {
		return nil, err
	}

	resized := imaging.Resize(img, 800, 0, imaging.Lanczos)

	buf := new(bytes.Buffer)
	if err := webp.Encode(buf, resized, &webp.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return buf, nil
}
