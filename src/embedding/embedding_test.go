package embedding

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"testing"
)

func TestBoolArrayToString(t *testing.T) {
	fmt.Println("Test: BoolArrayToString is running...")
	testWord := ""
	want := ""

	for i := -255; i < 513; i++ {
		testWord += string(rune(i))
		want += string(rune(i))
		boolArr := StringToBoolArray(testWord)
		msg := BoolArrayToString(boolArr)
		if msg != want {
			t.Fatalf(`boolArrayToString() = %q, want match for %q, (%d)`, msg, want, uint8(i))
		}
	}
	for i := -1; i < 4294967295; i += 1000 {
		testWord = string(rune(i))
		want = string(rune(i))
		boolArr := StringToBoolArray(testWord)
		msg := BoolArrayToString(boolArr)
		if msg != want {
			t.Fatalf(`boolArrayToString() = %q, want match for %q, (%d)`, msg, want, uint8(i))
		}
	}
}

func TestStego(t *testing.T) {
	fmt.Println("Test: Stego is running...")
	pass := ""
	for count := 0; count < 5; count++ {
		for i := 'a'; i < 'z'; i++ {
			pass += string(uint8(i))
		}
	}

	for i := 0; i < 100; i++ {
		want := ""
		if i%2 == 0 {
			for j := 0; j < 100; j++ {
				want += string(rune((rand.Intn(100) % (255 - 32)) + 32))
			}
		} else {
			for j := 0; j < 100; j++ {
				want += string(rune((rand.Intn(65535) % (65535 - 32)) + 32))
			}
		}

		X := 8*len(pass) + rand.Intn(255)
		Y := 8*len(pass) + rand.Intn(255)
		img := *image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{X, Y}})
		for x := 0; x < X; x++ {
			for y := 0; y < Y; y++ {
				img.Set(x, y, color.RGBA64{uint16(rand.Intn(255)), uint16(rand.Intn(255)), uint16(rand.Intn(255)), 255})
			}
		}
		Encode(&img, want, pass, len(pass)*32, 150, -150, 0)
		msg := Decode(&img, pass, len(pass)*32, 0)
		if msg != want {
			t.Fatalf(`TestStego() = %q, want match for %q`, msg, want)
		}

	}
}
