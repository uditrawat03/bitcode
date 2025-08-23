package utils

import (
	"log"
	"os"
	"path/filepath"
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func GetLogger(filename string) (*log.Logger, *os.File) {
	dir := filepath.Dir(filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("Failed to create log directory: " + err.Error())
	}

	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}

	logger := log.New(logFile, "[log] ", log.Ldate|log.Ltime|log.Lshortfile)
	return logger, logFile
}
