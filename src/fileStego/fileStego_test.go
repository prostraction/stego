package fileStego

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
)

func TestStegoRealFile(t *testing.T) {
	fmt.Println("Test: TestStegoRealFile is running...")
	pass := ""
	want := ""
	for j := 0; j < 100; j++ {
		want += string(rune((rand.Intn(100) % (65535 - 32)) + 32))
	}
	fmt.Println("WANT: ", want)
	for count := 0; count < 5; count++ {
		for i := 'a'; i < 'z'; i++ {
			pass += string(uint8(i))
		}
	}

	f_list, err := ioutil.ReadDir("test_images")

	if err != nil {
		t.Fatalf(`[ERR] Test: TestStegoRealFile: I/O err. %s`, err.Error())
	}
	for _, v := range f_list {
		EncodeFile("test_images//"+v.Name(), "test_images_out//"+v.Name()+"_stego"+".jpg", want, pass, len(want)*32, 50, -50)
	}
	f_list2, err := ioutil.ReadDir("test_images_out")
	for _, v := range f_list2 {
		str := DecodeFile(v.Name(), pass, len(want)*32)
		fmt.Println(str)

	}
}
