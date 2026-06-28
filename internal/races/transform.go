package races

import (
	"strconv"
)

type TableMapper interface {
	// Maps veteran card ID to chara name
	VeteranCardCharaName(ids []int) []string
}

type TableData struct {
	TrainedCharaIds []int
	Names           []string
	Distances       []string
	Styles          []string
	NumRaces        []int
	MaxScores       []int
	AvgScores       []int
}

// tableColumn represents TableData's header name and rendered data
type tableColumn struct {
	header string
	value  func(td TableData) []string
}

// Source of truth for column order and content
var tableColumns = []tableColumn{
	{"ID", func(td TableData) []string { return itoaSlice(td.TrainedCharaIds) }},
	{"Name", func(td TableData) []string { return td.Names }},
	{"Distance", func(td TableData) []string { return td.Distances }},
	{"Style", func(td TableData) []string { return td.Styles }},
	{"# Races", func(td TableData) []string { return itoaSlice(td.NumRaces) }},
	{"Max", func(td TableData) []string { return itoaSlice(td.MaxScores) }},
	{"Avg", func(td TableData) []string { return itoaSlice(td.AvgScores) }},
}

func NewTableData(dataStore TableMapper, ttrs TeamTrialResultSet) TableData {
	scores := ttrs.GetMyScores()
	umaData := ttrs.GetMyCharaData()
	distances := ttrs.GetUmaDistanceTypes()

	result := TableData{
		TrainedCharaIds: make([]int, 0, len(scores)),
		Names:           make([]string, 0, len(scores)),
		Distances:       make([]string, 0, len(scores)),
		Styles:          make([]string, 0, len(scores)),
		NumRaces:        make([]int, 0, len(scores)),
		MaxScores:       make([]int, 0, len(scores)),
		AvgScores:       make([]int, 0, len(scores)),
	}

	for trainedCharaId, scoreArray := range scores {
		uma := umaData[trainedCharaId]

		result.TrainedCharaIds = append(result.TrainedCharaIds, trainedCharaId)
		result.Names = append(result.Names, dataStore.VeteranCardCharaName([]int{uma.CardId})...)
		result.Distances = append(result.Distances, distances[uma.TrainedCharaId].String())
		result.Styles = append(result.Styles, uma.RunningStyle.String())
		result.NumRaces = append(result.NumRaces, scoreArray.Len())
		result.MaxScores = append(result.MaxScores, scoreArray.Max())
		result.AvgScores = append(result.AvgScores, scoreArray.Average())
	}
	return result
}

// Filter returns a new TableData containing only the rows at the given indices.
func (td TableData) Filter(indices []int) TableData {
	result := TableData{
		TrainedCharaIds: filterSlice(td.TrainedCharaIds, indices),
		Names:           filterSlice(td.Names, indices),
		Distances:       filterSlice(td.Distances, indices),
		NumRaces:        filterSlice(td.NumRaces, indices),
		MaxScores:       filterSlice(td.MaxScores, indices),
		AvgScores:       filterSlice(td.AvgScores, indices),
	}
	return result
}

func (td TableData) Len() int {
	return len(td.TrainedCharaIds)
}

// Header names for each column
func (td TableData) Headers() []string {
	headers := make([]string, len(tableColumns))
	for i, col := range tableColumns {
		headers[i] = col.header
	}
	return headers
}

// Columns returns the table contents in column-major order
func (td TableData) Columns() [][]string {
	cols := make([][]string, len(tableColumns))
	for i, col := range tableColumns {
		cols[i] = col.value(td)
	}
	return cols
}

// itoaSlice converts a slice of ints to a slice of strings
func itoaSlice(a []int) []string {
	result := make([]string, 0, len(a))
	for _, i := range a {
		result = append(result, strconv.Itoa(i))
	}
	return result
}

func filterSlice[T any](s []T, indices []int) []T {
	result := make([]T, 0, len(indices))
	for _, i := range indices {
		if i < 0 || i >= len(s) {
			continue
		}
		result = append(result, s[i])
	}
	return result
}
