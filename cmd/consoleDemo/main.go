package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"stego/internal/fileStego"
	"sync"

	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Pass    string `short:"p" long:"pass" default:"abcdefghijklmnopqrstuvwxyz" required:"true" description:"Password is just any letters combination of any size and used for encoding/decoding for this file. It MUST BE the same for encoding and decoding of one image."`
	Msg     string `short:"m" long:"message" required:"true" description:"Message which should be encoded (or decoding verification, not nessesary)."`
	MsgLen  int    `short:"l" long:"len" required:"true" description:"Length of message. MUST BE known to decoder and it's equal to 1 Msg's symbol = 32 bits."`
	Robust  int    `short:"r" long:"robust" default:"20" required:"true" description:"The main parameter of encoding. More Robust cause more visible hidden message, but it is more stable for compression. Value 20 is fine for most cases. 50 is visible, but image is not corrupted."`
	Action  string `short:"a" long:"action" required:"true" description:"Available values: d (decode), e (encode), b (benchmark)"`
	PathIn  string `short:"i" long:"input" required:"true" description:"Path to input files/dir"`
	PathOut string `short:"o" long:"output" required:"false" description:"Path to output files dir"`
}

// Password is just any letters combination of any size and MUST BE the same for encoding and decoding of one image
//var pass = flag.String("pass", "abcdefghijklmnopqrstuvwxyz", "Password  is just any letters combination of any size and used for encoding/decoding for this file. It MUST BE the same for encoding and decoding of one image.")

// Message which should be encoded
//var Msg = flag.String("Msg", "Test message.", "Message which should be encoded (or decoding verification, not nessesary)")

// Length of message. MUST BE known to decoder and it's equal to 1 Msg's symbol = 32 bits.
// Multiply len(Msg) by 32 or use large number to cover all message bits.
// Default value: 32*len(Msg) will be printed, if -l argument will not be used.
// If MsgLen for decoding < MsgLen for encoding, then message will be cut.
// 8x8 pixel block contains 1 bit of message. At least 32 * 8x8 (image with size 128x128) pixel blocks required to encode one symbol.
//var MsgLen = flag.Int("len", 0, "Length of message. MUST BE known to decoder and it's equal to 1 Msg's symbol = 32 bits.") // bits

// The main parameter of encoding. More Robust cause more visible hidden message, but it is more stable for compression
// Value 20 is fine for most cases. 50 is visible, but image is not corrupted
//var Robust = flag.Int("Robust", 20, "The main parameter of encoding. More Robust cause more visible hidden message, but it is more stable for compression. Value 20 is fine for most cases. 50 is visible, but image is not corrupted.")

// You can use "encode", "decode" or "bench" (encode and decode together)
const (
	encodeAction = iota
	decodeAction
	benchAction
)

var Action = decodeAction

// Input file/dir
var pathIn = ""

// Output file/dir
var pathOut = ""

func printHelp() {
	var options flags.Options = 1
	p := flags.NewParser(&opts, options)
	p.WriteHelp(os.Stdout)
}

func fillArgs(args []string) (err error) {
	for i := 0; i < len(args); i++ {
		//switch args[i] {
		/*
			case "-m":
				if i+1 < len(args) {
					Msg = args[i+1]
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
						Robust = r
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
						MsgLen = l
					}
					i++
				} else {
					return fmt.Errorf("%s requeires an argument", args[i])
				}
			case "-d":
				Action = decodeAction
			case "-e":
				Action = encodeAction
			case "-b":
				Action = benchAction
			default:
				if i > 0 {
					return fmt.Errorf("%s unknown argument", args[i])
				}
		*/
		//}
	}
	return nil
}

func concRun(procAction int, dirIn string, dirOut string) ([]string, []error, error) {
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
				switch procAction {
				case encodeAction:
					stackError[i] = fileStego.EncodeFile(dirIn+"//"+fList[i].Name(), dirOut+"//"+fList[i].Name(), opts.Msg, opts.Pass, opts.MsgLen, opts.Robust, -opts.Robust)
				case decodeAction:
					stackValue[i], stackError[i] = fileStego.DecodeFile(dirIn+"//"+fList[i].Name(), opts.Pass, opts.MsgLen)
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
	//printHelp()
	//args := os.Args
	//err := fillArgs(args)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	if _, err := flags.Parse(&opts); err != nil {
		fmt.Println("")
		printHelp()
		return
	}
	//fmt.Println(err)
	if opts.MsgLen == 0 {
		switch Action {
		case encodeAction, benchAction:
			// ??????????????????
			opts.MsgLen = len(opts.Msg) * 32
			fmt.Printf("Length of a message: %d. Use it for decoding.\n", opts.MsgLen)
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
				// ??????????????????????????????????
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

	switch Action {
	case benchAction:
		fallthrough
	case encodeAction:
		if !fi.IsDir() {
			if err := fileStego.EncodeFile(pathIn, pathOut, opts.Msg, opts.Pass, opts.MsgLen, opts.Robust, -opts.Robust); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), pathIn)
			} else {
				fmt.Println("Message encoded.")
			}
		} else {
			if Msgs, errs, err := concRun(encodeAction, pathIn, pathOut); err != nil {
				fmt.Println(err)
			} else {
				errCount := 0
				for i := 0; i < len(Msgs); i++ {
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
		if Action != benchAction {
			break
		}
		fallthrough
	case decodeAction:
		if Action == benchAction {
			pathIn = pathOut
		}
		if !fi.IsDir() {
			if MsgDecoded, err := fileStego.DecodeFile(pathIn, opts.Pass, opts.MsgLen); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), pathIn)
			} else {
				fmt.Printf("%s\n", MsgDecoded)
			}
		} else {
			if MsgsDecoded, errs, err := concRun(decodeAction, pathIn, pathOut); err != nil {
				fmt.Println(err)
			} else {
				fList, err := ioutil.ReadDir(pathIn)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				for i := 0; i < len(MsgsDecoded); i++ {
					if i < len(errs) && errs[i] != nil {
						fmt.Printf("Error: %s for %s\n", errs[i].Error(), pathIn)
					} else {
						fmt.Printf("[%s]\t\"%s\"\n", fList[i].Name(), MsgsDecoded[i])
					}
				}
			}
		}
	}
}
