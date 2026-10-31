package races

// VetTableMapper provides the chara titles that the vet table displays
type VetTableMapper interface {
	VeteranCardCharaTitle(ids []int) []string
}

// VetTableData holds the rows of the vet table, one slice per column
type VetTableData struct {
	TrainedCharaIds []int
	Names           []string
	Distances       []string
	DistanceTypes   []DistanceType
	Styles          []string
	StyleTypes      []RunStyle
	NumRaces        []int
	MaxScores       []int
	AvgScores       []int
	// Fraction of races with a full spurt
	SpurtRates []float32
	Fielded    []bool

	// Summary index order, which places the fielded team first
	origIndices []int
}

// Column indices of the vet table, in display order
const (
	VetIdColumn = iota
	VetFieldedColumn
	VetNameColumn
	VetDistanceColumn
	VetStyleColumn
	VetRacesColumn
	VetMaxColumn
	VetAvgColumn
	VetSpurtColumn
)

// vetTableColumns defines the header, cell text and sort order of each vet table column
var vetTableColumns = []Column[VetTableData]{
	VetIdColumn: {
		"ID",
		func(td VetTableData) []string { return itoaSlice(td.TrainedCharaIds) },
		func(td VetTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.TrainedCharaIds[i] < td.TrainedCharaIds[j]
			case Descending:
				return td.TrainedCharaIds[i] > td.TrainedCharaIds[j]
			}
			return false
		},
	},
	VetFieldedColumn: {
		"Fielded",
		func(td VetTableData) []string { return btoaSlice(td.Fielded) },
		func(td VetTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return compareBool(td.Fielded[j], td.Fielded[i])
			case Descending:
				return compareBool(td.Fielded[i], td.Fielded[j])
			}
			return false
		},
	},
	VetNameColumn: {
		"Name",
		func(td VetTableData) []string { return td.Names },
		func(td VetTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.Names[i] < td.Names[j]
			case Descending:
				return td.Names[i] > td.Names[j]
			}
			return false
		},
	},
	VetDistanceColumn: {
		"Distance",
		func(td VetTableData) []string { return td.Distances },
		func(td VetTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.DistanceTypes[i] < td.DistanceTypes[j]
			case Descending:
				return td.DistanceTypes[i] > td.DistanceTypes[j]
			}
			return false
		},
	},
	VetStyleColumn: {
		"Style",
		func(td VetTableData) []string { return td.Styles },
		func(td VetTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.StyleTypes[i] < td.StyleTypes[j]
			case Descending:
				return td.StyleTypes[i] > td.StyleTypes[j]
			}
			return false
		},
	},
	VetRacesColumn: {
		"# Races",
		func(td VetTableData) []string { return itoaSlice(td.NumRaces) },
		func(td VetTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.NumRaces[i] < td.NumRaces[j]
			case Descending:
				return td.NumRaces[i] > td.NumRaces[j]
			}
			return false
		},
	},
	VetMaxColumn: {
		"Max",
		func(td VetTableData) []string { return itoaSlice(td.MaxScores) },
		func(td VetTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.MaxScores[i] < td.MaxScores[j]
			case Descending:
				return td.MaxScores[i] > td.MaxScores[j]
			}
			return false
		},
	},
	VetAvgColumn: {
		"Avg",
		func(td VetTableData) []string { return itoaSlice(td.AvgScores) },
		func(td VetTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.AvgScores[i] < td.AvgScores[j]
			case Descending:
				return td.AvgScores[i] > td.AvgScores[j]
			}
			return false
		},
	},
	VetSpurtColumn: {
		"Spurt Rate",
		func(td VetTableData) []string { return percentSlice(td.SpurtRates) },
		func(td VetTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.SpurtRates[i] < td.SpurtRates[j]
			case Descending:
				return td.SpurtRates[i] > td.SpurtRates[j]
			}
			return false
		},
	},
}

// NewVetTable returns the vet table of the veterans in summary
func NewVetTable(dataStore VetTableMapper, summary TTUmaSummary) *Table[VetTableData] {
	return NewTable(newVetTableData(dataStore, summary), vetTableColumns)
}

