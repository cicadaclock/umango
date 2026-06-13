package main

import (
	"embed"
	"log"

	"github.com/cicadaclock/umango/internal/ui"
)

var (
	//go:embed assets/*
	assets embed.FS
)

func main() {
	err := ui.App(assets)
	if err != nil {
		log.Fatalf("fatal error occurred: %v", err)
	}
}
