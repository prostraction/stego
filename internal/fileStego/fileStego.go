package fileStego

import (
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"stego/pkg/embedding"
	"strings"
	"sync"

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
	case ".png":
		img, err = png.Decode(file)
	default:
		return nil, errors.New("wrong type of file")
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

func EncodeFile(pathIn string, pathOut string, encodedWord string, pass string, encodedWordLen int, addMod int, negMod int) error {
	img, err := openImage(pathIn)
	if err != nil {
		return err
	}
	b := img.Bounds()
	imgRGBA := image.NewRGBA(image.Rect(0, 0, b.Max.X, b.Max.Y))
	draw.Draw(imgRGBA, imgRGBA.Bounds(), img, b.Min, draw.Src)

	work := make(chan int)
	wg := sync.WaitGroup{}
	stack := make([]int, 3)
	for i := 0; i < 3; i++ {
		stack[i] = i
	}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			err = embedding.Encode(imgRGBA, encodedWord, pass, encodedWordLen, addMod, negMod, c)
		}(i)
	}
	go func() {
		for _, s := range stack {
			work <- s
		}
		close(work)
	}()
	wg.Wait()
	if err != nil {
		return err
	}
	writeImage(pathOut, imgRGBA)
	return nil
}

func DecodeFile(path string, pass string, encodedWordLen int) (string, error) {
	img, err := openImage(path)
	if err != nil {
		return "", err
	}
	b := img.Bounds()
	imgRGBA := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(imgRGBA, imgRGBA.Bounds(), img, b.Min, draw.Src)
	str, err := embedding.Decode(imgRGBA, pass, encodedWordLen, 0)
	return str, err
}
