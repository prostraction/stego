package stego_loader

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderFormats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stego-loader-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a synthetic image
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for x := 0; x < 200; x++ {
		for y := 0; y < 200; y++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 100, 255})
		}
	}

	formats := []string{".png", ".bmp", ".jpg"} // GIF is indexed, so it might not work well with DCT modifications without careful handling

	msg := "test message"
	pass := "secret"
	msgLen := len(msg) * 32

	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			inputPath := filepath.Join(tmpDir, "input"+ext)
			outputPath := filepath.Join(tmpDir, "output"+ext)

			// Save original image to disk
			if err := writeImage(inputPath, img); err != nil {
				t.Fatalf("Failed to save test image: %v", err)
			}

			// Encode
			if err := EncodeFile(inputPath, outputPath, msg, pass, msgLen, 50, -50); err != nil {
				t.Fatalf("EncodeFile failed: %v", err)
			}

			// Decode
			decoded, err := DecodeFile(outputPath, pass, msgLen)
			if err != nil {
				t.Fatalf("DecodeFile failed: %v", err)
			}

			if decoded != msg {
				t.Errorf("Message mismatch: got %q, want %q", decoded, msg)
			}
		})
	}
}

func TestLoaderErrors(t *testing.T) {
	_, err := DecodeFile("nonexistent.png", "pass", 128)
	if err == nil {
		t.Error("Should have failed for nonexistent file")
	}

	_, err = openImage("nonexistent.txt")
	if err == nil {
		t.Error("Should have failed for unsupported extension")
	}
}
