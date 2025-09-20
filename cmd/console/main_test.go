package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

/*

	var opts struct {
	Pass    string `short:"p" long:"pass" default:"abcdefghijklmnopqrstuvwxyz" required:"true" description:"Password is just any letters combination of any size and used for encoding/decoding for this file. It MUST BE the same for encoding and decoding of one image."`
	Msg     string `short:"m" long:"message" required:"false" description:"Message which should be encoded (or decoding verification, not nessesary)."`
	MsgLen  int    `short:"l" long:"len" required:"true" description:"Length of message. MUST BE known to decoder and it's equal to 1 Msg's symbol = 32 bits."`
	Robust  int    `short:"r" long:"robust" default:"20" required:"true" description:"The main parameter of encoding. More Robust cause more visible hidden message, but it is more stable for compression. Value 20 is fine for most cases. 50 is visible, but image is not corrupted."`
	Action  string `short:"a" long:"action" required:"true" description:"Available values: d, e, b (decode / encode / benchmark)"`
	PathIn  string `short:"i" long:"input" required:"true" description:"Path to input files/dir"`
	PathOut string `short:"o" long:"output" required:"false" description:"Path to output files dir"`
}


*/

func TestCreateTestFiles(t *testing.T) {
	fmt.Println("Test: CreateTestFiles is running...")

	for i := 0; i < 100; i++ {
		X := 87 + rand.Intn(711)
		Y := 23 + rand.Intn(933)
		img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{X, Y}})
		for x := 0; x < X; x++ {
			for y := 0; y < Y; y++ {
				img.Set(x, y, color.RGBA64{uint16(10000 + rand.Intn(40000)), uint16(10000 + rand.Intn(40000)), uint16(10000 + rand.Intn(40000)), 255})
			}
		}

		os.Mkdir("test_images", os.ModePerm)
		f, err := os.Create("test_images/test" + strconv.Itoa(i) + ".jpg")
		if err != nil {
			fmt.Println(err)
			return
		}

		jpeg.Encode(f, img, nil)
		f.Close()
	}
	fmt.Println("Test: CreateTestFiles done.")
}

func TestApp(t *testing.T) {
	opts.Pass = "sdflghdfjklghdfkjghldksfgsdfgds"
	opts.Msg = "Testing for app working!"
	opts.MsgLen = len(opts.Msg) * 32
	opts.Robust = 30
	Action = benchAction
	opts.PathIn = "test_images"
	opts.PathOut = "test_images_stego"

	fileInfo, err := CreateDir()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	switch Action {
	case benchAction:
		fallthrough
	case encodeAction:
		runFileOperation(fileInfo, "encode")
		if Action != benchAction {
			break
		}
		fallthrough
	case decodeAction:
		runFileOperation(fileInfo, "decode")
	}
}