// newVetTableData builds one row per veteran with at least one scored race
func newVetTableData(dataStore VetTableMapper, summary TTUmaSummary) VetTableData {
	n := summary.Len()
	result := VetTableData{
		TrainedCharaIds: make([]int, 0, n),
		Names:           make([]string, 0, n),
		Distances:       make([]string, 0, n),
		DistanceTypes:   make([]DistanceType, 0, n),
		Styles:          make([]string, 0, n),
		StyleTypes:      make([]RunStyle, 0, n),
		NumRaces:        make([]int, 0, n),
		MaxScores:       make([]int, 0, n),
		AvgScores:       make([]int, 0, n),
		SpurtRates:      make([]float32, 0, n),
		Fielded:         make([]bool, 0, n),
		origIndices:     make([]int, 0, n),
	}

	// The summary places the fielded team first, so the first FieldedCount indices are fielded
	for i := range n {
		scores := summary.Scores[i]
		if scores.Len() == 0 {
			continue
		}
		uma := summary.CharaData[i]

		result.TrainedCharaIds = append(result.TrainedCharaIds, summary.TrainedCharaIds[i])
		result.Names = append(result.Names, dataStore.VeteranCardCharaTitle([]int{uma.CardId})...)
		result.Distances = append(result.Distances, summary.DistanceTypes[i].String())
		result.DistanceTypes = append(result.DistanceTypes, summary.DistanceTypes[i])
		result.Styles = append(result.Styles, uma.RunningStyle.String())
		result.StyleTypes = append(result.StyleTypes, uma.RunningStyle)
		result.NumRaces = append(result.NumRaces, scores.Len())
		result.MaxScores = append(result.MaxScores, scores.Max())
		result.AvgScores = append(result.AvgScores, scores.Average())
		var spurtRate float32
		if summary.SpurtRaces[i] > 0 {
			spurtRate = float32(summary.FullSpurts[i]) / float32(summary.SpurtRaces[i])
		}
		result.SpurtRates = append(result.SpurtRates, spurtRate)
		result.Fielded = append(result.Fielded, i < summary.FieldedCount)
		result.origIndices = append(result.origIndices, i)
	}
	return result
}

// Filter returns the rows at the given indices
func (td VetTableData) Filter(indices []int) VetTableData {
	return VetTableData{
		TrainedCharaIds: filterSlice(td.TrainedCharaIds, indices),
		Names:           filterSlice(td.Names, indices),
		Distances:       filterSlice(td.Distances, indices),
		DistanceTypes:   filterSlice(td.DistanceTypes, indices),
		Styles:          filterSlice(td.Styles, indices),
		StyleTypes:      filterSlice(td.StyleTypes, indices),
		NumRaces:        filterSlice(td.NumRaces, indices),
		MaxScores:       filterSlice(td.MaxScores, indices),
		AvgScores:       filterSlice(td.AvgScores, indices),
		SpurtRates:      filterSlice(td.SpurtRates, indices),
		Fielded:         filterSlice(td.Fielded, indices),
		origIndices:     filterSlice(td.origIndices, indices),
	}
}

// Len returns the number of rows
func (td VetTableData) Len() int {
	return len(td.TrainedCharaIds)
}

// Swap exchanges rows i and j in every column
func (td VetTableData) Swap(i, j int) {
	swapSlice(td.TrainedCharaIds, i, j)
	swapSlice(td.Names, i, j)
	swapSlice(td.Distances, i, j)
	swapSlice(td.DistanceTypes, i, j)
	swapSlice(td.Styles, i, j)
	swapSlice(td.StyleTypes, i, j)
	swapSlice(td.NumRaces, i, j)
	swapSlice(td.MaxScores, i, j)
	swapSlice(td.AvgScores, i, j)
	swapSlice(td.SpurtRates, i, j)
	swapSlice(td.Fielded, i, j)
	swapSlice(td.origIndices, i, j)
}

// OrigIndex returns the position of a row before sorting
func (td VetTableData) OrigIndex(row int) int {
	return td.origIndices[row]
}

// GetTrainedCharaId returns the trained chara ID of a row
func (td VetTableData) GetTrainedCharaId(row int) int {
	return td.TrainedCharaIds[row]
}
