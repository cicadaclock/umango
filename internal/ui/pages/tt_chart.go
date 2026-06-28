package pages

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/s-daehling/fyne-charts/pkg/coord"
	gdata "github.com/s-daehling/fyne-charts/pkg/data"
	"github.com/s-daehling/fyne-charts/pkg/style"
)

func NewTeamTrialsPage(dataStore *data.DataStore) *fyne.Container {
	// Get data, hardcoded path for now
	home, _ := os.UserHomeDir()
	resultSet, _ := races.LoadRacesFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))

	// TT veteran table
	tableData := races.NewTableData(dataStore, resultSet)
	cols := tableData.Columns()
	table := newVetTable(tableData.Headers(), cols, tableData.ColumnWidths())
	// Filter buttons for TT veteran table

	// Individual score histogram
	scores := resultSet.GetMyScores()
	histogram := newScoreHistogram(*scores[2928])

	// Skill table
	skillTable := container.NewWithoutLayout()

	// Page containers
	rightSide := container.NewVSplit(histogram, skillTable)
	rightSide.SetOffset(0.4)
	split := container.NewHSplit(table, rightSide)
	split.SetOffset(0.7)
	return container.NewStack(split)
}

func newScoreHistogram(scoreArray races.ScoreArray) fyne.CanvasObject {
	// Labels
	chart := coord.NewCartesianNumericalChart("Score vs. Frequency")
	chart.SetXAxisLabel("Score")
	chart.SetYAxisLabel("Frequency")

	// Color
	pal := style.NewPaletteTriadic(theme.ColorNamePrimary)
	pal = style.NewPaletteLightDarkSet(pal.Names())

	// Data
	finalScoreNps := make([]*coord.NumericalPointSeries, 1)
	finalScoreData := []gdata.NumericalPoint{}
	steps := 10
	xPts, yPts := scoreArray.Density(steps)
	for i := range xPts {
		point := gdata.NumericalPoint{
			N:   float64(xPts[i]),
			Val: float64(yPts[i]),
		}
		finalScoreData = append(finalScoreData, point)
	}

	// Populate data
	nps, err := coord.NewNumericalPointSeries("data", pal.Next(), finalScoreData)
	if err != nil {
		log.Fatalf("error creating nps: %v", err)
	}
	finalScoreNps = append(finalScoreNps, nps)
	for _, nps := range finalScoreNps {
		_ = chart.AddBarSeries(nps, float64(scoreArray.StepSize(steps)))
	}

	return chart
}

// newVetTable summarizes all sampled races
func newVetTable(headers []string, cols [][]string, colWidths []int) *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(cols[0]), len(headers)
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

	for col, length := range colWidths {
		table.SetColumnWidth(col, (float32(length)*7)+24)
	}

	return table
}
