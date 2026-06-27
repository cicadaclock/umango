package pages

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/s-daehling/fyne-charts/pkg/coord"
	gdata "github.com/s-daehling/fyne-charts/pkg/data"
	"github.com/s-daehling/fyne-charts/pkg/style"
)

func NewTeamTrialsPage() *fyne.Container {
	// Get data, hardcoded path for now
	home, _ := os.UserHomeDir()
	resultSet, _ := races.LoadRacesFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
	tableData := races.NewTableData(resultSet)

	table := newVetTable(tableData)

	return container.NewStack(table)
}

func newTeamTrialsChart() *fyne.Container {
	// Get data, hardcoded for now
	home, _ := os.UserHomeDir()
	resultSet, _ := races.LoadRacesFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
	raceResultArray := make([]races.RaceResult, 500)
	for _, ttr := range resultSet.Set {
		raceResultArray = append(raceResultArray, ttr.RaceResultArray...)
	}
	soa := races.NewRaceResultsSoA(raceResultArray)

	chart := coord.NewCartesianNumericalChart("Total scores per distance")
	chart.SetXAxisLabel("Samples")
	chart.SetYAxisLabel("Score")

	pal := style.NewPaletteTriadic(theme.ColorNamePrimary)
	pal = style.NewPaletteLightDarkSet(pal.Names())
	raceAveragesContainer := container.NewHBox()
	var finalScoreNps []*coord.NumericalPointSeries
	for i := range races.DistanceTypeIter() {
		raceByType := soa.FilterByDistanceType(int(i))

		// Populate chart
		finalScoreData := []gdata.NumericalPoint{}
		for i := range raceByType.TeamTotalScores.Score {
			totalScore := raceByType.TeamTotalScores.Get(i)
			point := gdata.NumericalPoint{
				N:   float64(i),
				Val: float64(totalScore),
			}
			finalScoreData = append(finalScoreData, point)
		}
		nps, err := coord.NewNumericalPointSeries(i.String(), pal.Next(), finalScoreData)
		if err != nil {
			log.Fatalf("error creating nps: %v", err)
		}
		finalScoreNps = append(finalScoreNps, nps)

		// Create text
		averageText := fmt.Sprintf("%s: %d", i.String(), raceByType.TotalScoreAverage())
		raceAveragesContainer.Add(container.NewHBox(canvas.NewText(averageText, color.Black)))
	}

	for _, nps := range finalScoreNps {
		_ = chart.AddLineSeries(nps, true)
	}

	_ = chart.SetOrigin(0, 0)

	page := container.NewBorder(nil, raceAveragesContainer, nil, nil, chart)
	return page
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
