package races

import (
	"strconv"
)

type TableData struct {
	TrainedCharaIds []int
	NumRaces        []int
	MaxScores       []int
	AvgScores       []int
}

func NewTableData(ttrs TeamTrialResultSet) TableData {
	scores := make(map[int]*ScoreArray)
	for _, ttr := range ttrs.Set {
		for _, charaResult := range ttr.GetMyCharaResults() {
			if len(charaResult.ScoreEventArray) == 0 {
				continue
			}
			scoreArray := scores[charaResult.TrainedCharaId]
			if scoreArray == nil {
				scoreArray = &ScoreArray{}
				scores[charaResult.TrainedCharaId] = scoreArray
			}
			scoreArray.append(charaResult.TotalScore())
		}
	}

	result := TableData{
		TrainedCharaIds: make([]int, 0, len(scores)),
		NumRaces:        make([]int, 0, len(scores)),
		MaxScores:       make([]int, 0, len(scores)),
		AvgScores:       make([]int, 0, len(scores)),
	}

	for trainedCharaId, scoreArray := range scores {
		result.TrainedCharaIds = append(result.TrainedCharaIds, trainedCharaId)
		result.NumRaces = append(result.NumRaces, scoreArray.Len())
		result.MaxScores = append(result.MaxScores, scoreArray.Max())
		result.AvgScores = append(result.AvgScores, scoreArray.Average())
	}
	return result
}

func (td TableData) Len() int {
	return len(td.TrainedCharaIds)
}

func (td TableData) Headers() []string {
	headers := []string{
		"Name",
		"# Races",
		"Max",
		"Avg",
	}
	return headers
}

// Columns returns the table contents in column-major order
func (td TableData) Columns() [][]string {
	cols := make([][]string, len(td.Headers()))
	cols[0] = make([]string, 0, td.Len())
	cols[1] = make([]string, 0, td.Len())
	cols[2] = make([]string, 0, td.Len())
	cols[3] = make([]string, 0, td.Len())
	for i := range td.TrainedCharaIds {
		cols[0] = append(cols[0], strconv.Itoa(td.TrainedCharaIds[i]))
		cols[1] = append(cols[1], strconv.Itoa(td.NumRaces[i]))
		cols[2] = append(cols[2], strconv.Itoa(td.MaxScores[i]))
		cols[3] = append(cols[3], strconv.Itoa(td.AvgScores[i]))
	}
	return cols
}

// Filter returns a new TableData containing only the rows at the given indices.
func (td TableData) Filter(indices []int) TableData {
	out := TableData{}
	for _, i := range indices {
		if i < 0 || i >= td.Len() {
			continue
		}
		out.TrainedCharaIds = append(out.TrainedCharaIds, td.TrainedCharaIds[i])
		out.NumRaces = append(out.NumRaces, td.NumRaces[i])
		out.MaxScores = append(out.MaxScores, td.MaxScores[i])
		out.AvgScores = append(out.AvgScores, td.AvgScores[i])
	}
	return out
}
