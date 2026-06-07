package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/ui/app_theme"
	centersteppedlayout "github.com/cicadaclock/umango/internal/pkg/ui/center_stepped_layout"
	"github.com/cicadaclock/umango/internal/pkg/ui/veteranwidget"
)

var (
	windowWidth, windowHeight float32 = 1280, 720
	windowSize                        = fyne.NewSize(windowWidth, windowHeight)
)

func App(dataStore *data.DataStore) {
	a := app.New()
	// Use a custom theme that returns the default theme
	// if I need to override the defaults at any point.
	a.Settings().SetTheme(&app_theme.AppTheme{})
	window := a.NewWindow("Umango")
	window.Resize(windowSize)

	window.SetContent(mainMenu(dataStore, window))
	window.ShowAndRun()
}

func mainMenu(dataStore *data.DataStore, window fyne.Window) *fyne.Container {
	tabs := container.NewAppTabs(
		container.NewTabItem("Veterans", createTable()),
		container.NewTabItem("Optimizer", defaultWidget()),
		container.NewTabItem("Temp", temp(dataStore, window)),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	c := container.NewStack(tabs)
	return c
}

func temp(dataStore *data.DataStore, window fyne.Window) *fyne.Container {
	veteranWidget := veteranwidget.NewVeteranWidget(dataStore)
	content := centersteppedlayout.NewHStepped(0.6, 0.8, veteranWidget)
	veteranFileDialog := dialog.NewFileOpen(
		func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
			}
			if reader != nil {
				defer reader.Close()
				dataStore.VeteransJsonFilePath = reader.URI().Path()
				veteranWidget.Load()
				content.Refresh()
			}
		},
		window,
	)
	loadVeteranButton := widget.NewButton("Load",
		func() {
			veteranFileDialog.Resize(fyne.NewSize(500, 500))
			veteranFileDialog.Show()
		},
	)
	header := container.NewHBox(loadVeteranButton)
	main := container.NewBorder(header, nil, nil, nil, content)
	return main
}

func createTable() *ColumnGrid {
	columnGrid := NewColumnGrid([][]string{
		{"reallylongname1", "name2", "name3", "name4", "name5", "name6", "name1", "name2", "name3", "name4", "name5", "name6"},
		{"addr11", "addr2", "addr3"},
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

func defaultWidget() *fyne.Container {
	text := widget.NewLabel("Hello Fyne!")
	c := container.NewVBox(
		text,
		widget.NewButton("Hi!", func() {
			text.SetText("Welcome :)")
		}),
	)
	return c
}
