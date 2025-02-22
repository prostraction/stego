package stego_loader

import (
	"fmt"
	"image"
	"image/draw"
	"stego/internal/stego_encoding"
	"sync"
)

func EncodeFile(pathIn string, pathOut string, encodedWord string, pass string, encodedWordLen int, addMod int, negMod int) error {
	fmt.Println(pathOut)

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
			err = stego_encoding.Encode(imgRGBA, encodedWord, pass, encodedWordLen, addMod, negMod, c)
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
	fmt.Println(path)
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
