// Package stego_encoding provides the core logic for embedding and extracting
// messages in images using the Discrete Cosine Transform (DCT).
package stego_encoding

import (
	"errors"
	"fmt"
	"image"
	"stego/pkg/dct"
)

const (
	// countValidIndexes is the number of DCT indices used for embedding data.
	countValidIndexes = 25
	// sizeOfBlock2D is the total number of coefficients in an 8x8 DCT block.
	sizeOfBlock2D = 64
	// sizeOfBlock1D is the dimension of one side of a DCT block.
	sizeOfBlock1D = 8
)

// validIndexes are the specific DCT indices used to store message bits.
// These are typically middle-frequency coefficients to balance robustness and visibility.
var validIndexes = [countValidIndexes]int{7, 14, 15, 21, 22, 23, 28, 29, 30, 31, 35, 36, 37, 38, 42, 43, 44, 45, 48, 49, 50, 51, 55, 56, 57}

// abs returns the absolute value of a generic integer or byte.
func abs[T int | byte](value T) T {
	if value < 0 {
		return -value
	}
	return value
}

// StringToBoolArray converts a string into a slice of booleans (bits).
// Each character is represented by 32 bits.
func StringToBoolArray(value string) []bool {
	if len(value) == 0 {
		return nil
	}
	array := make([]bool, 0)
	runeValue := []rune(value)
	for i := 0; i < len(runeValue); i++ {
		symbol := make([]bool, 0, 32)
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
		for j := 0; j < 32-len(symbol); j++ {
			symbol = append(symbol, false)
		}
		for j := len(symbol) - 1; len(symbol) != 32; j-- {
			symbol = append(symbol, symbol[j])
		}
		array = append(array, symbol...)
	}
	return array
}

// BoolArrayToString converts a slice of booleans (bits) back into a string.
// Every 32 bits are converted into a single rune.
func BoolArrayToString(value []bool) string {
	if len(value) < 32 {
		return ""
	}
	retString := ""
	symbol := make([]bool, 0, 32)
	count := 1
	for i := 0; i < len(value); i++ {
		symbol = append(symbol, value[i])

		if count/32 == 1 {
			var r rune
			c := 0
			for j := 0; j < 32; j++ {
				if symbol[j] {
					r += (1 << c)
				}
				c++
			}
			retString += string(r)
			symbol = make([]bool, 0, 32)
			count = 0
		}
		count++
	}
	return retString
}

// Encode hides a message (encodedWord) into an image using DCT.
// pass: password used to jitter the choice of DCT indices for better security.
// encodedWordLen: expected total length of the message in bits.
// addMod, negMod: robustness parameters applied to DCT coefficients.
// channelSelected: the image channel (R=0, G=1, B=2) to embed the message in.
func Encode(img *image.RGBA, encodedWord string, pass string, encodedWordLen int, addMod int, negMod int, channelSelected int) error {
	bounds := img.Bounds()
	if bounds.Max.X < 8 || bounds.Max.Y < 8 {
		return errors.New("image resolution lower than 8x8")
	}
	if len(encodedWord) == 0 {
		return errors.New("no stegomessage provided to encode")
	}
	if encodedWordLen <= 0 {
		return fmt.Errorf("invalid encodedWordLen: %d", encodedWordLen)
	}
	if len(pass) == 0 {
		return errors.New("no password provided to encode")
	}

	currentSymbol := 0
	dctMatrix := make([]float32, sizeOfBlock2D)
	boolEncoded := StringToBoolArray(encodedWord)

	if len(boolEncoded) < encodedWordLen {
		emptySlice := make([]bool, encodedWordLen-len(boolEncoded))
		boolEncoded = append(boolEncoded, emptySlice...)
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
			dct.MakeDCT(&dctMatrix, &img.Pix, x, y, bounds.Max.X, 4, channelSelected)
			if boolEncoded[currentSymbol%encodedWordLen] {
				dctMatrix[pixelIndex1] = float32(addMod)
				dctMatrix[pixelIndex2] = float32(addMod)
				dctMatrix[pixelIndex3] = float32(negMod)
			} else {
				dctMatrix[pixelIndex1] = float32(negMod)
				dctMatrix[pixelIndex2] = float32(negMod)
				dctMatrix[pixelIndex3] = float32(addMod)
			}
			dct.MakeIDCT(&dctMatrix, &img.Pix, x, y, bounds.Max.X, 4, channelSelected)
			currentSymbol++
		}
	}
	return nil
}

// Decode extracts a hidden message from an image using DCT.
// pass: password must match the one used during encoding.
// encodedWordLen: length of the message in bits to extract.
// channelSelected: the image channel to extract from.
func Decode(img *image.RGBA, pass string, encodedWordLen int, channelSelected int) (string, error) {
	bounds := img.Bounds()
	if bounds.Max.X < 8 || bounds.Max.Y < 8 {
		return "", errors.New("image resolution lower than 8x8")
	}
	if encodedWordLen <= 0 {
		return "", fmt.Errorf("invalid msg length: %d", encodedWordLen)
	}
	if len(pass) == 0 {
		return "", errors.New("no password provided to decode")
	}

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

			dct.MakeDCT(&dctMatrix, &img.Pix, x, y, bounds.Max.X, 4, channelSelected)
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
	msg := BoolArrayToString(codedWordBool)
	// Trim trailing null bytes
	for i := len(msg) - 1; i >= 0; i-- {
		if msg[i] == 0x00 {
			for i >= 0 && msg[i] == 0x00 {
				i--
			}
			msg = msg[:i+1]
			break
		}
	}
	return msg, nil
}
