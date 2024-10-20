package main

import (
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
)

var Opts struct {
	Pass    string `short:"p" long:"pass" default:"abcdefghijklmnopqrstuvwxyz" required:"true" description:"Password is just any letters combination of any size and used for encoding/decoding for this file. It MUST BE the same for encoding and decoding of one image."`
	Msg     string `short:"m" long:"message" required:"true" description:"Message which should be encoded (or decoding verification, not nessesary)."`
	MsgLen  int    `short:"l" long:"len" required:"true" description:"Length of message. MUST BE known to decoder and it's equal to 1 Msg's symbol = 32 bits."`
	Robust  int    `short:"r" long:"robust" default:"20" required:"true" description:"The main parameter of encoding. More Robust cause more visible hidden message, but it is more stable for compression. Value 20 is fine for most cases. 50 is visible, but image is not corrupted."`
	Action  string `short:"a" long:"action" required:"true" description:"Available values: d, e, b (decode / encode / benchmark)"`
	PathIn  string `short:"i" long:"input" required:"true" description:"Path to input files/dir"`
	PathOut string `short:"o" long:"output" required:"false" description:"Path to output files dir"`
}

// You can use "encode", "decode" or "bench" (encode and decode together)
const (
	encodeAction = iota
	decodeAction
	benchAction
)

var Action = decodeAction

func printHelp() {
	var options flags.Options = 1
	p := flags.NewParser(&Opts, options)
	p.WriteHelp(os.Stdout)
}

func main() {
	if _, err := flags.Parse(&Opts); err != nil {
		fmt.Println("")
		printHelp()
		return
	}

	switch (Opts.Action) {
		case "b": Action = benchAction
		case "e": Action = encodeAction
		case "d": Action = decodeAction
	}

	if Opts.MsgLen == 0 {
		switch Action {
		case encodeAction, benchAction:
			Opts.MsgLen = len(Opts.Msg) * 32
			fmt.Printf("Length of a message: %d. Use it for decoding.\n", Opts.MsgLen)
		default:
			fmt.Println("Specify length of a message!")
			return
		}
	}

	fileInfo, err := CreateDir()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("called ",  Action)

	switch Action {
	case benchAction:
		fallthrough
	case encodeAction:
		runFileOperation(fileInfo, "encode")
		if Action != benchAction {
			break
		}
		fallthrough
	case decodeAction:
		runFileOperation(fileInfo, "decode")
	}
}
