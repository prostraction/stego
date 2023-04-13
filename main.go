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
var path = ""

// Output file/dir
var pathTo = ""

func printHelp() {

}

func fillArgs(args []string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m":
			if i+1 < len(args) {
				msg = args[i+1]
				i++
			} else {
				fmt.Println(args[i], "requeires an argument")
				printHelp()
				return
			}
		case "-p":
			if i+1 < len(args) {
				path = args[i+1]
				i++
			} else {
				fmt.Println(args[i], "requeires an argument")
				printHelp()
				return
			}
		case "-o":
			if i+1 < len(args) {
				pathTo = args[i+1]
				i++
			} else {
				fmt.Println(args[i], "requeires an argument")
				printHelp()
				return
			}
		case "-r":
			if i+1 < len(args) {
				r, err := strconv.Atoi(args[i+1])
				if err != nil {
					fmt.Println(err.Error())
					fmt.Println("Using default value ", robust)
				} else {
					robust = r
				}
				i++
			} else {
				fmt.Println(args[i], "requeires an argument")
				printHelp()
				return
			}
		case "-l":
			if i+1 < len(args) {
				l, err := strconv.Atoi(args[i+1])
				if err != nil {
					fmt.Println(err.Error())
					fmt.Println("Using default value ", msgLen)
				} else {
					msgLen = l
				}
				i++
			} else {
				fmt.Println(args[i], "requeires an argument")
				printHelp()
				return
			}
		case "-d":
			dirProc = true
		}
	}
}

func main() {
	args := os.Args
	fillArgs(args)
	switch operation {
	case "decode":
		fmt.Println(fileStego.DecodeFile(path, pass, msgLen))
	case "encode":
		fileStego.EncodeFile(path, pathTo, msg, pass, msgLen, robust, -robust)
	case "bench":

	}
}
