package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func App() {
	a := app.New()
	window := a.NewWindow("Hello")
	window.Resize(fyne.NewSize(640, 360))

	window.SetContent(CreateTable())

	window.ShowAndRun()
}

func CreateTable() *ColumnGrid {
	columnGrid := NewColumnGrid([][]string{
		{"reallylongname1", "name2", "name3", "name4", "name5", "name6", "name1", "name2", "name3", "name4", "name5", "name6"},
		{"addr1", "addr2", "addr3"},
		{"addr21", "addr2", "addr3"},
		{"addr31", "addr2", "addr3"},
		{"addr41", "addr2", "addr3"},
		{"addr51", "addr2", "addr3"},
		{"addr61", "addr2", "addr3"},
		{"addr71", "addr2", "addr3"},
	})
	columnGrid.ColumnPadding = 10.0
	columnGrid.RowPadding = 5.0
	return columnGrid
}

func DefaultWidget() *fyne.Container {
	text := widget.NewLabel("Hello Fyne!")
	c := container.NewVBox(
		text,
		widget.NewButton("Hi!", func() {
			text.SetText("Welcome :)")
		}),
	)
	return c
}
