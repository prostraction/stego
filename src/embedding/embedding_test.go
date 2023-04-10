package embedding

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math/rand"
	"os"
	"strconv"
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
			t.Fatalf(`[ERR] Test: BoolArrayToString() = %q, want match for %q, (%d)`, msg, want, uint8(i))
		}
	}
	for i := -1; i < 4294967295; i += 1000 {
		testWord = string(rune(i))
		want = string(rune(i))
		boolArr := StringToBoolArray(testWord)
		msg := BoolArrayToString(boolArr)
		if msg != want {
			t.Fatalf(`[ERR] Test: BoolArrayToString() = %q, want match for %q, (%d)`, msg, want, uint8(i))
		}
	}
}

func TestStegoNoRobust(t *testing.T) {
	fmt.Println("Test: StegoNoRobust is running...")
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
				img.Set(x, y, color.RGBA64{uint16(10000 + rand.Intn(40000)), uint16(10000 + rand.Intn(40000)), uint16(10000 + rand.Intn(40000)), 255})
			}
		}
		Encode(&img, want, pass, len(pass)*32, 150, -150, 0)
		msg := Decode(&img, pass, len(pass)*32, 0)
		if msg != want {
			t.Fatalf(`[ERR] Test: StegoNoRobust() = %q, want match for %q`, msg, want)
		}
	}
}

func TestStegoRobust(t *testing.T) {
	fmt.Println("Test: StegoRobust is running...")
	pass := ""
	for count := 0; count < 5; count++ {
		for i := 'a'; i < 'z'; i++ {
			pass += string(uint8(i))
		}
	}

	var validatePercentAvg float32 = 0
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
		img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{X, Y}})
		for x := 0; x < X; x++ {
			for y := 0; y < Y; y++ {
				img.Set(x, y, color.RGBA64{uint16(10000 + rand.Intn(40000)), uint16(10000 + rand.Intn(40000)), uint16(10000 + rand.Intn(40000)), 255})
			}
		}
		Encode(img, want, pass, len(pass)*32, 50, -50, 0)
		Encode(img, want, pass, len(pass)*32, 50, -50, 1)
		Encode(img, want, pass, len(pass)*32, 50, -50, 2)

		os.Mkdir("test_images", os.ModePerm)
		f, err := os.Create("test_images/test" + strconv.Itoa(i) + ".jpg")
		if err != nil {
			fmt.Println(err)
			return
		}

		jpeg.Encode(f, img, nil)
		f.Close()

		f, _ = os.Open("test_images/test" + strconv.Itoa(i) + ".jpg")
		imgWrote, _ := jpeg.Decode(f)
		b := imgWrote.Bounds()
		imgTest := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(imgTest, imgTest.Bounds(), imgWrote, b.Min, draw.Src)
		f.Close()

		msg := Decode(imgTest, pass, len(pass)*32, 0)
		var validateBytes float32 = 0
		for i := 0; i < len(want); i++ {
			if i < len(msg) {
				if msg[i] == want[i] {
					validateBytes++
				}
			}
		}
		validatePercentAvg += validateBytes / float32(len(want)) * 100.
		//if msg != want {
		//	fmt.Println("[MSG] Loss = ", 100.*validateBytes/float32(len(want)), "%")
		//}
	}
	fmt.Println("Test: StegoRobust is finished with ", validatePercentAvg/float32(100), "% average stegomessage recovered.")
}
