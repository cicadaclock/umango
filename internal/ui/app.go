package ui

import (
	"embed"
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/cicadaclock/umango/internal/ui/apptheme"
	"github.com/cicadaclock/umango/internal/ui/pages"
	"golang.org/x/sync/errgroup"
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
	a.Settings().SetTheme(apptheme.NewAppTheme(font))
	window := a.NewWindow("Umango")
	window.Resize(windowSize)
	window.SetContent(loadingScreen())

	// Async load master.mdb and the saved races
	// TODO: Also load veteran list from pre-existing config file?
	go func() {
		var (
			dataStore *data.DataStore
			resultSet races.TeamTrialResultSet
		)

		var g errgroup.Group
		g.SetLimit(runtime.GOMAXPROCS(0))
		g.Go(func() error {
			dataStore, err = data.Init()
			if err != nil {
				return err
			}
			return nil
		})
		g.Go(func() error {
			resultSet, err = pages.LoadTeamTrialResults()
			if err != nil {
				return err
			}
			return nil
		})
		err := g.Wait()
		if err != nil {
			fyne.Do(func() {
				window.SetContent(loadingErrorScreen(err))
			})
		}

		ttData := pages.NewTeamTrialsData(dataStore, resultSet)

		fyne.Do(func() {
			window.SetContent(mainMenu(dataStore, ttData, window))
		})
	}()

	window.ShowAndRun()
	return nil
}

func loadingScreen() *fyne.Container {
	progress := widget.NewProgressBarInfinite()
	return container.NewCenter(container.NewVBox(
		widget.NewLabel("Loading master.mdb and saved races..."),
		progress,
	))
}

func loadingErrorScreen(err error) *fyne.Container {
	label := widget.NewLabel(fmt.Sprintf("Failed to load master.mdb:\n%v", err))
	label.Alignment = fyne.TextAlignCenter
	return container.NewCenter(label)
}

func mainMenu(dataStore *data.DataStore, ttData pages.TeamTrialsData, window fyne.Window) *fyne.Container {
	tabs := container.NewAppTabs(
		container.NewTabItem("TT Chart", pages.NewTeamTrialsPage(ttData)),
		container.NewTabItem("Veteran List", pages.VeteranList(dataStore, window)),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	c := container.NewStack(tabs)
	return c
}
