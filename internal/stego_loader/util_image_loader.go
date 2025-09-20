package stego_loader

import (
	"path/filepath"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
	"os"
	"errors"
	"github.com/edwvee/exiffix"
)

// Opens file and return image from it
func openImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	switch filepath.Ext(strings.ToLower(path)) {
	case ".jpeg":
		fallthrough
	case ".jpg":
		img, _, err = exiffix.Decode(file)
	case ".JPG":
		img, _, err = exiffix.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		return nil, errors.New("wrong type of file: " + path)
	}
	return
}

func writeImage(path string, img *image.RGBA) error {
	f_out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f_out.Close()
	switch filepath.Ext(strings.ToLower(path)) {
	case ".jpeg":
		fallthrough
	case ".jpg":
		return jpeg.Encode(f_out, img, nil)
	case ".png":
		return png.Encode(f_out, img)
	default:
		return errors.New("unknown type of file" + path)
	}
}