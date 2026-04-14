package stego_loader

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/edwvee/exiffix"
	"golang.org/x/image/bmp"
)

// openImage reads an image from the specified path.
// Supports JPEG, PNG, BMP, and GIF.
func openImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpeg", ".jpg":
		img, _, err = exiffix.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	case ".bmp":
		img, err = bmp.Decode(file)
	case ".gif":
		img, err = gif.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode %s image: %w", ext, err)
	}

	return img, nil
}

// writeImage saves an RGBA image to the specified path based on its extension.
func writeImage(path string, img *image.RGBA) error {
	fOut, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer fOut.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpeg", ".jpg":
		return jpeg.Encode(fOut, img, nil)
	case ".png":
		return png.Encode(fOut, img)
	case ".bmp":
		return bmp.Encode(fOut, img)
	case ".gif":
		return gif.Encode(fOut, img, nil)
	default:
		return fmt.Errorf("unsupported output format: %s", ext)
	}
}
