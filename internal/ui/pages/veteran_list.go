package pages

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/ui/centersteppedlayout"
	"github.com/cicadaclock/umango/internal/ui/veteranwidget"
)

func VeteranList(dataStore *data.DataStore, window fyne.Window) *fyne.Container {
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
