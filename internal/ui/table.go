package ui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/cicadaclock/umango/internal/races"
)

const (
	tableCellPadding  = unit.Dp(10)
	sortArrowWidth    = unit.Dp(7)
	sortArrowHeight   = unit.Dp(4)
	sortArrowGap      = unit.Dp(5)
	selectionBarWidth = unit.Dp(2)
	// Width the column fitting measures text in, wider than any cell
	unboundedTextWidth = 1 << 20
)

// tableColumn sets how a tableView sizes, aligns and decorates one column
type tableColumn struct {
	// Text whose width the column always fits; empty fits the widest cell of the first layout
	sample string
	// Side of the square icon that tableView.layoutIcon draws centered in place of the cell text;
	// zero draws the text
	icon unit.Dp
	// Whether the column takes the width the others leave, truncating its cells with an ellipsis
	// down to the width of its header
	fill      bool
	alignment text.Alignment
}

// tableView draws a column-major string table under a locked header row
//
// Columns fit their content at the first layout and keep those widths, so filtering or refilling
// the rows never shifts them. update reports header and row clicks; the owner sorts or selects
// its data and stores the re-rendered columns and sort state back into the view.
type tableView struct {
	grid    component.GridState
	headers []string
	styles  []tableColumn
	columns [][]string
	// Fitted column widths, nil until the first layout
	widths []unit.Dp
	// Column widths of the current frame in pixels, with the fill column widened
	pixelWidths  []int
	headerClicks []gesture.Click
	rowClicks    []gesture.Click
	rowHeight    unit.Dp
	selectable   bool
	// Rows highlighted as the selection and as the comparison, -1 when none
	selectedRow, comparedRow int
	sortColumn               int
	sortState                races.SortState
	// layoutIcon draws the icon of a cell, in a column with one, into the exact constraints
	layoutIcon func(gtx layout.Context, row, col int)
}

func newTableView(
	headers []string,
	styles []tableColumn,
	rowHeight unit.Dp,
	selectable bool,
) tableView {
	return tableView{
		headers:      headers,
		styles:       styles,
		headerClicks: make([]gesture.Click, len(headers)),
		rowHeight:    rowHeight,
		selectable:   selectable,
		selectedRow:  -1,
		comparedRow:  -1,
	}
}

// setColumns replaces the table contents
func (tv *tableView) setColumns(columns [][]string) {
	tv.columns = columns
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

// update reports the clicks since the last frame
//
// Returns:
// - sortColumn: Column whose header was clicked, -1 when none was
// - clickedRow: Row that was clicked, -1 when none was
// - rowModifiers: Modifier keys held during the row click
func (tv *tableView) update(
	gtx layout.Context,
) (sortColumn, clickedRow int, rowModifiers key.Modifiers) {
	sortColumn, clickedRow = -1, -1
	for col := range tv.headerClicks {
		if ok, _ := clicked(gtx, &tv.headerClicks[col]); ok {
			sortColumn = col
		}
	}
	for row := range tv.rowClicks {
		if ok, modifiers := clicked(gtx, &tv.rowClicks[row]); ok {
			clickedRow, rowModifiers = row, modifiers
		}
	}
	return sortColumn, clickedRow, rowModifiers
}

// clicked drains the click's events and reports whether any completed a click, with the
// modifier keys held during the last one
func clicked(gtx layout.Context, click *gesture.Click) (ok bool, modifiers key.Modifiers) {
	for {
		e, more := click.Update(gtx.Source)
		if !more {
			return ok, modifiers
		}
		if e.Kind == gesture.KindClick {
			ok, modifiers = true, e.Modifiers
		}
	}
}

func (tv *tableView) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if tv.widths == nil {
		tv.fitColumns(gtx, theme)
	}
	table := component.Table(theme, &tv.grid)
	styleScrollbar(&table.VScrollbarStyle)
	styleScrollbar(&table.HScrollbarStyle)
	tv.updatePixelWidths(gtx, gtx.Constraints.Max.X-gtx.Dp(table.VScrollbarStyle.Width()))
	rowHeight := gtx.Dp(tv.rowHeight)
	dims := table.Layout(gtx, tv.rowCount(), len(tv.headers),
		func(axis layout.Axis, index, _ int) int {
			if axis == layout.Horizontal {
				return tv.pixelWidths[index]
			}
			return rowHeight
		},
		func(gtx layout.Context, col int) layout.Dimensions {
			return tv.layoutHeader(gtx, theme, col)
		},
		func(gtx layout.Context, row, col int) layout.Dimensions {
			return tv.layoutCell(gtx, theme, row, col)
		},
	)
	// component.Table draws the vertical scrollbar past the width it reports
	dims.Size.X = gtx.Constraints.Max.X
	return dims
}

