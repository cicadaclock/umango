package races

type SkillTableMapper interface {
	// Maps Skill ID to skill name
	SkillIdName(ids []int) []string
}

// SkillTableData holds the Skills in column-major order
type SkillTableData struct {
	Ids   []int
	Names []string
	Procs []int
	Freq  []float32

	// Original ordering
	origIndices []int
}

// Source of truth for column order and content
var skillTableColumns = []Column[SkillTableData]{
	{
		"Name",
		func(td SkillTableData) []string { return td.Names },
		func(td SkillTableData, state SortState, i, j int) bool {
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
		"Procs",
		func(td SkillTableData) []string { return itoaSlice(td.Procs) },
		func(td SkillTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.Procs[i] < td.Procs[j]
			case Descending:
				return td.Procs[i] > td.Procs[j]
			}
			return false
		},
	},
	{
		"Frequency",
		func(td SkillTableData) []string { return ftoaSlice(td.Freq) },
		func(td SkillTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.Freq[i] < td.Freq[j]
			case Descending:
				return td.Freq[i] > td.Freq[j]
			}
			return false
		},
	},
}

// NewSkillTable builds the TT veteran table from a summary
func NewSkillTable(dataStore SkillTableMapper) *Table[SkillTableData] {
	return NewTable(newSkillTableData(dataStore), skillTableColumns)
}

func newSkillTableData(dataStore SkillTableMapper) SkillTableData {
	n := 10
	result := SkillTableData{
		Ids:   make([]int, 0, n),
		Names: make([]string, 0, n),
		Procs: make([]int, 0, n),
		Freq:  make([]float32, 0, n),
	}
	return result
}

// Filter returns a new SkillTableData containing only the rows at the given indices.
func (td SkillTableData) Filter(indices []int) SkillTableData {
	return SkillTableData{
		Ids:         filterSlice(td.Ids, indices),
		Names:       filterSlice(td.Names, indices),
		Procs:       filterSlice(td.Procs, indices),
		Freq:        filterSlice(td.Freq, indices),
		origIndices: filterSlice(td.origIndices, indices),
	}
}

func (td SkillTableData) Len() int {
	return len(td.Ids)
}

// Swap exchanges rows i and j across all columns
func (td SkillTableData) Swap(i, j int) {
	swapSlice(td.Ids, i, j)
	swapSlice(td.Names, i, j)
	swapSlice(td.Procs, i, j)
	swapSlice(td.Freq, i, j)
	swapSlice(td.origIndices, i, j)
}

func (td SkillTableData) OrigIndex(row int) int {
	return td.origIndices[row]
}
