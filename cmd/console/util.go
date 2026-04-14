package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"stego/internal/stego_loader"
	"stego/pkg/video"

	"golang.org/x/sync/errgroup"
)

func runFileOperation(opts commonOptions, action string, message string) error {
	fileInfo, err := os.Stat(opts.pathIn)
	if err != nil {
		return fmt.Errorf("error stating path: %w (%s)", err, opts.pathIn)
	}

	// Prepare output path if not specified
	if opts.pathOut == "" {
		if fileInfo.IsDir() {
			opts.pathOut = fileInfo.Name() + "_stego"
		} else {
			name := filepath.Base(opts.pathIn)
			ext := filepath.Ext(name)
			base := name[:len(name)-len(ext)]
			opts.pathOut = base + "_stego" + ext
		}
	}

	if !fileInfo.IsDir() {
		return processSingleFile(opts, action, message)
	}

	return processDirectory(opts, action, message)
}

func processSingleFile(opts commonOptions, action string, message string) error {
	var err error
	if action == "encode" {
		if opts.video {
			err = video.EncodeVideo(opts.pathIn, opts.pathOut, message, opts.pass, opts.msgLen, opts.robust, -opts.robust)
		} else {
			err = stego_loader.EncodeFile(opts.pathIn, opts.pathOut, message, opts.pass, opts.msgLen, opts.robust, -opts.robust)
		}
		if err == nil {
			fmt.Printf("Message encoded successfully: %s -> %s\n", opts.pathIn, opts.pathOut)
		}
	} else if action == "decode" {
		var msgDecoded string
		if opts.video {
			msgDecoded, err = video.DecodeVideo(opts.pathIn, opts.pass, opts.msgLen)
		} else {
			msgDecoded, err = stego_loader.DecodeFile(opts.pathIn, opts.pass, opts.msgLen)
		}
		if err == nil {
			fmt.Printf("[%s] Decoded message: %q\n", opts.pathIn, msgDecoded)
		}
	}
	return err
}

func processDirectory(opts commonOptions, action string, message string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(opts.pathOut, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	files, err := ioutil.ReadDir(opts.pathIn)
	if err != nil {
		return fmt.Errorf("failed to read input directory: %w", err)
	}

	g, _ := errgroup.WithContext(context.Background())
	// Limit concurrency to avoid too many open files or overwhelming the CPU/ffmpeg
	g.SetLimit(10)

	for _, f := range files {
		if f.IsDir() {
			continue // Skip nested directories for now
		}

		fileName := f.Name()
		inputFilePath := filepath.Join(opts.pathIn, fileName)
		outputFilePath := filepath.Join(opts.pathOut, fileName)

		g.Go(func() error {
			fileOpts := opts
			fileOpts.pathIn = inputFilePath
			fileOpts.pathOut = outputFilePath

			if action == "encode" {
				var err error
				if opts.video {
					err = video.EncodeVideo(inputFilePath, outputFilePath, message, opts.pass, opts.msgLen, opts.robust, -opts.robust)
				} else {
					err = stego_loader.EncodeFile(inputFilePath, outputFilePath, message, opts.pass, opts.msgLen, opts.robust, -opts.robust)
				}
				if err != nil {
					return fmt.Errorf("file %s: %w", fileName, err)
				}
				fmt.Printf("Encoded: %s\n", fileName)
			} else {
				var msgDecoded string
				var err error
				if opts.video {
					msgDecoded, err = video.DecodeVideo(inputFilePath, opts.pass, opts.msgLen)
				} else {
					msgDecoded, err = stego_loader.DecodeFile(inputFilePath, opts.pass, opts.msgLen)
				}
				if err != nil {
					return fmt.Errorf("file %s: %w", fileName, err)
				}
				fmt.Printf("[%s] Decoded: %q\n", fileName, msgDecoded)
			}
			return nil
		})
	}

	return g.Wait()
}