// fitColumns sizes each column to fit its header beside a sort arrow, and its icon, sample or
// widest cell
func (tv *tableView) fitColumns(gtx layout.Context, theme *material.Theme) {
	gtx.Constraints = layout.Constraints{Max: image.Pt(unboundedTextWidth, gtx.Constraints.Max.Y)}
	padding := gtx.Dp(tableCellPadding)
	// Rounded the way layoutHeader rounds them, so headers never lose a pixel to truncation
	arrow := gtx.Dp(sortArrowWidth) + gtx.Dp(sortArrowGap)
	tv.widths = make([]unit.Dp, len(tv.headers))
	for col, style := range tv.styles {
		header := measureLabel(gtx, headerLabel(theme, tv.headers[col], colorText)).X + arrow
		content := 0
		switch {
		case style.fill:
		case style.icon > 0:
			content = gtx.Dp(style.icon)
		case style.sample != "":
			content = measureLabel(gtx, textLabel(theme, textSize, style.sample, colorText)).X
		default:
			for _, s := range tv.columns[col] {
				content = max(content, measureLabel(gtx, textLabel(theme, textSize, s, colorText)).X)
			}
		}
		tv.widths[col] = unit.Dp(float32(max(header, content)+2*padding) / gtx.Metric.PxPerDp)
	}
}

// updatePixelWidths converts the fitted widths to pixels, widening the fill column to take what
// the other columns leave of available
func (tv *tableView) updatePixelWidths(gtx layout.Context, available int) {
	tv.pixelWidths = tv.pixelWidths[:0]
	fill, total := -1, 0
	for col, width := range tv.widths {
		px := gtx.Dp(width)
		tv.pixelWidths = append(tv.pixelWidths, px)
		total += px
		if tv.styles[col].fill {
			fill = col
		}
	}
	if fill >= 0 && available > total {
		tv.pixelWidths[fill] += available - total
	}
}

