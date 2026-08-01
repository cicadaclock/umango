package races

type VetTableMapper interface {
	// Maps veteran card ID to chara name
	VeteranCardCharaTitle(ids []int) []string
}

// VetTableData holds the TT veteran rows in column-major order
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
	Fielded         []bool

	// Original ordering that sorts by currently used umas
	origIndices []int
}

// Source of truth for column order and content
var vetTableColumns = []Column[VetTableData]{
	{
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
	{
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
	{
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
	{
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
	{
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
	{
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
	{
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
	{
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
}

// NewVetTable builds the TT veteran table from a summary
func NewVetTable(dataStore VetTableMapper, summary TTUmaSummary) *Table[VetTableData] {
	return NewTable(newVetTableData(dataStore, summary), vetTableColumns)
}

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
		Fielded:         make([]bool, 0, n),
		origIndices:     make([]int, 0, n),
	}

	// Summary orders the current team first
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
		result.Fielded = append(result.Fielded, i < summary.FieldedCount)
		result.origIndices = append(result.origIndices, i)
	}
	return result
}

// Filter returns a new VetTableData containing only the rows at the given indices.
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
		Fielded:         filterSlice(td.Fielded, indices),
		origIndices:     filterSlice(td.origIndices, indices),
	}
}

func (td VetTableData) Len() int {
	return len(td.TrainedCharaIds)
}

// Swap exchanges rows i and j across all columns
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
	swapSlice(td.Fielded, i, j)
	swapSlice(td.origIndices, i, j)
}

func (td VetTableData) OrigIndex(row int) int {
	return td.origIndices[row]
}

// GetTrainedCharaId returns a single TrainedCharaId from a given row
func (td VetTableData) GetTrainedCharaId(row int) int {
	return td.TrainedCharaIds[row]
}
