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
	"github.com/cicadaclock/umango/internal/races"
	"github.com/s-daehling/fyne-charts/pkg/coord"
	gdata "github.com/s-daehling/fyne-charts/pkg/data"
	"github.com/s-daehling/fyne-charts/pkg/style"
)

type DistanceType int

const (
	_                   = iota
	Sprint DistanceType = iota
	Mile
	Medium
	Long
	Dirt
)

var distances = []DistanceType{Sprint, Mile, Medium, Long, Dirt}

type RunStyle int

const (
	_              = iota
	Front RunStyle = iota
	Pace
	Late
	End
)

var runStyles = []RunStyle{Front, Pace, Late, End}

func TeamTrialsChart() *fyne.Container {
	// Get data, hardcoded for now
	home, _ := os.UserHomeDir()
	results, _ := races.LoadRaceResultsFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
	soa := races.NewRaceResultsSoA(results)

	chart := coord.NewCartesianNumericalChart("Total scores per distance")
	chart.SetXAxisLabel("Samples")
	chart.SetYAxisLabel("Score")

	pal := style.NewPaletteTriadic(theme.ColorNamePrimary)
	pal = style.NewPaletteLightDarkSet(pal.Names())
	raceAveragesContainer := container.NewHBox()
	var finalScoreNps []*coord.NumericalPointSeries
	for _, i := range distances {
		raceByType := soa.FilterByDistanceType(int(i))

		// Populate chart
		finalScoreData := []gdata.NumericalPoint{}
		for i := range raceByType.TeamTotalScores {
			totalScore := raceByType.TeamTotalScores[i]
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

func (dt DistanceType) String() string {
	switch dt {
	case Sprint:
		return "Sprint"
	case Mile:
		return "Mile"
	case Medium:
		return "Medium"
	case Long:
		return "Long"
	case Dirt:
		return "Dirt"
	}
	return ""
}

func (rs RunStyle) String() string {
	switch rs {
	case Front:
		return "Front"
	case Pace:
		return "Pace"
	case Late:
		return "Late"
	case End:
		return "End"
	}
	return ""
}
