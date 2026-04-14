// Package stego_loader provides high-level functions for encoding and decoding
// messages in image files of various formats.
package stego_loader

import (
	"image"
	"image/draw"
	"stego/internal/stego_encoding"
	"sync"
)

// EncodeFile opens an image file at pathIn, embeds a message into it,
// and saves the result to pathOut.
func EncodeFile(pathIn string, pathOut string, encodedWord string, pass string, encodedWordLen int, addMod int, negMod int) error {
	img, err := openImage(pathIn)
	if err != nil {
		return err
	}
	b := img.Bounds()
	imgRGBA := image.NewRGBA(image.Rect(0, 0, b.Max.X, b.Max.Y))
	draw.Draw(imgRGBA, imgRGBA.Bounds(), img, b.Min, draw.Src)

	wg := sync.WaitGroup{}
	var encodeErr error
	var once sync.Once

	// Process 3 channels (R, G, B) in parallel
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			err := stego_encoding.Encode(imgRGBA, encodedWord, pass, encodedWordLen, addMod, negMod, c)
			if err != nil {
				once.Do(func() {
					encodeErr = err
				})
			}
		}(i)
	}
	wg.Wait()

	if encodeErr != nil {
		return encodeErr
	}

	return writeImage(pathOut, imgRGBA)
}

// DecodeFile opens an image file at path and extracts a hidden message from it.
func DecodeFile(path string, pass string, encodedWordLen int) (string, error) {
	img, err := openImage(path)
	if err != nil {
		return "", err
	}
	b := img.Bounds()
	imgRGBA := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(imgRGBA, imgRGBA.Bounds(), img, b.Min, draw.Src)
	str, err := stego_encoding.Decode(imgRGBA, pass, encodedWordLen, 0)
	return str, err
}
