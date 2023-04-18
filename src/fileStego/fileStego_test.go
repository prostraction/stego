package fileStego

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"testing"
)

func Encoding(t *testing.T, dirIn string, dirOut string, want string, pass string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	f_list, err := ioutil.ReadDir(dirIn)
	if err != nil {
		t.Fatalf(`[ERR] %s`, err.Error())
	}
	os.Mkdir(dirOut, os.ModePerm)
	_, err = os.Stat(dirOut)
	if err != nil {
		t.Fatalf(err.Error())
	}

	work := make(chan int)
	wait := sync.WaitGroup{}
	stack := make([]int, len(f_list))
	for i := 0; i < len(f_list); i++ {
		stack[i] = i
	}
	for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for i := range work {
				errEnc := EncodeFile(dirIn+"//"+f_list[i].Name(), dirOut+"//"+f_list[i].Name(), want, pass, len(want)*32, 50, -50)
				if errEnc != nil {
					err = errEnc
				}
			}
		}()
	}
	go func() {
		for _, s := range stack {
			work <- s
		}
		close(work)
	}()
	if err != nil {
		t.Fatal("[ERR] FileStegoTest: Encoding: " + err.Error())
	}
	wait.Wait()
}

func Decoding(t *testing.T, dirIn string, dirOut string, want string, pass string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	f_list, err := ioutil.ReadDir(dirIn)
	if err != nil {
		t.Fatalf(`[ERR] %s`, err.Error())
	}
	work := make(chan int)
	wait := sync.WaitGroup{}
	stackMsgs := make([]string, len(f_list))
	stackNames := make([]int, len(f_list))
	for i := 0; i < len(f_list); i++ {
		stackNames[i] = i
	}
	for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for s := range work {
				stackMsgs[s], err = DecodeFile(dirOut+"//"+f_list[s].Name(), pass, len(want)*32)
				if err != nil {
					fmt.Printf(`[ERR] %s\n`, err.Error())
				}
			}
		}()
	}
	go func() {
		for _, s := range stackNames {
			work <- s
		}
		close(work)
	}()
	wait.Wait()

	var avgPercentRecovered float32 = 0.
	for k, _ := range stackMsgs {
		str := stackMsgs[k]
		validBytes := 0
		for i := 0; i < len(str); i++ {
			if i < len(want) && str[i] == want[i] {
				validBytes++
			}
		}
		percent := 100 * float32(validBytes) / float32(len(want))
		if percent < 80 {
			fmt.Println("[WARN] TestStegoRealFile: Decoding: only", percent, "% recovered [", str, "]")
		}
		avgPercentRecovered += percent

	}
	avgPercentRecovered /= float32(len(stackMsgs))
	fmt.Println("[INFO] TestStegoRealFile: Decoding: ", avgPercentRecovered, "% recovered.")
}

func TestStegoRealFile(t *testing.T) {
	fmt.Println("Test: TestStegoRealFile is running...")
	pass := ""
	want := ""
	for j := 0; j < 100; j++ {
		want += string(rune((rand.Intn(100) % (65535 - 32)) + 32))
	}
	for count := 0; count < 5; count++ {
		for i := 'a'; i < 'z'; i++ {
			pass += string(uint8(i))
		}
	}

	Encoding(t, "test_images", "test_images_out", want, pass)
	Decoding(t, "test_images", "test_images_out", want, pass)
}
