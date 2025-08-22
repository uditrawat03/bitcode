package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/uditrawat03/bitcode/internal/app"
)

func main() {
	// Initialize logger
	logger, logFile := getLogger("./log/bitcode.log")
	defer logFile.Close()

	logger.Println("Here is the bitcode application started")

	// Create app
	myApp := app.CreateApp(logger)

	// Ensure proper cleanup
	defer myApp.Shutdown()

	// Initialize app
	if err := myApp.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize app: %v", err)
	}

	// Run app
	myApp.Run()
}

func getLogger(filename string) (*log.Logger, *os.File) {
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
