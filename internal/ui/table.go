package ui

import (
	"image"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

const (
	tableRowHeight   = unit.Dp(32)
	tableCellPadding = unit.Dp(8)
	// Column width per character of the longest cell, plus fixed padding
	tableCharWidth    = 8
	tableColumnMargin = 24
)

// tableView draws a column-major string table under a locked header row
//
// update reports header and row clicks; the owner sorts or selects its data and stores the
// re-rendered columns back into the view.
type tableView struct {
	grid         component.GridState
	headers      []string
	columns      [][]string
	widths       []unit.Dp
	headerClicks []gesture.Click
	rowClicks    []gesture.Click
	selectable   bool
	selectedRow  int
}

func newTableView(headers []string, selectable bool) tableView {
	return tableView{
		headers:      headers,
		headerClicks: make([]gesture.Click, len(headers)),
		selectable:   selectable,
		selectedRow:  -1,
	}
}

// setColumns replaces the table contents, sizing each column to its longest cell
func (tv *tableView) setColumns(columns [][]string, charWidths []int) {
	tv.columns = columns
	tv.widths = tv.widths[:0]
	for _, chars := range charWidths {
		tv.widths = append(tv.widths, unit.Dp(chars*tableCharWidth+tableColumnMargin))
	}
	if tv.selectable && len(tv.rowClicks) != tv.rowCount() {
		tv.rowClicks = make([]gesture.Click, tv.rowCount())
	}
}

func (tv *tableView) rowCount() int {
	if len(tv.columns) == 0 {
		return 0
	}
	return len(tv.columns[0])
}

func (tv *tableView) scrollToTop() {
	tv.grid.Vertical.First, tv.grid.Vertical.Offset = 0, 0
}

// update returns the column whose header was clicked and the row that was clicked, each -1
// when there was no click
func (tv *tableView) update(gtx layout.Context) (sortColumn, clickedRow int) {
	sortColumn, clickedRow = -1, -1
	for col := range tv.headerClicks {
		if clicked(gtx, &tv.headerClicks[col]) {
			sortColumn = col
		}
	}
	for row := range tv.rowClicks {
		if clicked(gtx, &tv.rowClicks[row]) {
			clickedRow = row
		}
	}
	return sortColumn, clickedRow
}

// clicked drains the click's events and reports whether any completed a click
func clicked(gtx layout.Context, click *gesture.Click) bool {
	result := false
	for {
		e, ok := click.Update(gtx.Source)
		if !ok {
			return result
		}
		if e.Kind == gesture.KindClick {
			result = true
		}
	}
}

func (tv *tableView) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	metric := gtx.Metric
	rowHeight := gtx.Dp(tableRowHeight)
	return component.Table(theme, &tv.grid).Layout(gtx, tv.rowCount(), len(tv.headers),
		func(axis layout.Axis, index, _ int) int {
			if axis == layout.Horizontal {
				return metric.Dp(tv.widths[index])
			}
			return rowHeight
		},
		func(gtx layout.Context, col int) layout.Dimensions {
			size := gtx.Constraints.Min
			defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, colorHeader)
			tv.headerClicks[col].Add(gtx.Ops)
			pointer.CursorPointer.Add(gtx.Ops)
			layoutCellText(gtx, theme, tv.headers[col])
			return layout.Dimensions{Size: size}
		},
		func(gtx layout.Context, row, col int) layout.Dimensions {
			size := gtx.Constraints.Min
			defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
			if row == tv.selectedRow {
				paint.Fill(gtx.Ops, colorSelection)
			}
			if tv.selectable {
				tv.rowClicks[row].Add(gtx.Ops)
				pointer.CursorPointer.Add(gtx.Ops)
			}
			layoutCellText(gtx, theme, tv.columns[col][row])
			return layout.Dimensions{Size: size}
		},
	)
}

// layoutCellText draws s on one line, inset from the left edge and centered vertically in the
// cell given by the exact constraints
func layoutCellText(gtx layout.Context, theme *material.Theme, s string) {
	padding := gtx.Dp(tableCellPadding)
	cellHeight := gtx.Constraints.Max.Y
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max.X = max(gtx.Constraints.Max.X-2*padding, 0)

	label := material.Body1(theme, s)
	label.MaxLines = 1
	macro := op.Record(gtx.Ops)
	dims := label.Layout(gtx)
	call := macro.Stop()
	defer op.Offset(image.Pt(padding, (cellHeight-dims.Size.Y)/2)).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
}
