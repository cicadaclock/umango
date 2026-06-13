package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ColumnGrid is a simple table implementation that renders all text as
// *canvas.Text. The benefit over widget.Table is the rendering performance
// as there is no complex refresh behavior that needs to be implemented.
type ColumnGrid struct {
	widget.BaseWidget
	ShowColumnHeaders, ShowRowHeaders bool
	ColumnHeaders, RowHeaders         []string

	Columns                   []Column // Slice of columns
	ColumnPadding, RowPadding float32  // Padding between columns, rows
}

func NewColumnGrid(data [][]string) *ColumnGrid {
	columns := make([]Column, 0, len(data))
	for _, colData := range data {
		column := Column{
			Data: colData,
		}
		columns = append(columns, column)
	}
	item := &ColumnGrid{
		Columns: columns,
	}

	item.ExtendBaseWidget(item)

	return item
}

func (item *ColumnGrid) CreateRenderer() fyne.WidgetRenderer {
	columns := createColumns(item.Columns, item.ColumnPadding, item.RowPadding)
	grid := createGrid(columns, item.ColumnPadding, item.RowPadding)
	table := container.NewStack(columns, grid)
	tableScroller := container.NewScroll(table)
	return widget.NewSimpleRenderer(tableScroller)
}

// createGrid returns a *fyne.Container with no layout that contains a grid of
// lines arranged based on the *fyne.Container HBox containing VBox c.
func createGrid(c *fyne.Container, columnPadding, rowPadding float32) *fyne.Container {
	pos := c.Position()
	// Get width and height data from container
	var width, height float32
	var widths, heights []float32
	for _, vContainer := range c.Objects {
		vBox := vContainer.(*fyne.Container)
		w := vBox.Size().Width
		widths = append(widths, w)
		width += w
	}
	for _, cell := range c.Objects[0].(*fyne.Container).Objects {
		h := cell.Size().Height
		heights = append(heights, h)
		height += h
	}
	width += columnPadding * float32(len(widths))
	height += rowPadding * float32(len(heights))

	// Create lines
	lines := make([]*canvas.Line, 0, len(widths)+len(heights)+2)
	var x, y float32 = -1 * (columnPadding / 2), -1 * (rowPadding / 2)
	line := createLine(color.Black, 0.0, height, pos.X+x, pos.Y-(rowPadding/2))
	lines = append(lines, line)
	line = createLine(color.Black, width, 0.0, pos.X-(columnPadding/2), pos.Y+y)
	lines = append(lines, line)
	for _, w := range widths {
		x += w + columnPadding
		line := createLine(color.Black, 0.0, height, pos.X+x, pos.Y-(rowPadding/2))
		lines = append(lines, line)
	}
	for _, h := range heights {
		y += h + rowPadding
		line := createLine(color.Black, width, 0.0, pos.X-(columnPadding/2), pos.Y+y)
		lines = append(lines, line)
	}

	// Return as container
	grid := container.NewWithoutLayout()
	for _, line := range lines {
		grid.Add(line)
	}
	return grid
}

// Returns a line with the given width, height and moves it to x, y
func createLine(color color.Color, width, height, x, y float32) *canvas.Line {
	line := canvas.NewLine(color)
	line.Resize(fyne.NewSize(width, height))
	line.Move(fyne.NewPos(x, y))
	return line
}

// Returns an HBox with a column for each slice of data
func createColumns(data []Column, columnPadding, rowPadding float32) *fyne.Container {
	columns := container.NewHBox()
	for _, column := range data {
		column := createColumnContainer(column, rowPadding)
		columns.Add(column)
	}
	columns.Layout = layout.NewCustomPaddedHBoxLayout(columnPadding)
	return columns
}

// Returns a VBox with a row for each given string
func createColumnContainer(column Column, rowPadding float32) *fyne.Container {
	columnContainer := container.NewVBox()
	for _, row := range column.Data {
		cell := canvas.NewText(row, color.Black)
		columnContainer.Add(cell)
	}
	columnContainer.Layout = layout.NewCustomPaddedVBoxLayout(rowPadding)
	return columnContainer
}

type Column struct {
	Data []string
}
