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
					stackError[i] = fileStego.EncodeFile(dirIn+"//"+fList[i].Name(), dirOut+"//"+fList[i].Name(), Opts.Msg, Opts.Pass, Opts.MsgLen, Opts.Robust, -Opts.Robust)
				case decodeAction:
					stackValue[i], stackError[i] = fileStego.DecodeFile(dirIn+"//"+fList[i].Name(), Opts.Pass, Opts.MsgLen)
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
		Opts.PathIn = Opts.PathOut
	}

	if !fileInfo.IsDir() {
		if action == "encode" {
			fmt.Println(Opts.PathOut)
			if err := fileStego.EncodeFile(Opts.PathIn, Opts.PathOut, Opts.Msg, Opts.Pass, Opts.MsgLen, Opts.Robust, -Opts.Robust); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), Opts.PathIn)
			} else {
				fmt.Println("Message encoded.")
			}
		} else if action == "decode" {
			if MsgDecoded, err := fileStego.DecodeFile(Opts.PathIn, Opts.Pass, Opts.MsgLen); err != nil {
				fmt.Printf("Error: %s for %s\n", err.Error(), Opts.PathIn)
			} else {
				fmt.Printf("%s\n", MsgDecoded)
			}
		}
	} else {
		fmt.Println(Opts.PathIn, Opts.PathOut)
		var Msgs []string
		var errs []error
		var err error

		if action == "encode" {
			Msgs, errs, err = concRun(encodeAction, Opts.PathIn, Opts.PathOut)
		} else {
			Msgs, errs, err = concRun(decodeAction, Opts.PathIn, Opts.PathOut)
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		fList, err := ioutil.ReadDir(Opts.PathIn)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		errCount := 0
		for i := 0; i < len(Msgs); i++ {
			if i < len(errs) && errs[i] != nil {
				fmt.Printf("Error: %s for %s\n", errs[i].Error(), Opts.PathIn)
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
	fileInfo, err := os.Stat(Opts.PathIn)
	if err != nil {
		return nil, fmt.Errorf("error stating path: %w (%s)", err, Opts.PathIn)
	}

	if Opts.PathOut == "" {
		if fileInfo.IsDir() {
			return createStegoDir(fileInfo.Name())
		}
		return createStegoFile(Opts.PathIn)
	}

	return createOutputPath(fileInfo)
}

func createStegoDir(dirName string) (os.FileInfo, error) {
	Opts.PathOut = dirName + "_stego"
	err := os.Mkdir(Opts.PathOut, os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			return os.Stat(Opts.PathOut)
		}
		return nil, fmt.Errorf("unable to create directory %s: %w", Opts.PathOut, err)
	}
	return os.Stat(Opts.PathOut)
}

func createStegoFile(PathIn string) (os.FileInfo, error) {
	name, ext := splitFileName(Opts.PathIn)
	if name == "" {
		return nil, fmt.Errorf("no file type specified for %s. Aborting.", Opts.PathIn)
	}
	Opts.PathOut = name + "_stego" + ext
	if _, err := os.Create(Opts.PathOut); err != nil {
		return nil, fmt.Errorf("unable to create file %s: %w", Opts.PathOut, err)
	}
	return os.Stat(Opts.PathOut)
}

func createOutputPath(fileInfo os.FileInfo) (os.FileInfo, error) {
	if fileInfo.IsDir() {
		if err := os.Mkdir(Opts.PathOut, os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to create directory %s: %w", Opts.PathOut, err)
		}
	} else {
		if _, err := os.Create(Opts.PathOut); err != nil {
			return nil, fmt.Errorf("unable to create file %s: %w", Opts.PathOut, err)
		}
	}

	return os.Stat(Opts.PathOut)
}

func splitFileName(filePath string) (string, string) {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			return filePath[:i], filePath[i:]
		}
	}
	return "", ""
}
