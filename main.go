package main

import (
	"fmt"
	"os"
	"stego/src/fileStego"
	"strconv"
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
var operation = "decode"

// You can set dirProc to true for encoding/decoding each image in directory
var dirProc = false

// Input file/dir
var pathIn = ""

// Output file/dir
var pathOut = ""

func printHelp() {

}

func fillArgs(args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m":
			if i+1 < len(args) {
				msg = args[i+1]
				i++
			} else {
				var e error = fmt.Errorf("%s requeires an argument", args[i])
				return e
			}
		case "-p":
			if i+1 < len(args) {
				pathIn = args[i+1]
				i++
			} else {
				var e error = fmt.Errorf("%s requeires an argument", args[i])
				return e
			}
		case "-o":
			if i+1 < len(args) {
				pathOut = args[i+1]
				i++
			} else {
				var e error = fmt.Errorf("%s requeires an argument", args[i])
				return e
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
				var e error = fmt.Errorf("%s requeires an argument", args[i])
				return e
			}
		case "-l":
			if i+1 < len(args) {
				l, err := strconv.Atoi(args[i+1])
				if err != nil {
					return err
				} else {
					msgLen = l
				}
				i++
			} else {
				var e error = fmt.Errorf("%s requeires an argument", args[i])
				return e
			}
		case "-d":
			operation = "decode"
		case "-e":
			operation = "encode"
		case "-b":
			operation = "bench"
		}
	}
	return nil
}

func main() {
	args := os.Args
	err := fillArgs(args)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if msgLen == 0 {
		if operation == "encode" || operation == "bench" {
			msgLen = len(msg) * 32
			fmt.Printf("Length of message: %d. Use it for decoding.\n", msgLen)
		} else {
			fmt.Println("Specify length of message!")
			return
		}
	}
	fi1, err := os.Stat(pathIn)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if pathOut == "" {
		if fi1.IsDir() {
			pathOut = fi1.Name() + "stego"
			os.Mkdir(pathOut, os.ModePerm)
			_, verifyPermErr := os.Stat(pathOut)
			if verifyPermErr != nil {
				fmt.Println("No output directory was specified. Unable to create a new directory", pathOut, "aborting.")
				fmt.Println(verifyPermErr.Error())
				return
			}
		} else {

		}
	}

	switch operation {
	case "decode":
		// msg, err :=
		fmt.Println(fileStego.DecodeFile(pathIn, pass, msgLen))
	case "encode":
		// err :=
		fileStego.EncodeFile(pathIn, pathOut, msg, pass, msgLen, robust, -robust)
	case "bench":
		// res, err :=
	}
}
