package embedding

import (
	"image"
	_ "image"
	"stego/src/dct"
)

func abs[T int | byte](value T) T {
	if value < 0 {
		return -value
	}
	return value
}
func stringToBoolArray(value string) []bool {
	if len(value) == 0 {
		return nil
	}
	array := make([]bool, 0)
	for i := 0; i < len(value); i++ {
		symbol := make([]bool, 0, 8)
		ch := value[i]
		for ch != 0 {
			t := ch % 2
			if t == 0 {
				symbol = append(symbol, false)
			} else {
				symbol = append(symbol, true)
			}
			ch /= 2
		}
		//fmt.Println(len(symbol))
		for j := 0; j < 8-len(symbol); j++ {
			//fmt.Println(j)
			array = append(array, false)
		}
		for j := len(symbol) - 1; j >= 0; j-- {
			array = append(array, symbol[j])
		}
		symbol = nil
	}
	return array
}

func Encode(img *image.RGBA, encodedWord string, pass string, addMod int, negMod int, channelSelected int) bool {
	bounds := (*img).Bounds()
	if bounds.Max.X < 8 || bounds.Max.Y < 8 {
		return false
	}
	if len(encodedWord) == 0 {
		return false
	}
	if len(pass) == 0 {
		return false
	}
	const countValidIndexes = 25
	const sizeOfBlock2D = 64
	const sizeOfBlock1D = 8
	validIndexes := [countValidIndexes]int{7, 14, 15, 21, 22, 23, 28, 29, 30, 31, 35, 36, 37, 38, 42, 43, 44, 45, 48, 49, 50, 51, 55, 56, 57}

	currentSymbol := 0
	dctMatrix := make([]float32, sizeOfBlock2D)
	idctMatrix := make([]float32, sizeOfBlock2D)

	for x := 0; x < bounds.Max.X-bounds.Max.X%sizeOfBlock1D-1; x += sizeOfBlock1D {
		for y := 0; y < bounds.Max.Y-bounds.Max.Y%sizeOfBlock1D-1; y += sizeOfBlock1D {
			pixelIndex1 := validIndexes[pass[y%len(pass)]%countValidIndexes]
			pixelIndex2 := validIndexes[pass[x%len(pass)]%countValidIndexes]
			pixelIndex3 := validIndexes[abs(pass[y%len(pass)]%countValidIndexes-pass[x%len(pass)]%countValidIndexes)%countValidIndexes]

			for pixelIndex1 == pixelIndex2 {
				pixelIndex2++
				pixelIndex2 = validIndexes[pixelIndex2%countValidIndexes]
			}
			for (pixelIndex2 == pixelIndex3) || (pixelIndex1 == pixelIndex3) {
				pixelIndex3++
				pixelIndex3 = validIndexes[pixelIndex3%countValidIndexes]
			}

			dct.MakeDCT(&dctMatrix, &(*img).Pix, x, y, bounds.Max.X, 4, 0)

			currentSymbol++
		}

	}

	return true
}
