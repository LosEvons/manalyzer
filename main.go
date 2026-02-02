package main

import (
	"log"

	gui "manalyzer/src"
)

func main() {
	// Initialize logger
	if err := gui.InitLogger(); err != nil {
		log.Printf("Warning: Failed to initialize logger: %v", err)
	}
	defer gui.CloseLogger()

	// Add panic recovery at the top level
	defer func() {
		if r := recover(); r != nil {
			gui.LogPanic(r)
			log.Fatalf("Fatal panic: %v", r)
		}
	}()

	gui.LogInfo("Starting Manalyzer")

	ui := gui.New()
	if err := ui.Start(); err != nil {
		gui.LogError("UI error: %v", err)
		log.Fatalf("UI error %v", err)
	}

	gui.LogInfo("Manalyzer exited normally")
}
