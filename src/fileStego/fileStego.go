package fileStego

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"stego/src/embedding"
	"strings"
	"sync"

	"github.com/edwvee/exiffix"
)

// Opens file and return image from it
func openImage(path string) (image.Image, error) {
	var img image.Image
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if filepath.Ext(strings.ToLower(file.Name())) == ".jpeg" || filepath.Ext(strings.ToLower(file.Name())) == ".jpg" {
		img, _, err = exiffix.Decode(file)
		if err != nil {
			return nil, err
		}
	} else if filepath.Ext(strings.ToLower(file.Name())) == ".png" {
		img, err = png.Decode(file)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("wrong type of file")
	}
	return img, nil
}

func clamp(value, min, max uint32) uint32 {
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}
}

func EncodeFile(pathFrom string, pathTo string, encodedWord string, pass string, encodedWordLen int, addMod int, negMod int) bool {
	img, err := openImage(pathFrom)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	b := img.Bounds()
	//fmt.Println(pathFrom, b.Max.X, b.Max.Y)
	imgRGBA := image.NewRGBA(image.Rect(0, 0, b.Max.X, b.Max.Y))
	draw.Draw(imgRGBA, imgRGBA.Bounds(), img, b.Min, draw.Src)
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			imgRGBA.Pix[x+y*b.Max.X] = uint8(clamp(uint32(imgRGBA.Pix[x+y*b.Max.X]), 1, 254))
			//r, g, b, a := imgRGBA.At(x, y).RGBA()
			//r = clamp(r, 0, 65535)
			//g = clamp(g, 0, 65535)
			//b = clamp(b, 0, 65535)
			//a = clamp(a, 1, 254)
			//imgRGBA.Set(x, y, color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(a)})
		}
	}

	work := make(chan int)
	wg := sync.WaitGroup{}
	stack := make([]int, 3)
	for i := 0; i < 3; i++ {
		stack[i] = i
	}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			embedding.Encode(imgRGBA, encodedWord, pass, len(pass)*32, addMod, negMod, c)
		}(i)
	}
	go func() {
		for _, s := range stack {
			work <- s
		}
		close(work)
	}()
	wg.Wait()
	f_out, err := os.Create(pathTo)
	if err != nil {
		fmt.Println(err)
		return false
	}
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			imgRGBA.Pix[x+y*b.Max.X] = uint8(clamp(uint32(imgRGBA.Pix[x+y*b.Max.X]), 1, 254))
			//r, g, b, a := imgRGBA.At(x, y).RGBA()
			//r = clamp(r, 0, 65535)
			//g = clamp(g, 0, 65535)
			//b = clamp(b, 0, 65535)
			//a = clamp(a, 1, 254)
			//imgRGBA.Set(x, y, color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(a)})
		}
	}
	jpeg.Encode(f_out, imgRGBA, nil)
	return true
}

func DecodeFile(path string, pass string, encodedWordLen int) string {
	img, _ := openImage(path)
	b := img.Bounds()
	imgRGBA := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(imgRGBA, imgRGBA.Bounds(), img, b.Min, draw.Src)

	str := embedding.Decode(imgRGBA, pass, len(pass)*32, 0)
	return str
}
