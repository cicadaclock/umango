package pages

import (
	"math"
	"os"
	"path/filepath"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/s-daehling/fyne-charts/pkg/coord"
	gdata "github.com/s-daehling/fyne-charts/pkg/data"
)

const (
	// Size modifier for histogram to prevent flickering between bars when resizing
	BAR_WIDTH_MODIFIER = 1.05
	BAR_WIDTH          = 5000
)

func NewTeamTrialsPage(dataStore *data.DataStore) *fyne.Container {
	// Get data, hardcoded path for now
	home, err := os.UserHomeDir()
	if err != nil {
		return container.NewWithoutLayout()
	}
	resultSet, err := races.LoadRacesFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
	if err != nil {
		return container.NewWithoutLayout()
	}

	// Individual score histograms
	scores := resultSet.GetMyScores()
	maxScore, maxFreq := 0, 0
	umaScoreData := make(map[int]*coord.NumericalPointSeries, len(scores))
	for trainedCharaId, scoreArray := range scores {
		nps, freq := calculateScoreData(*scoreArray, BAR_WIDTH)
		umaScoreData[trainedCharaId] = nps
		// Max score for histogram range
		maxScore = int(math.Max(float64(maxScore), float64(scoreArray.Max())))
		maxFreq = int(math.Max(float64(maxFreq), float64(freq)))
	}
	chart := newScoreHistogram(maxScore, maxFreq)

	// Skill table
	skillTable := container.NewWithoutLayout()

	// TT veteran table
	tableData := races.NewTableData(dataStore, resultSet)
	headers := tableData.Headers()
	cols := tableData.Columns()
	table := newVetTable(headers, cols, tableData.ColumnWidths())
	// Select row
	table.OnSelected = func(id widget.TableCellID) {
		i := tableData.GetTrainedCharaId(id.Row)
		swapHistogram(chart, umaScoreData[i], float64(BAR_WIDTH)*BAR_WIDTH_MODIFIER)
	}
	// Sort on header click
	table.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		b := cell.(*widget.Button)
		// Set the button only when we need to
		if b.Text != headers[id.Col] {
			b.SetText(headers[id.Col])
			b.OnTapped = func() {
				table.UnselectAll()
				tableData.Sort(id.Col)
				for i, col := range tableData.Columns() {
					copy(cols[i], col)
				}
				table.Refresh()
			}
		}
	}
	// Filter buttons for TT veteran table
	filters := container.NewWithoutLayout()

	// Page containers
	leftSide := container.NewBorder(filters, nil, nil, nil, table)
	rightSide := container.NewVSplit(chart, skillTable)
	rightSide.SetOffset(0.45)

	split := container.NewHSplit(leftSide, rightSide)
	split.SetOffset(0.7)
	return container.NewStack(split)
}

// Replaces a chart's point series with new ones
func swapHistogram(chart *coord.CartesianNumericalChart, nps *coord.NumericalPointSeries, barWidth float64) {
	chart.RemoveSeries("data")
	chart.AddBarSeries(nps, barWidth)
}

func newScoreHistogram(maxScore, maxFreq int) *coord.CartesianNumericalChart {
	// Labels
	chart := coord.NewCartesianNumericalChart("Score vs. Frequency")
	chart.SetXAxisLabel("Score")
	chart.SetYAxisLabel("Frequency")
	chart.HideLegend()
	_ = chart.SetOrigin(0.0, 0.0)
	_ = chart.SetXRange(0.0, float64(maxScore+10000))
	_ = chart.SetYRange(0.0, float64(maxFreq)+1)
	return chart
}

// calculateScoreData transforms ScoreArray Scores into histogram coordinates,
// returning the series and the tallest bucket's frequency
func calculateScoreData(scoreArray races.ScoreArray, stepSize int) (*coord.NumericalPointSeries, int) {
	finalScoreData := []gdata.NumericalPoint{}
	xPts, yPts := scoreArray.HistogramCoords(stepSize)
	for i := range xPts {
		point := gdata.NumericalPoint{
			N:   float64(xPts[i]),
			Val: float64(yPts[i]),
		}
		finalScoreData = append(finalScoreData, point)
	}
	// Not polar data so this will never error
	nps, _ := coord.NewNumericalPointSeries("data", theme.ColorNamePrimary, finalScoreData)
	return nps, slices.Max(yPts)
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

	// Headers
	table.ShowHeaderRow = true
	table.CreateHeader = func() fyne.CanvasObject {
		b := widget.NewButton("temp", func() {})
		return b
	}
	for col, length := range colWidths {
		table.SetColumnWidth(col, (float32(length)*7)+24)
	}

	return table
}