// layoutHeader draws a header cell, with an arrow on the sort column pointing in the sort
// direction
func (tv *tableView) layoutHeader(
	gtx layout.Context,
	theme *material.Theme,
	col int,
) layout.Dimensions {
	size := gtx.Constraints.Min
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	click := &tv.headerClicks[col]
	if click.Hovered() {
		paint.Fill(gtx.Ops, colorHover)
	}
	line := gtx.Dp(borderWidth)
	fillRect(gtx, colorBorder, image.Rect(0, size.Y-line, size.X, size.Y))
	if col < len(tv.headers)-1 {
		fillRect(gtx, colorBorder, image.Rect(size.X-line, 0, size.X, size.Y))
	}
	click.Add(gtx.Ops)
	pointer.CursorPointer.Add(gtx.Ops)

	sorted := tv.sortState != races.Unsorted && col == tv.sortColumn
	c := colorTextMuted
	if sorted || click.Hovered() {
		c = colorText
	}
	padding := gtx.Dp(tableCellPadding)
	arrowWidth, arrowGap := gtx.Dp(sortArrowWidth), gtx.Dp(sortArrowGap)
	gtx.Constraints = layout.Constraints{
		Max: image.Pt(max(size.X-2*padding-arrowWidth-arrowGap, 0), size.Y),
	}
	macro := op.Record(gtx.Ops)
	dims := headerLabel(theme, tv.headers[col], c).Layout(gtx)
	call := macro.Stop()

	// The arrow leads end aligned text so the text lines up with the cells below
	var textX, arrowX int
	switch tv.styles[col].alignment {
	case text.End:
		textX = size.X - padding - dims.Size.X
		arrowX = textX - arrowGap - arrowWidth
	case text.Middle:
		textX = (size.X - dims.Size.X) / 2
		arrowX = textX + dims.Size.X + arrowGap
	default:
		textX = padding
		arrowX = textX + dims.Size.X + arrowGap
	}
	offset := op.Offset(image.Pt(textX, (size.Y-dims.Size.Y)/2)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
	if sorted {
		drawSortArrow(gtx, arrowX, size.Y/2, tv.sortState == races.Descending)
	}
	return layout.Dimensions{Size: size}
}

// layoutCell draws a body cell on its row's stripe, hover, selection or comparison background
func (tv *tableView) layoutCell(
	gtx layout.Context,
	theme *material.Theme,
	row, col int,
) layout.Dimensions {
	size := gtx.Constraints.Min
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	var highlight highlightColors
	switch row {
	case tv.selectedRow:
		highlight = selectionHighlight
	case tv.comparedRow:
		highlight = comparisonHighlight
	}
	highlighted := highlight != highlightColors{}
	switch {
	case highlighted:
		paint.Fill(gtx.Ops, highlight.fill)
	case tv.selectable && tv.rowClicks[row].Hovered():
		paint.Fill(gtx.Ops, colorHover)
	case row%2 == 1:
		paint.Fill(gtx.Ops, colorStripe)
	}
	if col < len(tv.headers)-1 {
		fillRect(gtx, colorColumnRule, image.Rect(size.X-gtx.Dp(borderWidth), 0, size.X, size.Y))
	}
	if highlighted && col == 0 {
		fillRect(gtx, highlight.accent, image.Rect(0, 0, gtx.Dp(selectionBarWidth), size.Y))
	}
	if tv.selectable {
		tv.rowClicks[row].Add(gtx.Ops)
		pointer.CursorPointer.Add(gtx.Ops)
	}

	style := tv.styles[col]
	if style.icon > 0 {
		side := gtx.Dp(style.icon)
		defer op.Offset(size.Sub(image.Pt(side, side)).Div(2)).Push(gtx.Ops).Pop()
		gtx.Constraints = layout.Exact(image.Pt(side, side))
		tv.layoutIcon(gtx, row, col)
		return layout.Dimensions{Size: size}
	}

	c := colorText
	if highlighted {
		c = colorTextBright
	}
	label := textLabel(theme, textSize, tv.columns[col][row], c)
	label.Alignment = style.alignment
	padding := gtx.Dp(tableCellPadding)
	width := max(size.X-2*padding, 0)
	gtx.Constraints = layout.Constraints{Min: image.Pt(width, 0), Max: image.Pt(width, size.Y)}
	macro := op.Record(gtx.Ops)
	dims := label.Layout(gtx)
	call := macro.Stop()
	defer op.Offset(image.Pt(padding, (size.Y-dims.Size.Y)/2)).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

func headerLabel(theme *material.Theme, s string, c color.NRGBA) material.LabelStyle {
	label := textLabel(theme, smallTextSize, s, c)
	label.Font.Weight = font.SemiBold
	return label
}

// drawSortArrow draws a small accent triangle from x rightward, centered on y, pointing up for an
// ascending sort and down for a descending one
func drawSortArrow(gtx layout.Context, x, y int, descending bool) {
	left, width := float32(x), float32(gtx.Dp(sortArrowWidth))
	height := float32(gtx.Dp(sortArrowHeight))
	tip, base := float32(y)-height/2, float32(y)+height/2
	if descending {
		tip, base = base, tip
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(left, base))
	path.LineTo(f32.Pt(left+width, base))
	path.LineTo(f32.Pt(left+width/2, tip))
	path.Close()
	paint.FillShape(gtx.Ops, colorAccent, clip.Outline{Path: path.End()}.Op())
}
