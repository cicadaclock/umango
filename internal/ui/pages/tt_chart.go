package pages

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/s-daehling/fyne-charts/pkg/coord"
	gdata "github.com/s-daehling/fyne-charts/pkg/data"
	"github.com/s-daehling/fyne-charts/pkg/style"
)

func TeamTrialsChart() *fyne.Container {
	// Get data, hardcoded for now
	home, _ := os.UserHomeDir()
	results, _ := races.LoadRaceResultsFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
	soa := races.NewRaceResultsSoA(results)
	// charaResults := soa.CharaResultSoA()
	// for trainedCharaId, resultSoA := range charaResults {
	// }

	chart := coord.NewCartesianNumericalChart("Title of Example Chart")
	chart.SetXAxisLabel("Samples")
	chart.SetYAxisLabel("Score")

	pal := style.NewPaletteTriadic(theme.ColorNamePrimary)
	pal = style.NewPaletteLightDarkSet(pal.Names())
	for i := 1; i <= 5; i++ {
		raceByType := soa.FilterByDistanceType(i)
		numData := []gdata.NumericalPoint{}
		for i, totalScore := range raceByType.TeamTotalScores {
			point := gdata.NumericalPoint{
				N:   float64(i),
				Val: float64(totalScore),
			}
			numData = append(numData, point)
		}
		nps, err := coord.NewNumericalPointSeries(string(rune(i)), pal.Next(), numData)
		if err != nil {
			log.Fatalf("fuck")
		}
		_ = chart.AddLineSeries(nps, true)
	}

	return container.NewBorder(nil, nil, nil, nil, chart)
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
