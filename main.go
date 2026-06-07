package main

import (
	"embed"
	"log"

	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/ui"
)

var (
	//go:embed assets/*
	assets embed.FS
)

func main() {
	dataStore, err := data.Init()
	if err != nil {
		log.Fatalf("error loading data store: %v", err)
	}

	err = ui.App(assets, dataStore)
	if err != nil {
		log.Fatalf("fatal error occurred: %v", err)
	}
}
