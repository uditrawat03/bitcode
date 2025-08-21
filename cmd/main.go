package main

import (
	"log"

	"github.com/uditrawat03/bitcode/internal/app"
)

func main() {
	app := app.CreateApp()

	// Ensure proper cleanup
	defer app.Shutdown()

	if err := app.Initialize(); err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	app.Run()
}
