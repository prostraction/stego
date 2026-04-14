package main

import (
	"flag"
	"fmt"
	"os"
	"stego/internal/config"
	"stego/internal/telegram_bot"
)

// Version is the current version of the application.
const Version = "1.0.0"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	config.LoadConfig()

	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	switch os.Args[1] {
	case "encode", "e":
		return handleEncode(os.Args[2:])
	case "decode", "d":
		return handleDecode(os.Args[2:])
	case "bench", "b":
		return handleBench(os.Args[2:])
	case "bot":
		return handleBot(os.Args[2:])
	case "-V", "--version", "version":
		fmt.Printf("Go-Stego version %s\n", Version)
		return nil
	case "-h", "--help", "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown subcommand %q", os.Args[1])
	}
}

func printUsage() {
	fmt.Println("Go-Stego: A steganography tool using DCT")
	fmt.Println("\nUsage:")
	fmt.Println("  stego [subcommand] [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  encode, e  Encode a message into an image or video")
	fmt.Println("  decode, d  Decode a message from an image or video")
	fmt.Println("  bench, b   Perform both encoding and decoding for benchmarking")
	fmt.Println("  bot        Start the Telegram bot interface")
	fmt.Println("  version    Print version information")
	fmt.Println("  help       Print this help message")
	fmt.Println("\nUse 'stego [subcommand] --help' for more information on a subcommand.")
}

type commonOptions struct {
	pass    string
	robust  int
	pathIn  string
	pathOut string
	video   bool
	msgLen  int
}

func handleEncode(args []string) error {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	msg := fs.String("m", "", "Message to encode")
	pass := fs.String("p", "abcdefghijklmnopqrstuvwxyz", "Password for encoding")
	robust := fs.Int("r", 20, "Robustness parameter (default 20)")
	pathIn := fs.String("i", "", "Input file or directory")
	pathOut := fs.String("o", "", "Output file or directory (optional)")
	video := fs.Bool("v", false, "Treat input as video")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *msg == "" {
		return fmt.Errorf("message (-m) is required for encoding")
	}
	if *pathIn == "" {
		return fmt.Errorf("input path (-i) is required")
	}

	opts := commonOptions{
		pass:    *pass,
		robust:  *robust,
		pathIn:  *pathIn,
		pathOut: *pathOut,
		video:   *video,
		msgLen:  len(*msg) * 32,
	}

	fmt.Printf("Length of a message: %d bits. Use it for decoding.\n", opts.msgLen)

	return runFileOperation(opts, "encode", *msg)
}

func handleDecode(args []string) error {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	pass := fs.String("p", "abcdefghijklmnopqrstuvwxyz", "Password for decoding")
	msgLen := fs.Int("l", 0, "Length of message in bits (required for decoding)")
	pathIn := fs.String("i", "", "Input file or directory")
	video := fs.Bool("v", false, "Treat input as video")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *pathIn == "" {
		return fmt.Errorf("input path (-i) is required")
	}
	if *msgLen <= 0 {
		return fmt.Errorf("message length (-l) is required and must be positive for decoding")
	}

	opts := commonOptions{
		pass:   *pass,
		pathIn: *pathIn,
		msgLen: *msgLen,
		video:  *video,
	}

	return runFileOperation(opts, "decode", "")
}

func handleBench(args []string) error {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	msg := fs.String("m", "", "Message to encode")
	pass := fs.String("p", "abcdefghijklmnopqrstuvwxyz", "Password for encoding/decoding")
	robust := fs.Int("r", 20, "Robustness parameter (default 20)")
	pathIn := fs.String("i", "", "Input file or directory")
	pathOut := fs.String("o", "", "Output file or directory (optional)")
	video := fs.Bool("v", false, "Treat input as video")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *msg == "" {
		return fmt.Errorf("message (-m) is required for benchmarking")
	}
	if *pathIn == "" {
		return fmt.Errorf("input path (-i) is required")
	}

	opts := commonOptions{
		pass:    *pass,
		robust:  *robust,
		pathIn:  *pathIn,
		pathOut: *pathOut,
		video:   *video,
		msgLen:  len(*msg) * 32,
	}

	fmt.Printf("Benchmark: Length of a message: %d bits.\n", opts.msgLen)

	if err := runFileOperation(opts, "encode", *msg); err != nil {
		return fmt.Errorf("benchmark encoding failed: %w", err)
	}

	// For benchmark decoding, the input is the output of the encoding phase.
	decodeOpts := opts
	decodeOpts.pathIn = opts.pathOut
	return runFileOperation(decodeOpts, "decode", "")
}

func handleBot(args []string) error {
	fs := flag.NewFlagSet("bot", flag.ExitOnError)
	token := fs.String("t", config.GetEnv("TELEGRAM_BOT_TOKEN", ""), "Telegram Bot Token (or set TELEGRAM_BOT_TOKEN in config/app.env)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *token == "" {
		return fmt.Errorf("telegram bot token is required (-t or TELEGRAM_BOT_TOKEN env var)")
	}

	b := bot.StegoBot{Token: *token}
	return b.Start()
}
