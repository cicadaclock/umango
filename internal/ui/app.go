package ui

import (
	"embed"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/ui/apptheme"
	"github.com/cicadaclock/umango/internal/ui/pages"
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
		container.NewTabItem("Veterans", pages.Veterans()),
		container.NewTabItem("Optimizer", pages.Optimizer()),
		container.NewTabItem("Veteran List", pages.VeteranList(dataStore, window)),
		container.NewTabItem("TT Chart", pages.NewTeamTrialsPage(dataStore)),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	c := container.NewStack(tabs)
	return c
}
