package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
)

func main() {

	var db data.DB
	db.Open()
	defer db.SqlDB.Close()
	dataStore, err := data.Load(db)
	if err != nil {
		log.Fatalf("error loading data store: %v", err)
	}
	fmt.Println((*dataStore).FactorNames[10680103])

	a := app.New()
	w := a.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))

	// w.ShowAndRun()
}
