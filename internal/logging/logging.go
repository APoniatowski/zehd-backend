package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func LogIt(logFunction string, logOutput string, message string) {
	errCloseLogger := Logger(logFunction, logOutput, message)
	if errCloseLogger != nil {
		log.Println(errCloseLogger)
	}
}

func Logger(logFunction, logOutput, message string) error {
	currentDate := time.Now().Format("2006-01-02 15:04:05")
	pathString := os.Getenv("HOME") + "/log/"
	path, _ := filepath.Abs(pathString)
	err := os.MkdirAll(path, os.ModePerm)
	if err == nil || os.IsExist(err) {
		logFile, err := os.OpenFile(pathString+"backend.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer func() {
			errClose := logFile.Close()
			if errClose != nil {
				log.Println(errClose)
			}
		}()
		logger := log.New(logFile, "", log.LstdFlags)
		logger.SetPrefix(currentDate)
		logger.Print(logFunction + " [ " + logOutput + " ] ==> " + message)
	} else {
		return err
	}
	fmt.Println(logFunction + " [ " + logOutput + " ] ==> " + message)
	return nil
}
