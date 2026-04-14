// Package dct provides implementations of the Discrete Cosine Transform (DCT)
// and Inverse Discrete Cosine Transform (IDCT) for 8x8 pixel blocks.
// This is used as a core component for frequency-domain steganography.
package dct

import (
	"math"
)

// pixel returns the value of a specific channel of a pixel at (x, y) in the image buffer.
func pixel(img *[]uint8, x int, y int, width int, channels int, channelSelected int) uint8 {
	return (*img)[channelSelected+(channels*(x+(y*width)))]
}

// MakeDCT performs a naive 2D Discrete Cosine Transform on an 8x8 block of an image.
// It processes a single channel of the image starting at (xPos, yPos).
// The resulting 64 DCT coefficients are stored in the provided dctMatrix (must be size 64).
func MakeDCT(dctMatrix *[]float32, img *[]uint8, xPos int, yPos int, width int, channels int, channelSelected int) {
	for u := 0; u < 8; u++ {
		for v := 0; v < 8; v++ {
			var sum float64
			for x := 0; x < 8; x++ {
				for y := 0; y < 8; y++ {
					p := float64(pixel(img, xPos+x, yPos+y, width, channels, channelSelected))
					sum += p * math.Cos((2*float64(x)+1)*float64(u)*math.Pi/16.0) *
						math.Cos((2*float64(y)+1)*float64(v)*math.Pi/16.0)
				}
			}
			cu := 1.0
			if u == 0 {
				cu = 1.0 / math.Sqrt(2)
			}
			cv := 1.0
			if v == 0 {
				cv = 1.0 / math.Sqrt(2)
			}
			(*dctMatrix)[v*8+u] = float32(0.25 * cu * cv * sum)
		}
	}
}

// MakeIDCT performs a naive 2D Inverse Discrete Cosine Transform on an 8x8 block of coefficients.
// It transforms the coefficients in dctMatrix back into pixel values and writes them
// to a single channel of the provided image buffer starting at (xPos, yPos).
func MakeIDCT(dctMatrix *[]float32, img *[]uint8, xPos int, yPos int, width int, channels int, channelSelected int) {
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			var sum float64
			for u := 0; u < 8; u++ {
				for v := 0; v < 8; v++ {
					cu := 1.0
					if u == 0 {
						cu = 1.0 / math.Sqrt(2)
					}
					cv := 1.0
					if v == 0 {
						cv = 1.0 / math.Sqrt(2)
					}
					sum += cu * cv * float64((*dctMatrix)[v*8+u]) *
						math.Cos((2*float64(x)+1)*float64(u)*math.Pi/16.0) *
						math.Cos((2*float64(y)+1)*float64(v)*math.Pi/16.0)
				}
			}
			val := 0.25 * sum
			if val < 0 {
				val = 0
			}
			if val > 255 {
				val = 255
			}
			(*img)[channelSelected+(channels*((xPos+x)+width*(yPos+y)))] = uint8(math.Round(val))
		}
	}
}
