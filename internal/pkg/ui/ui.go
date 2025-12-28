package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/veteran"
)

var (
	windowWidth, windowHeight float32 = 640, 360
	windowSize                        = fyne.NewSize(windowWidth, windowHeight)
)

func App(dataStore *data.DataStore) {
	a := app.New()
	// Use a custom theme that returns the default theme
	// if I need to override the defaults at any point.
	a.Settings().SetTheme(&myTheme{})
	window := a.NewWindow("Umango")
	window.Resize(windowSize)

	window.SetContent(mainMenu(dataStore))
	window.ShowAndRun()
}

func mainMenu(dataStore *data.DataStore) *fyne.Container {
	tabs := container.NewAppTabs(
		container.NewTabItem("Veterans", createTable()),
		container.NewTabItem("Optimizer", defaultWidget()),
		container.NewTabItem("Temp", temp(dataStore)),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	c := container.NewStack(tabs)
	return c
}

func temp(dataStore *data.DataStore) *VeteranWidget {
	veteran := veteran.Veteran{
		FactorIdArray: []int{
			303,
			3202,
			1000401,
			1001101,
			2003501,
			2004901,
			2010503,
			2011601,
			2015603,
		},
	}
	veteranWidget := NewVeteranView(dataStore, veteran)
	return veteranWidget
}

func createTable() *ColumnGrid {
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
