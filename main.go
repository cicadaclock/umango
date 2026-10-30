package main

import (
	"embed"
	"log"

	"github.com/cicadaclock/umango/internal/ui"
)

//go:generate go run ./cmd/appicon

var (
	//go:embed assets/font
	assets embed.FS
)

func main() {
	err := ui.App(assets)
	if err != nil {
		log.Fatalf("fatal error occurred: %v", err)
	}
}
