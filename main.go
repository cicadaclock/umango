// Umango displays the team trial score statistics of each veteran from the saved races.
package main

import (
	"embed"
	"log"

	"github.com/cicadaclock/umango/internal/ui"
)

// go generate writes the Windows application icon resource.
//go:generate go run ./cmd/appicon

var (
	// Font files embedded in the executable
	//go:embed assets/font
	assets embed.FS
)

// main runs the user interface and terminates the process when it fails to start
func main() {
	err := ui.App(assets)
	if err != nil {
		log.Fatalf("fatal error occurred: %v", err)
	}
}
