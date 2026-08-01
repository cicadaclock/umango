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

// TableSource is the row storage backing a Table
//
// T is the implementing type itself
type TableSource[T any] interface {
	Len() int
	// Swap exchanges rows i and j across all columns
	Swap(i, j int)
	// OrigIndex returns a row's position before any sorting
	OrigIndex(row int) int
	// Filter returns a copy holding only the rows at the given indices
	Filter(indices []int) T
}

// Column contains the header name, rendered data, and row ordering
type Column[T TableSource[T]] struct {
	Header string
	Value  func(data T) []string
	Less   func(data T, state SortState, i, j int) bool
}

// Table renders and sorts a TableSource through its columns
type Table[T TableSource[T]] struct {
	data    T
	columns []Column[T]

	// Which column to sort by
	sortColumn int
	// How to sort the column
	sortState SortState
}

func NewTable[T TableSource[T]](data T, columns []Column[T]) *Table[T] {
	return &Table[T]{data: data, columns: columns}
}

// Data returns the backing rows in their current order
func (t *Table[T]) Data() T {
	return t.data
}

// Filter returns a new unsorted table over the rows at the given indices
func (t *Table[T]) Filter(indices []int) *Table[T] {
	return NewTable(t.data.Filter(indices), t.columns)
}

func (t *Table[T]) Len() int {
	return t.data.Len()
}

func (t *Table[T]) Swap(i, j int) {
	t.data.Swap(i, j)
}

// Less compares the sortColumn's rows i and j
//
// Unsorted rows compare by their original indices to revert sorting
func (t *Table[T]) Less(i, j int) bool {
	if t.sortState == Unsorted {
		return t.data.OrigIndex(i) < t.data.OrigIndex(j)
	}
	return t.columns[t.sortColumn].Less(t.data, t.sortState, i, j)
}

// Sort reorders all rows in place by the given column
//
// Consecutive sorts on the same column cycle through ascending, descending, and the
// original order
func (t *Table[T]) Sort(col int) {
	if col == t.sortColumn {
		t.sortState = t.sortState.Next()
	} else {
		t.sortState = Ascending
	}
	t.sortColumn = col
	sort.Sort(t)
}

// Headers returns the header name of each column
func (t *Table[T]) Headers() []string {
	headers := make([]string, len(t.columns))
	for i, col := range t.columns {
		headers[i] = col.Header
	}
	return headers
}

// Columns returns the table contents in column-major order
func (t *Table[T]) Columns() [][]string {
	cols := make([][]string, len(t.columns))
	for i, col := range t.columns {
		cols[i] = col.Value(t.data)
	}
	return cols
}

// ColumnWidths returns the length of the longest string in a column
// including headers
func (t *Table[T]) ColumnWidths() []int {
	lengths := make([]int, len(t.columns))
	for i, col := range t.columns {
		lengths[i] = len(col.Header)
		for _, val := range col.Value(t.data) {
			if len(val) > lengths[i] {
				lengths[i] = len(val)
			}
		}
	}
	return lengths
}

// itoaSlice converts a slice of ints to a slice of strings
func itoaSlice(a []int) []string {
	result := make([]string, 0, len(a))
	for _, i := range a {
		result = append(result, strconv.Itoa(i))
	}
	return result
}

// ftoaSlice converts a slice of floats to a slice of strings
func ftoaSlice(a []float32) []string {
	result := make([]string, 0, len(a))
	for _, i := range a {
		result = append(result, strconv.FormatFloat(float64(i), 'f', 2, 32))
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
