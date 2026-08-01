package races

import "slices"

type SkillTableMapper interface {
	// Maps Skill ID to skill name
	SkillNames(ids []int) []string
}

// SkillTableData holds the Skills in column-major order
type SkillTableData struct {
	Ids   []int
	Names []string
	Procs []int
	Rate  []float32

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
		"Rate",
		func(td SkillTableData) []string { return ftoaSlice(td.Rate) },
		func(td SkillTableData, state SortState, i, j int) bool {
			switch state {
			case Ascending:
				return td.Rate[i] < td.Rate[j]
			case Descending:
				return td.Rate[i] > td.Rate[j]
			}
			return false
		},
	},
}

// NewSkillTable builds an empty Skill table
//
// The table has no rows until you call RefillSkillTable.
func NewSkillTable() *Table[SkillTableData] {
	return NewTable(SkillTableData{}, skillTableColumns)
}

// RefillSkillTable replaces the rows of t with the skill data of one uma
func RefillSkillTable(t *Table[SkillTableData], dataStore SkillTableMapper, procs map[int]int, size int) {
	td := t.data
	n := len(procs)
	td.Ids = slices.Grow(td.Ids[:0], n)
	td.Names = td.Names[:0]
	td.Procs = slices.Grow(td.Procs[:0], n)
	td.Rate = slices.Grow(td.Rate[:0], n)
	td.origIndices = slices.Grow(td.origIndices[:0], n)

	// Sort by ID
	for skillId := range procs {
		td.Ids = append(td.Ids, skillId)
	}
	slices.SortFunc(td.Ids, func(a, b int) int {
		if d := procs[b] - procs[a]; d != 0 {
			return d
		}
		return a - b
	})

	var rate float32
	if size > 0 {
		rate = 1 / float32(size)
	}
	for i, skillId := range td.Ids {
		td.Procs = append(td.Procs, procs[skillId])
		td.Rate = append(td.Rate, float32(procs[skillId])*rate)
		td.origIndices = append(td.origIndices, i)
	}
	td.Names = append(td.Names, dataStore.SkillNames(td.Ids)...)

	t.data = td
	t.sortColumn = 0
	t.sortState = Unsorted
}

// Filter returns a new SkillTableData containing only the rows at the given indices.
func (td SkillTableData) Filter(indices []int) SkillTableData {
	return SkillTableData{
		Ids:         filterSlice(td.Ids, indices),
		Names:       filterSlice(td.Names, indices),
		Procs:       filterSlice(td.Procs, indices),
		Rate:        filterSlice(td.Rate, indices),
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
	swapSlice(td.Rate, i, j)
	swapSlice(td.origIndices, i, j)
}

func (td SkillTableData) OrigIndex(row int) int {
	return td.origIndices[row]
}
