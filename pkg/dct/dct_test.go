package dct

import (
	"math"
	"testing"
)

func TestDCTIDCT(t *testing.T) {
	// Create an 8x8 block of pixel values
	img := make([]uint8, 8*8*4) // RGBA
	for i := 0; i < 64; i++ {
		img[i*4] = uint8(i * 3) // Just some data in the red channel
	}

	dctMatrix := make([]float32, 64)
	MakeDCT(&dctMatrix, &img, 0, 0, 8, 4, 0)

	// Verify we got some non-zero coefficients
	hasNonZero := false
	for _, v := range dctMatrix {
		if v != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Error("DCT resulted in all zero coefficients")
	}

	// Inverse transform
	newImg := make([]uint8, 8*8*4)
	MakeIDCT(&dctMatrix, &newImg, 0, 0, 8, 4, 0)

	// Compare original and reconstructed
	for i := 0; i < 64; i++ {
		orig := img[i*4]
		recon := newImg[i*4]
		// Allow for small rounding errors due to float32 precision and math.Round
		if math.Abs(float64(orig)-float64(recon)) > 1.0 {
			t.Errorf("Pixel %d mismatch: got %d, want %d", i, recon, orig)
		}
	}
}

func TestDCTFlatImage(t *testing.T) {
	// A flat image (all same color) should have only the DC coefficient non-zero.
	img := make([]uint8, 8*8*4)
	for i := 0; i < 64; i++ {
		img[i*4] = 128
	}

	dctMatrix := make([]float32, 64)
	MakeDCT(&dctMatrix, &img, 0, 0, 8, 4, 0)

	// DC coefficient is at index 0
	if math.Abs(float64(dctMatrix[0])-128*8) > 1.0 {
		t.Errorf("DC coefficient mismatch: got %f, want ~1024", dctMatrix[0])
	}

	for i := 1; i < 64; i++ {
		if math.Abs(float64(dctMatrix[i])) > 0.1 {
			t.Errorf("AC coefficient %d should be ~0, got %f", i, dctMatrix[i])
		}
	}
}
