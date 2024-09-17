package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"stego/internal/fileStego"
	"strconv"
	"sync"
)

// Password is just any letters combination of any size and MUST BE the same for encoding and decoding of one image
var pass = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Message which should be encoded
var msg = "Test message."

// Length of message. MUST BE known to decoder and it's equal to 1 msg's symbol = 32 bits.
// Multiply len(msg) by 32 or use large number to cover all message bits.
// Default value: 32*len(msg) will be printed, if -l argument will not be used.
// If msgLen for decoding < msgLen for encoding, then message will be cut.
// 8x8 pixel block contains 1 bit of message. At least 32 * 8x8 (image with size 128x128) pixel blocks required to encode one symbol.
var msgLen = 0 // bits

// The main parameter of encoding. More robust cause more visible hidden message, but it is more stable for compression
// Value 20 is fine for most cases. 50 is visible, but image is not corrupted
var robust = 20

// You can use "encode", "decode" or "bench" (encode and decode together)
const (
	encodeOperation = iota
	decodeOperation
	benchOperation
)

var operation = decodeOperation

// Input file/dir
var pathIn = ""

// Output file/dir
var pathOut = ""

func printHelp() {

}

func fillArgs(args []string) (err error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m":
			if i+1 < len(args) {
				msg = args[i+1]
				i++
			} else {
				return fmt.Errorf("%s requeires an argument", args[i])
			}
		case "-p":
			if i+1 < len(args) {
				pathIn = args[i+1]
				i++
			} else {
				return fmt.Errorf("%s requeires an argument", args[i])
			}
		case "-o":
			if i+1 < len(args) {
				pathOut = args[i+1]
				i++
			} else {
				return fmt.Errorf("%s requeires an argument", args[i])
			}
		case "-r":
			if i+1 < len(args) {
				r, err := strconv.Atoi(args[i+1])
				if err != nil {
					return err
				} else {
					robust = r
				}
				i++
			} else {
				return fmt.Errorf("%s requeires an argument", args[i])
			}
		case "-l":
			if i+1 < len(args) {
				l, err := strconv.Atoi(args[i+1])
				if err != nil {
					return err
				} else {
					if l < 1 {
						return fmt.Errorf("%s can`t be <= 0", args[i])
					}
					msgLen = l
				}
				i++
			} else {
				return fmt.Errorf("%s requeires an argument", args[i])
			}
		case "-d":
			operation = decodeOperation
		case "-e":
			operation = encodeOperation
		case "-b":
			operation = benchOperation
		default:
			if i > 0 {
				return fmt.Errorf("%s unknown argument", args[i])
			}
		}
	}
	return nil
}

func concRun(procOperation int, dirIn string, dirOut string) ([]string, []error, error) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fList, err := ioutil.ReadDir(dirIn)
	if err != nil {
		return nil, nil, err
	}
	work := make(chan int)
	wait := sync.WaitGroup{}
	stackValue := make([]string, len(fList))
	stackNames := make([]int, len(fList))
	stackError := make([]error, len(fList))
	for i := 0; i < len(fList); i++ {
		stackNames[i] = i
	}
	for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for i := range work {
				switch procOperation {
				case encodeOperation:
					stackError[i] = fileStego.EncodeFile(dirIn+"//"+fList[i].Name(), dirOut+"//"+fList[i].Name(), msg, pass, msgLen, robust, -robust)
				case decodeOperation:
					stackValue[i], stackError[i] = fileStego.DecodeFile(dirIn+"//"+fList[i].Name(), pass, msgLen)
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
	return stackValue, stackError, nil
}

func main() {
	args := os.Args
	err := fillArgs(args)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if msgLen == 0 {
		switch operation {
		case encodeOperation, benchOperation:
			msgLen = len(msg) * 32
			fmt.Printf("Length of a message: %d. Use it for decoding.\n", msgLen)
		default:
			fmt.Println("Specify length of a message!")
			return
		}
	}
	fi, err := os.Stat(pathIn)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if pathOut == "" {
		if fi.IsDir() {
			pathOut = fi.Name() + "_stego"
			os.Mkdir(pathOut, os.ModePerm)
			_, verifyPermErr := os.Stat(pathOut)
			if verifyPermErr != nil {
				fmt.Println("No output directory was specified. Unable to create a new directory", pathOut, "aborting.")
				fmt.Println(verifyPermErr.Error())
				return
			}
		} else {
			name, ext := func(str string) (string, string) {
				for i := len(str) - 1; i >= 0; i-- {
					if str[i] == '.' {
						return pathIn[0:i], pathIn[i:]
					}
				}
				return "", ""
			}(pathIn)
			if name == "" {
				fmt.Printf("No file type specified for %s. Aborting.\n", pathIn)
				return
			}
			pathOut = name + "_stego" + ext
		}
	} else {
		if fi.IsDir() {
			os.Mkdir(pathOut, os.ModePerm)
		} else {
			os.Create(pathOut)
		}
		_, verifyPermErr := os.Stat(pathOut)
		if verifyPermErr != nil {
			fmt.Println("No output path was specified. Unable to create a new file/directory", pathOut, "aborting.")
			fmt.Println(verifyPermErr.Error())
			return
		}
	}

	switch operation {
	case benchOperation:
		fallthrough
	case encodeOperation:
		if !fi.IsDir() {
			if err := fileStego.EncodeFile(pathIn, pathOut, msg, pass, msgLen, robust, -robust); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), pathIn)
			} else {
				fmt.Println("Message encoded.")
			}
		} else {
			if msgs, errs, err := concRun(encodeOperation, pathIn, pathOut); err != nil {
				fmt.Println(err)
			} else {
				errCount := 0
				for i := 0; i < len(msgs); i++ {
					if i < len(errs) && errs[i] != nil {
						fmt.Printf("Error: %s\n", errs[i].Error())
						errCount++
					}
				}
				if errCount == 0 {
					fmt.Println("All messages encoded.")
				}
			}
		}
		if operation != benchOperation {
			break
		}
		fallthrough
	case decodeOperation:
		if operation == benchOperation {
			pathIn = pathOut
		}
		if !fi.IsDir() {
			if msgDecoded, err := fileStego.DecodeFile(pathIn, pass, msgLen); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), pathIn)
			} else {
				fmt.Printf("%s\n", msgDecoded)
			}
		} else {
			if msgsDecoded, errs, err := concRun(decodeOperation, pathIn, pathOut); err != nil {
				fmt.Println(err)
			} else {
				fList, err := ioutil.ReadDir(pathIn)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				for i := 0; i < len(msgsDecoded); i++ {
					if i < len(errs) && errs[i] != nil {
						fmt.Printf("Error: %s for %s\n", errs[i].Error(), pathIn)
					} else {
						fmt.Printf("[%s]\t\"%s\"\n", fList[i].Name(), msgsDecoded[i])
					}
				}
			}
		}
	}
}
