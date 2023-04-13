package fileStego

import (
	"fmt"
	"io/ioutil"
	"math/rand"
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
				EncodeFile(dirIn+"//"+f_list[i].Name(), dirOut+"//"+f_list[i].Name(), want, pass, len(want), 50, -50)
			}
		}()
	}
	go func() {
		for _, s := range stack {
			work <- s
		}
		close(work)
	}()
	wait.Wait()

	//if err != nil {
	//	t.Fatalf(`[ERR] Test: TestStegoRealFile: I/O err. %s`, err.Error())
	//}
}

func Decoding(t *testing.T, dirIn string, dirOut string, want string, pass string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	f_list, err := ioutil.ReadDir(dirIn)
	if err != nil {
		t.Fatalf(`[ERR] %s`, err.Error())
	}
	work := make(chan string)
	wait := sync.WaitGroup{}
	stack := make([]string, len(f_list))
	for i := 0; i < len(f_list); i++ {
		stack[i] = ""
	}
	for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for s := 0; s < len(f_list); s++ {
				stack[s] = DecodeFile(dirOut+"//"+f_list[s].Name(), pass, len(want))
			}
		}()
	}
	go func() {
		for _, s := range stack {
			work <- s
		}
		close(work)
	}()
	wait.Wait()

	for k, _ := range stack {
		str := stack[k]
		validBytes := 0
		for i := 0; i < len(str); i++ {
			if i < len(want) && str[i] == want[i] {
				validBytes++
			}
		}
		fmt.Println("[", 100*float32(validBytes)/float32(len(want)), "% recoved]", str)
	}

	//if err != nil {
	//	t.Fatalf(`[ERR] Test: TestStegoRealFile: I/O err. %s`, err.Error())
	//}
}

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

	Encoding(t, "test_images", "test_images_out", want, pass)
	Decoding(t, "test_images", "test_images_out", want, pass)
}
