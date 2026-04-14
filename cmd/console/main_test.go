package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestCLI(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stego-cli-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a larger synthetic image (200x200) to hold enough bits
	inputPath := filepath.Join(tmpDir, "input.png")
	outputPath := filepath.Join(tmpDir, "output.png")
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for x := 0; x < 200; x++ {
		for y := 0; y < 200; y++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 100, 255})
		}
	}
	f, err := os.Create(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(f, img)
	f.Close()

	pass := "secret"
	msg := "CLI test message"
	msgLen := len(msg) * 32

	// Test handleEncode directly
	if err := handleEncode([]string{"-i", inputPath, "-o", outputPath, "-m", msg, "-p", pass}); err != nil {
		t.Fatalf("handleEncode failed: %v", err)
	}

	// Test handleDecode directly
	if err := handleDecode([]string{"-i", outputPath, "-p", pass, "-l", fmt.Sprintf("%d", msgLen)}); err != nil {
		t.Fatalf("handleDecode failed: %v", err)
	}
}
