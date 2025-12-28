package main

import (
	"log"

	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/ui"
)

func main() {
	dataStore, err := data.Init()
	if err != nil {
		log.Fatalf("error loading data store: %v", err)
	}

	ui.App(dataStore)
}
