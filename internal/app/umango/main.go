package main

import (
	"fmt"
	"log"

	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/ui"
)

func main() {
	dataStore, err := data.Init()
	if err != nil {
		log.Fatalf("error loading data store: %v", err)
	}
	fmt.Println((*dataStore).FactorNames[10680103])

	ui.App()
}
