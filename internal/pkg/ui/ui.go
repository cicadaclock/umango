package ui

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/races"
	"github.com/cicadaclock/umango/internal/pkg/ui/app_theme"
	centersteppedlayout "github.com/cicadaclock/umango/internal/pkg/ui/center_stepped_layout"
	"github.com/cicadaclock/umango/internal/pkg/ui/veteranwidget"
)

var (
	windowWidth, windowHeight float32 = 1280, 720
	windowSize                        = fyne.NewSize(windowWidth, windowHeight)
)

func App(assets embed.FS) error {
	a := app.New()
	// Use a custom theme that returns the default theme
	// if I need to override the defaults at any point.
	font, err := assets.ReadFile("assets/font/Inter-VariableFont_opsz,wght.ttf")
	if err != nil {
		return fmt.Errorf("read assets: %w", err)
	}
	a.Settings().SetTheme(app_theme.NewAppTheme(font))
	window := a.NewWindow("Umango")
	window.Resize(windowSize)
	window.SetContent(loadingScreen())

	// Async load master.mdb, shouldn't be too long but don't want to hang the UI
	// TODO: Also load veteran list from pre-existing config file?
	go func() {
		dataStore, err := data.Init()
		fyne.Do(func() {
			if err != nil {
				window.SetContent(loadingErrorScreen(err))
				return
			}
			window.SetContent(mainMenu(dataStore, window))
		})
	}()

	window.ShowAndRun()
	return nil
}

func loadingScreen() *fyne.Container {
	progress := widget.NewProgressBarInfinite()
	return container.NewCenter(container.NewVBox(
		widget.NewLabel("Loading master.mdb..."),
		progress,
	))
}

func loadingErrorScreen(err error) *fyne.Container {
	label := widget.NewLabel(fmt.Sprintf("Failed to load master.mdb:\n%v", err))
	label.Alignment = fyne.TextAlignCenter
	return container.NewCenter(label)
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("get home dir: %v", err)
	}
	dir := storage.NewFileURI(filepath.Join(homeDir, "Documents", "Saved races"))
	listDir, err := storage.ListerForURI(dir)
	if err != nil {
		log.Fatalf("lister home dir uri: %v", err)
	}
	veteranFileDialog.SetLocation(listDir)
	veteranFileDialog.Resize(fyne.NewSize(500, 500))
	loadVeteranButton := widget.NewButton("Load",
		func() {
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

func testTTData() {
	home, _ := os.UserHomeDir()
	results, _ := races.LoadRaceResultsFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
	soa := races.NewRaceResultsSoA(results)
	fmt.Println("Unique charas: ", soa.UniqueCharas())
	for k, v := range soa.CharaResultSoA() {
		fmt.Println(k, v.TotalScoreAverage(), v.TotalScore)
	}
	fmt.Println("Team")
	mile := soa.FilterByDistanceType(1)
	fmt.Println(mile.TotalScoreAverage(), mile.TeamTotalScores)
}
