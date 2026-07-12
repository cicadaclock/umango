package races

import (
	"sort"
	"strconv"
)

type SortState int

const (
	Unsorted SortState = iota
	Ascending
	Descending
	sortStateSize
)

func (state SortState) Next() SortState {
	return (state + 1) % sortStateSize
}

type TableMapper interface {
	// Maps veteran card ID to chara name
	VeteranCardCharaTitle(ids []int) []string
}

type TableData struct {
	TrainedCharaIds []int
	Names           []string
	Distances       []string
	DistanceTypes   []DistanceType
	Styles          []string
	StyleTypes      []RunStyle
	NumRaces        []int
	MaxScores       []int
	AvgScores       []int
	Fielded         []bool

	// Original ordering that sorts by currently used umas
	origIndexes []int

	// Which column to sort by
	sortColumn int
	// How to sort the column
	sortState SortState
}

// tableColumn represents TableData's header name, rendered data,
// and row ordering for sorting
type tableColumn struct {
	header string
	value  func(td TableData) []string
	less   func(td TableData, i, j int) bool
}

// Source of truth for column order and content
var tableColumns = []tableColumn{
	{
		"ID",
		func(td TableData) []string { return itoaSlice(td.TrainedCharaIds) },
		func(td TableData, i, j int) bool {
			switch td.sortState {
			case Ascending:
				return td.TrainedCharaIds[i] < td.TrainedCharaIds[j]
			case Descending:
				return td.TrainedCharaIds[i] > td.TrainedCharaIds[j]
			}
			return false
		},
	},
	{
		"Fielded",
		func(td TableData) []string { return btoaSlice(td.Fielded) },
		func(td TableData, i, j int) bool {
			switch td.sortState {
			case Ascending:
				return compareBool(td.Fielded[j], td.Fielded[i])
			case Descending:
				return compareBool(td.Fielded[i], td.Fielded[j])
			}
			return false
		},
	},
	{
		"Name",
		func(td TableData) []string { return td.Names },
		func(td TableData, i, j int) bool {
			switch td.sortState {
			case Ascending:
				return td.Names[i] < td.Names[j]
			case Descending:
				return td.Names[i] > td.Names[j]
			}
			return false
		},
	},
	{
		"Distance",
		func(td TableData) []string { return td.Distances },
		func(td TableData, i, j int) bool {
			switch td.sortState {
			case Ascending:
				return td.DistanceTypes[i] < td.DistanceTypes[j]
			case Descending:
				return td.DistanceTypes[i] > td.DistanceTypes[j]
			}
			return false
		},
	},
	{
		"Style",
		func(td TableData) []string { return td.Styles },
		func(td TableData, i, j int) bool {
			switch td.sortState {
			case Ascending:
				return td.StyleTypes[i] < td.StyleTypes[j]
			case Descending:
				return td.StyleTypes[i] > td.StyleTypes[j]
			}
			return false
		},
	},
	{
		"# Races",
		func(td TableData) []string { return itoaSlice(td.NumRaces) },
		func(td TableData, i, j int) bool {
			switch td.sortState {
			case Ascending:
				return td.NumRaces[i] < td.NumRaces[j]
			case Descending:
				return td.NumRaces[i] > td.NumRaces[j]
			}
			return false
		},
	},
	{
		"Max",
		func(td TableData) []string { return itoaSlice(td.MaxScores) },
		func(td TableData, i, j int) bool {
			switch td.sortState {
			case Ascending:
				return td.MaxScores[i] < td.MaxScores[j]
			case Descending:
				return td.MaxScores[i] > td.MaxScores[j]
			}
			return false
		},
	},
	{
		"Avg",
		func(td TableData) []string { return itoaSlice(td.AvgScores) },
		func(td TableData, i, j int) bool {
			switch td.sortState {
			case Ascending:
				return td.AvgScores[i] < td.AvgScores[j]
			case Descending:
				return td.AvgScores[i] > td.AvgScores[j]
			}
			return false
		},
	},
}

