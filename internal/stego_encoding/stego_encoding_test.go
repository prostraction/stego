package stego_encoding

import (
	"image"
	"image/color"
	"testing"
)

func TestBitConversion(t *testing.T) {
	tests := []struct {
		name string
		val  string
	}{
		{"empty", ""},
		{"simple", "hello"},
		{"unicode", "привет"},
		{"long", "this is a much longer message to test the bit conversion"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bits := StringToBoolArray(tt.val)
			back := BoolArrayToString(bits)
			if back != tt.val {
				t.Errorf("Bit conversion failed: got %q, want %q", back, tt.val)
			}
		})
	}
}

func TestEncodeDecode(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	// Fill with some texture
	for x := 0; x < 200; x++ {
		for y := 0; y < 200; y++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), uint8((x + y) / 2), 255})
		}
	}

	pass := "secret"
	msg := "stego test"
	msgLen := len(msg) * 32

	// Encode
	if err := Encode(img, msg, pass, msgLen, 50, -50, 0); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	decoded, err := Decode(img, pass, msgLen, 0)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded != msg {
		t.Errorf("Decoded message mismatch: got %q, want %q", decoded, msg)
	}
}

func TestEncodeErrors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4)) // Too small
	err := Encode(img, "test", "pass", 128, 50, -50, 0)
	if err == nil {
		t.Error("Should have failed for small image")
	}

	imgLarge := image.NewRGBA(image.Rect(0, 0, 100, 100))
	err = Encode(imgLarge, "", "pass", 128, 50, -50, 0)
	if err == nil {
		t.Error("Should have failed for empty message")
	}

	err = Encode(imgLarge, "test", "", 128, 50, -50, 0)
	if err == nil {
		t.Error("Should have failed for empty password")
	}
}
