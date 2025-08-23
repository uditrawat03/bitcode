package main

import (
	"context"

	"github.com/uditrawat03/bitcode/internal/app"
	"github.com/uditrawat03/bitcode/internal/utils"
)

func main() {
	ctx := context.Background() // top-level context for LSP calls

	// Initialize logger
	logger, logFile := utils.GetLogger("./log/bitcode.log")
	defer logFile.Close()

	logger.Println("Here is the bitcode application started")

	// Create app
	myApp := app.CreateApp(ctx, logger)

	// Ensure proper cleanup
	defer myApp.Shutdown()

	// Initialize app
	if err := myApp.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize app: %v", err)
	}

	// Run app
	myApp.Run()
}
