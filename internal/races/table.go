package races

import (
	"sort"
	"strconv"
)

// SortState is the sort direction of a table column
type SortState int

// Sort states in the order that repeated header clicks select them
const (
	Unsorted SortState = iota
	Ascending
	Descending
	// Number of sort states
	sortStateSize
)

// Next returns the sort state that the next header click selects
func (state SortState) Next() SortState {
	return (state + 1) % sortStateSize
}

// TableSource is the row storage of a Table; T is the implementing type
type TableSource[T any] interface {
	// Number of rows
	Len() int
	// Exchanges rows i and j in every column
	Swap(i, j int)
	// Position of a row before sorting
	OrigIndex(row int) int
	// Copy holding only the rows at the given indices
	Filter(indices []int) T
}

// Column defines the header, cell text and sort order of one table column
type Column[T TableSource[T]] struct {
	Header string
	Value  func(data T) []string
	Less   func(data T, state SortState, i, j int) bool
}

// Table sorts and renders the rows of a TableSource through its columns
type Table[T TableSource[T]] struct {
	data    T
	columns []Column[T]

	// Column and direction of the current sort
	sortColumn int
	sortState  SortState
}

// NewTable returns an unsorted table of data
func NewTable[T TableSource[T]](data T, columns []Column[T]) *Table[T] {
	return &Table[T]{data: data, columns: columns}
}

// Data returns the rows in their current order
func (t *Table[T]) Data() T {
	return t.data
}

// Filter returns a new unsorted table of the rows at the given indices
func (t *Table[T]) Filter(indices []int) *Table[T] {
	return NewTable(t.data.Filter(indices), t.columns)
}

// Len, Swap and Less implement sort.Interface over the rows
func (t *Table[T]) Len() int {
	return t.data.Len()
}

func (t *Table[T]) Swap(i, j int) {
	t.data.Swap(i, j)
}

// Less orders rows by the sort column, or by their original position when unsorted
func (t *Table[T]) Less(i, j int) bool {
	if t.sortState == Unsorted {
		return t.data.OrigIndex(i) < t.data.OrigIndex(j)
	}
	return t.columns[t.sortColumn].Less(t.data, t.sortState, i, j)
}

// Sort sorts the rows by a column. Repeated sorts on the same column cycle through ascending,
// descending and unsorted.
func (t *Table[T]) Sort(col int) {
	state := Ascending
	if col == t.sortColumn {
		state = t.sortState.Next()
	}
	t.SortBy(col, state)
}

// SortBy sorts the rows by a column in the given direction; Unsorted restores the original order
func (t *Table[T]) SortBy(col int, state SortState) {
	t.sortColumn, t.sortState = col, state
	sort.Sort(t)
}

// SortedColumn returns the sort column and the sort direction
func (t *Table[T]) SortedColumn() (col int, state SortState) {
	return t.sortColumn, t.sortState
}

// Headers returns the header of each column
func (t *Table[T]) Headers() []string {
	headers := make([]string, len(t.columns))
	for i, col := range t.columns {
		headers[i] = col.Header
	}
	return headers
}

// Columns returns the cell text of every row, one slice per column
func (t *Table[T]) Columns() [][]string {
	cols := make([][]string, len(t.columns))
	for i, col := range t.columns {
		cols[i] = col.Value(t.data)
	}
	return cols
}

// itoaSlice formats each int as decimal text
func itoaSlice(a []int) []string {
	result := make([]string, 0, len(a))
	for _, i := range a {
		result = append(result, strconv.Itoa(i))
	}
	return result
}

// ftoaSlice formats each float with two decimals
func ftoaSlice(a []float32) []string {
	result := make([]string, 0, len(a))
	for _, i := range a {
		result = append(result, strconv.FormatFloat(float64(i), 'f', 2, 32))
	}
	return result
}

// percentSlice formats each fraction as a percentage with one decimal
func percentSlice(a []float32) []string {
	result := make([]string, 0, len(a))
	for _, f := range a {
		result = append(result, strconv.FormatFloat(float64(f)*100, 'f', 1, 32)+"%")
	}
	return result
}

// btoaSlice formats each bool as true or false
func btoaSlice(a []bool) []string {
	result := make([]string, 0, len(a))
	for _, i := range a {
		result = append(result, strconv.FormatBool(i))
	}
	return result
}

// swapSlice exchanges elements i and j
func swapSlice[T any](s []T, i, j int) {
	s[i], s[j] = s[j], s[i]
}

// filterSlice returns the elements at the given indices, skipping indices out of range
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

// compareBool reports whether x orders after y, with true after false
func compareBool(x, y bool) bool {
	return x && !y
}