func NewTableData(dataStore TableMapper, ttrs TeamTrialResultSet) TableData {
	scores := ttrs.GetMyScores()
	umaData := ttrs.GetMyCharaData()
	distances := ttrs.GetUmaDistanceTypes()

	result := TableData{
		TrainedCharaIds: make([]int, 0, len(scores)),
		Names:           make([]string, 0, len(scores)),
		Distances:       make([]string, 0, len(scores)),
		DistanceTypes:   make([]DistanceType, 0, len(scores)),
		Styles:          make([]string, 0, len(scores)),
		StyleTypes:      make([]RunStyle, 0, len(scores)),
		NumRaces:        make([]int, 0, len(scores)),
		MaxScores:       make([]int, 0, len(scores)),
		AvgScores:       make([]int, 0, len(scores)),
		Fielded:         make([]bool, 0, len(scores)),
		origIndexes:     make([]int, 0, len(scores)),
	}

	// Sort rows by latest-race order, the first 5-15 umas are the current team
	trainedCharaIds, count := ttrs.GetMyUmaOrder()
	for i, trainedCharaId := range trainedCharaIds {
		scoreArray, ok := scores[trainedCharaId]
		if !ok {
			continue
		}
		uma := umaData[trainedCharaId]

		result.TrainedCharaIds = append(result.TrainedCharaIds, trainedCharaId)
		result.Names = append(result.Names, dataStore.VeteranCardCharaTitle([]int{uma.CardId})...)
		result.Distances = append(result.Distances, distances[uma.TrainedCharaId].String())
		result.DistanceTypes = append(result.DistanceTypes, distances[uma.TrainedCharaId])
		result.Styles = append(result.Styles, uma.RunningStyle.String())
		result.StyleTypes = append(result.StyleTypes, uma.RunningStyle)
		result.NumRaces = append(result.NumRaces, scoreArray.Len())
		result.MaxScores = append(result.MaxScores, scoreArray.Max())
		result.AvgScores = append(result.AvgScores, scoreArray.Average())
		result.Fielded = append(result.Fielded, i < count)
		result.origIndexes = append(result.origIndexes, i)
	}
	return result
}

// Filter returns a new TableData containing only the rows at the given indices.
func (td TableData) Filter(indices []int) TableData {
	result := TableData{
		TrainedCharaIds: filterSlice(td.TrainedCharaIds, indices),
		Names:           filterSlice(td.Names, indices),
		Distances:       filterSlice(td.Distances, indices),
		DistanceTypes:   filterSlice(td.DistanceTypes, indices),
		Styles:          filterSlice(td.Styles, indices),
		StyleTypes:      filterSlice(td.StyleTypes, indices),
		NumRaces:        filterSlice(td.NumRaces, indices),
		MaxScores:       filterSlice(td.MaxScores, indices),
		AvgScores:       filterSlice(td.AvgScores, indices),
		Fielded:         filterSlice(td.Fielded, indices),
		origIndexes:     filterSlice(td.origIndexes, indices),
	}
	return result
}

func (td TableData) Len() int {
	return len(td.TrainedCharaIds)
}

// Swap exchanges rows i and j across all columns
func (td TableData) Swap(i, j int) {
	swapSlice(td.TrainedCharaIds, i, j)
	swapSlice(td.Names, i, j)
	swapSlice(td.Distances, i, j)
	swapSlice(td.DistanceTypes, i, j)
	swapSlice(td.Styles, i, j)
	swapSlice(td.StyleTypes, i, j)
	swapSlice(td.NumRaces, i, j)
	swapSlice(td.MaxScores, i, j)
	swapSlice(td.AvgScores, i, j)
	swapSlice(td.Fielded, i, j)
	swapSlice(td.origIndexes, i, j)
}

// Less compares the sortColumn's rows i and j
//
// Unsorted rows compare by their original indices to revert sorting
func (td TableData) Less(i, j int) bool {
	if td.sortState == Unsorted {
		return td.origIndexes[i] < td.origIndexes[j]
	}
	return tableColumns[td.sortColumn].less(td, i, j)
}

// Sort reorders all rows in place by the given column
//
// Consecutive sorts on the same column cycle through ascending, descending, and the
// original order
func (td *TableData) Sort(col int) {
	if col == td.sortColumn {
		td.sortState = td.sortState.Next()
	} else {
		td.sortState = Ascending
	}
	td.sortColumn = col
	sort.Sort(td)
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

// ColumnWidths returns the length of the longest string in a column
// including headers
func (td TableData) ColumnWidths() []int {
	lengths := make([]int, len(tableColumns))
	for i, col := range tableColumns {
		lengths[i] = len(col.header)
		for _, val := range col.value(td) {
			if len(val) > lengths[i] {
				lengths[i] = len(val)
			}
		}
	}
	return lengths
}

// GetTrainedCharaId returns a single TrainedCharaId from a given row
func (td TableData) GetTrainedCharaId(row int) int {
	return td.TrainedCharaIds[row]
}

// itoaSlice converts a slice of ints to a slice of strings
func itoaSlice(a []int) []string {
	result := make([]string, 0, len(a))
	for _, i := range a {
		result = append(result, strconv.Itoa(i))
	}
	return result
}

// btoaSlice converts a slice of bools to a slice of strings
func btoaSlice(a []bool) []string {
	result := make([]string, 0, len(a))
	for _, i := range a {
		result = append(result, strconv.FormatBool(i))
	}
	return result
}

func swapSlice[T any](s []T, i, j int) {
	s[i], s[j] = s[j], s[i]
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

// compareBool reports whether x is greater than y, where true > false
func compareBool(x, y bool) bool {
	return x && !y
}
