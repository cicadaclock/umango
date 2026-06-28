package pages

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/races"
)

func NewTeamTrialsPage(dataStore *data.DataStore) *fyne.Container {
	// Get data, hardcoded path for now
	home, _ := os.UserHomeDir()
	resultSet, _ := races.LoadRacesFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
	tableData := races.NewTableData(dataStore, resultSet)

	table := newVetTable(tableData)

	return container.NewStack(table)
}

// newVetTable summarizes all sampled races
func newVetTable(tableData races.TableData) *widget.Table {
	headers := tableData.Headers()

	// column-oriented for better data parsing
	cols := tableData.Columns()

	table := widget.NewTable(
		func() (int, int) {
			return tableData.Len(), len(headers)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			text := cols[id.Col][id.Row]
			// SetText calls Refresh(), so set the text only when we need to
			if label.Text != text {
				label.SetText(text)
			}
		},
	)

	table.ShowHeaderRow = true
	table.CreateHeader = func() fyne.CanvasObject {
		label := widget.NewLabel("")
		label.TextStyle.Bold = true
		return label
	}
	table.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		label := cell.(*widget.Label)
		// SetText calls Refresh(), so set the text only when we need to
		if label.Text != headers[id.Col] {
			label.SetText(headers[id.Col])
		}
	}

	for col, header := range headers {
		table.SetColumnWidth(col, float32(len(header))*9+24)
	}

	return table
}
