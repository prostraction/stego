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
func StringToBoolArray(value string) []bool {
	if len(value) == 0 {
		return nil
	}
	array := make([]bool, 0)
	runeValue := []rune(value)
	for i := 0; i < len(runeValue); i++ {
		symbol := make([]bool, 0, 8)
		ch := runeValue[i]
		for ch != 0 {
			t := ch % 2
			if t == 0 {
				symbol = append(symbol, false)
			} else {
				symbol = append(symbol, true)
			}
			ch /= 2
		}
		for j := 0; j < 8-len(symbol); j++ {
			array = append(array, false)
		}
		for j := len(symbol) - 1; j >= 0; j-- {
			array = append(array, symbol[j])
		}
		symbol = nil
	}
	return array
}

func BoolArrayToString(value []bool) string {
	if len(value) == 0 {
		return ""
	}
	retString := ""
	symbol := make([]bool, 0, 8)
	count := 1
	for i := 0; i < len(value); i++ {
		if symbol == nil {
			symbol = make([]bool, 0, 8)
		}
		symbol = append(symbol, value[i])
		if count/8 == 1 {
			ch := 0x00
			ind := 7
			c := 0
			for ind >= 0 {
				if symbol[ind] {
					ch += (1 << c)
				}
				ind--
				c++
			}
			retString += string(rune(ch))
			symbol = nil
			count = 0
		}
		count++
	}
	return retString
}

func Encode(img *image.RGBA, encodedWord string, pass string, encodedWordLen int, addMod int, negMod int, channelSelected int) bool {
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

	boolPass := StringToBoolArray(pass)
	if len(boolPass) < encodedWordLen {
		emptySlice := make([]bool, encodedWordLen-len(pass))
		boolPass = append(boolPass, emptySlice...)
	}
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

			dct.MakeDCT(&dctMatrix, &(*img).Pix, x, y, bounds.Max.X, 4, int8(channelSelected))
			if boolPass[currentSymbol%encodedWordLen] {
				dctMatrix[pixelIndex1] = float32(addMod)
				dctMatrix[pixelIndex2] = float32(addMod)
				dctMatrix[pixelIndex3] = float32(negMod)
			} else {
				dctMatrix[pixelIndex1] = float32(negMod)
				dctMatrix[pixelIndex2] = float32(negMod)
				dctMatrix[pixelIndex3] = float32(addMod)
			}
			dct.MakeIDCT(&dctMatrix, &img.Pix, x, y, bounds.Max.X, 4, int8(channelSelected))
			currentSymbol++
		}
	}
	return true
}

func Decode(img *image.RGBA, pass string, encodedWordLen int, channelSelected int) bool {
	bounds := (*img).Bounds()
	if bounds.Max.X < 8 || bounds.Max.Y < 8 {
		return false
	}

	const countValidIndexes = 25
	const sizeOfBlock2D = 64
	const sizeOfBlock1D = 8
	validIndexes := [countValidIndexes]int{7, 14, 15, 21, 22, 23, 28, 29, 30, 31, 35, 36, 37, 38, 42, 43, 44, 45, 48, 49, 50, 51, 55, 56, 57}

	currentSymbol := 0
	dctMatrix := make([]float32, sizeOfBlock2D)
	codedWordCounter := make([]int, encodedWordLen)
	codedWordBool := make([]bool, encodedWordLen)
	blocksCount := 0

	for x := 0; x < bounds.Max.X-bounds.Max.X%sizeOfBlock1D-1; x += sizeOfBlock1D {
		for y := 0; y < bounds.Max.Y-bounds.Max.Y%sizeOfBlock1D-1; y += sizeOfBlock1D {
			if currentSymbol%encodedWordLen == 0 {
				blocksCount++
			}

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

			dct.MakeDCT(&dctMatrix, &(*img).Pix, x, y, bounds.Max.X, 4, int8(channelSelected))
			// Encoded '1'
			if dctMatrix[pixelIndex3] <= dctMatrix[pixelIndex2] && dctMatrix[pixelIndex3] <= dctMatrix[pixelIndex1] {
				codedWordCounter[currentSymbol%encodedWordLen]++
			}
			currentSymbol++
		}
	}

	for i := 0; i < encodedWordLen; i++ {
		if codedWordCounter[i] >= ((blocksCount + 1) / 2) {
			codedWordBool[i] = true
		}
	}

	return true
}
