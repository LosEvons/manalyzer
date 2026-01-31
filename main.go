package main

import (
	"log"

	gui "manalyzer/src"
)

func main() {
	ui := gui.New()
	if err := ui.Start(); err != nil {
		log.Fatalf("UI error %v", err)
	}
}
