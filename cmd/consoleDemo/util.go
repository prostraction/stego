package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"stego/internal/fileStego"
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

func runFileOperation(fileInfo os.FileInfo, action string) {
	if Action == benchAction && action == "decode" {
		pathIn = pathOut
	}

	if !fileInfo.IsDir() {
		if action == "encode" {
			if err := fileStego.EncodeFile(pathIn, pathOut, opts.Msg, opts.Pass, opts.MsgLen, opts.Robust, -opts.Robust); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), pathIn)
			} else {
				fmt.Println("Message encoded.")
			}
		} else if action == "decode" {
			if MsgDecoded, err := fileStego.DecodeFile(pathIn, opts.Pass, opts.MsgLen); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), pathIn)
			} else {
				fmt.Printf("%s\n", MsgDecoded)
			}
		}
	} else {
		var Msgs []string
		var errs []error
		var err error

		if action == "encode" {
			Msgs, errs, err = concRun(encodeAction, pathIn, pathOut)
		} else {
			Msgs, errs, err = concRun(decodeAction, pathIn, pathOut)
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		fList, err := ioutil.ReadDir(pathIn)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		errCount := 0
		for i := 0; i < len(Msgs); i++ {
			if i < len(errs) && errs[i] != nil {
				fmt.Printf("Error: %s for %s\n", errs[i].Error(), pathIn)
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
	fileInfo, err := os.Stat(pathIn)
	if err != nil {
		return nil, fmt.Errorf("error stating path: %w", err)
	}

	if pathOut == "" {
		if fileInfo.IsDir() {
			return createStegoDir(fileInfo.Name())
		}
		return createStegoFile(pathIn)
	}

	return createOutputPath(fileInfo)
}

func createStegoDir(dirName string) (os.FileInfo, error) {
	pathOut = dirName + "_stego"
	if err := os.Mkdir(pathOut, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create directory %s: %w", pathOut, err)
	}
	return os.Stat(pathOut)
}

func createStegoFile(pathIn string) (os.FileInfo, error) {
	name, ext := splitFileName(pathIn)
	if name == "" {
		return nil, fmt.Errorf("no file type specified for %s. Aborting.", pathIn)
	}
	pathOut = name + "_stego" + ext
	if _, err := os.Create(pathOut); err != nil {
		return nil, fmt.Errorf("unable to create file %s: %w", pathOut, err)
	}
	return os.Stat(pathOut)
}

func createOutputPath(fileInfo os.FileInfo) (os.FileInfo, error) {
	if fileInfo.IsDir() {
		if err := os.Mkdir(pathOut, os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to create directory %s: %w", pathOut, err)
		}
	} else {
		if _, err := os.Create(pathOut); err != nil {
			return nil, fmt.Errorf("unable to create file %s: %w", pathOut, err)
		}
	}

	return os.Stat(pathOut)
}

func splitFileName(filePath string) (string, string) {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			return filePath[:i], filePath[i:]
		}
	}
	return "", ""
}
