package pages

import (
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

// TeamTrialsData holds everything the team trials page needs to render
type TeamTrialsData struct {
	Summary    races.TTUmaSummary
	VetTable   *races.Table[races.VetTableData]
	SkillTable *races.Table[races.SkillTableData]

	// Data needed for SkillTable, todo refactor
	// dataStore maps the skill IDs to skill names at each refresh
	dataStore *data.DataStore
	// skillCols holds the rendered skill table in column-major order
	skillCols [][]string
	// skillWidget shows skillCols. It's nil until the page is built
	skillWidget *widget.Table
}

// LoadTeamTrialResults loads the saved team trial races from a hardcoded
// directory
func LoadTeamTrialResults() (races.TeamTrialResultSet, error) {
	// Hardcoded path for now
	home, err := os.UserHomeDir()
	if err != nil {
		return races.TeamTrialResultSet{}, err
	}
	return races.LoadRacesFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
}

// NewTeamTrialsData aggregates loaded races into the page's inputs
func NewTeamTrialsData(dataStore *data.DataStore, resultSet races.TeamTrialResultSet) TeamTrialsData {
	summary := resultSet.Summarize()
	skillTable := races.NewSkillTable()
	return TeamTrialsData{
		Summary:    summary,
		VetTable:   races.NewVetTable(dataStore, summary),
		SkillTable: skillTable,
		dataStore:  dataStore,
		skillCols:  make([][]string, len(skillTable.Headers())),
	}
}

// NewTeamTrialsPage builds the page widgets and must be called on the UI thread.
func NewTeamTrialsPage(ttData TeamTrialsData) *fyne.Container {
	// Individual score histograms
	maxScore, maxFreq := 0, 0
	umaScoreData := make(map[int]*coord.NumericalPointSeries, ttData.Summary.Len())
	for i, scoreArray := range ttData.Summary.Scores {
		if scoreArray.Len() == 0 {
			continue
		}
		nps, freq := calculateScoreData(scoreArray, BAR_WIDTH)
		umaScoreData[ttData.Summary.TrainedCharaIds[i]] = nps
		// Max score for histogram range
		maxScore = max(maxScore, scoreArray.Max())
		maxFreq = max(maxFreq, freq)
	}
	chart := newScoreHistogram(maxScore, maxFreq)

	// Skill table. It stays empty until a vet table row is selected.
	skillHeaders := ttData.SkillTable.Headers()
	skillTable := newTable(skillHeaders, ttData.skillCols, ttData.SkillTable.ColumnWidths())
	ttData.skillWidget = skillTable
	// Sort on header click
	skillTable.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		b := cell.(*widget.Button)
		// Set the button only when we need to
		if b.Text != skillHeaders[id.Col] {
			b.SetText(skillHeaders[id.Col])
			b.OnTapped = func() {
				skillTable.UnselectAll()
				ttData.SkillTable.Sort(id.Col)
				// A sort keeps the row count. Thus copy into the buffers.
				for i, col := range ttData.SkillTable.Columns() {
					copy(ttData.skillCols[i], col)
				}
				skillTable.Refresh()
			}
		}
	}

	// TT veteran table
	vetTable := ttData.VetTable
	headers := vetTable.Headers()
	cols := vetTable.Columns()
	table := newTable(headers, cols, vetTable.ColumnWidths())
	// Select row
	table.OnSelected = func(id widget.TableCellID) {
		i := vetTable.Data().GetTrainedCharaId(id.Row)
		swapHistogram(chart, umaScoreData[i], float64(BAR_WIDTH)*BAR_WIDTH_MODIFIER)
		ttData.RefreshSkillTable(i)
	}
	// Sort on header click
	table.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		b := cell.(*widget.Button)
		// Set the button only when we need to
		if b.Text != headers[id.Col] {
			b.SetText(headers[id.Col])
			b.OnTapped = func() {
				table.UnselectAll()
				vetTable.Sort(id.Col)
				for i, col := range vetTable.Columns() {
					copy(cols[i], col)
				}
				table.Refresh()
			}
		}
	}
	table.Select(widget.TableCellID{Row: 0, Col: 0})
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

// RefreshSkillTable fills the skill table with the skill data of the
// uma with the given trained chara ID
func (ttData TeamTrialsData) RefreshSkillTable(trainedCharaId int) {
	i := slices.Index(ttData.Summary.TrainedCharaIds, trainedCharaId)
	if i < 0 {
		return
	}
	races.RefillSkillTable(
		ttData.SkillTable,
		ttData.dataStore,
		ttData.Summary.SkillActivations[i],
		ttData.Summary.Scores[i].Len(),
	)

	for c, col := range ttData.SkillTable.Columns() {
		ttData.skillCols[c] = append(ttData.skillCols[c][:0], col...)
	}
	if ttData.skillWidget == nil {
		return
	}
	ttData.skillWidget.UnselectAll()
	for c, length := range ttData.SkillTable.ColumnWidths() {
		ttData.skillWidget.SetColumnWidth(c, (float32(length)*7)+24)
	}
	ttData.skillWidget.ScrollToTop()
	ttData.skillWidget.Refresh()
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

// newTable creates a table with the given headers and columns
func newTable(headers []string, cols [][]string, colWidths []int) *widget.Table {
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
