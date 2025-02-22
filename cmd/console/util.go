package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"stego/internal/stego_loader"
	"sync"
)

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
					stackError[i] = stego_loader.EncodeFile(dirIn+"//"+fList[i].Name(), dirOut+"//"+fList[i].Name(), opts.Msg, opts.Pass, opts.MsgLen, opts.Robust, -opts.Robust)
				case decodeAction:
					stackValue[i], stackError[i] = stego_loader.DecodeFile(dirIn+"//"+fList[i].Name(), opts.Pass, opts.MsgLen)
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

func runFileOperation(fileInfo os.FileInfo, action string) {
	fmt.Println(action)
	if Action == benchAction && action == "decode" {
		opts.PathIn = opts.PathOut
	}

	if !fileInfo.IsDir() {
		if action == "encode" {
			fmt.Println(opts.PathOut)
			if err := stego_loader.EncodeFile(opts.PathIn, opts.PathOut, opts.Msg, opts.Pass, opts.MsgLen, opts.Robust, -opts.Robust); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), opts.PathIn)
			} else {
				fmt.Println("Message encoded.")
			}
		} else if action == "decode" {
			if MsgDecoded, err := stego_loader.DecodeFile(opts.PathIn, opts.Pass, opts.MsgLen); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), opts.PathIn)
			} else {
				fmt.Printf("%s\n", MsgDecoded)
			}
		}
	} else {
		fmt.Println(opts.PathIn, opts.PathOut)
		var Msgs []string
		var errs []error
		var err error

		if action == "encode" {
			Msgs, errs, err = concRun(encodeAction, opts.PathIn, opts.PathOut)
		} else {
			Msgs, errs, err = concRun(decodeAction, opts.PathIn, opts.PathOut)
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		fList, err := ioutil.ReadDir(opts.PathIn)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		errCount := 0
		for i := 0; i < len(Msgs); i++ {
			if i < len(errs) && errs[i] != nil {
				fmt.Printf("Error: %s for %s\n", errs[i].Error(), opts.PathIn)
				errCount++
			} else if action == "decode" {
				fmt.Printf("[%s]\t\"%s\"\n", fList[i].Name(), Msgs[i])
			}
		}

		if errCount == 0 && action == "encode" {
			fmt.Println("All messages encoded.")
		}
	}
}

func CreateDir() (os.FileInfo, error) {
	fileInfo, err := os.Stat(opts.PathIn)
	if err != nil {
		return nil, fmt.Errorf("error stating path: %w (%s)", err, opts.PathIn)
	}

	if opts.PathOut == "" {
		if fileInfo.IsDir() {
			return createStegoDir(fileInfo.Name())
		}
		return createStegoFile(opts.PathIn)
	}

	return createOutputPath(fileInfo)
}

func createStegoDir(dirName string) (os.FileInfo, error) {
	opts.PathOut = dirName + "_stego"
	err := os.Mkdir(opts.PathOut, os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			return os.Stat(opts.PathOut)
		}
		return nil, fmt.Errorf("unable to create directory %s: %w", opts.PathOut, err)
	}
	return os.Stat(opts.PathOut)
}

func createStegoFile(PathIn string) (os.FileInfo, error) {
	name, ext := splitFileName(opts.PathIn)
	if name == "" {
		return nil, fmt.Errorf("no file type specified for %s. Aborting.", opts.PathIn)
	}
	opts.PathOut = name + "_stego" + ext
	if _, err := os.Create(opts.PathOut); err != nil {
		return nil, fmt.Errorf("unable to create file %s: %w", opts.PathOut, err)
	}
	return os.Stat(opts.PathOut)
}

func createOutputPath(fileInfo os.FileInfo) (os.FileInfo, error) {
	if fileInfo.IsDir() {
		if err := os.Mkdir(opts.PathOut, os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to create directory %s: %w", opts.PathOut, err)
		}
	} else {
		if _, err := os.Create(opts.PathOut); err != nil {
			return nil, fmt.Errorf("unable to create file %s: %w", opts.PathOut, err)
		}
	}

	return os.Stat(opts.PathOut)
}

func splitFileName(filePath string) (string, string) {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			return filePath[:i], filePath[i:]
		}
	}
	return "", ""
}
