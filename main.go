package main

import (
	"fmt"
	"stego/src/dct"
)

func main() {
	img := make([]uint8, 64)
	dctMatrix := make([]float32, 64)

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			img[8*i+j] = uint8(8*j + i)
		}
	}

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			fmt.Print(img[8*j+i], " ")
		}
		fmt.Print("\n")
	}

	dct.MakeDCT(&dctMatrix, &img, 0, 0, 8, 1, 0)
	dct.MakeIDCT(&dctMatrix, &img, 0, 0, 8, 1, 0)

	fmt.Println("------------")
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			fmt.Print(img[8*j+i], " ")
		}
		fmt.Print("\n")
	}

}
