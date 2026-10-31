package ui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

const (
	panelTabHeight  = unit.Dp(28)
	panelTabPadding = unit.Dp(12)
	panelRadius     = unit.Dp(6)
)

// panel is a docked pane: a title tab above a rounded body, outlined in the accent color while
// focused. The panel's address tags the pointer presses inside it.
type panel struct {
	title string
}

// focusPanel returns the index of the last panel pressed since the previous frame, or focus
// when none was pressed
func focusPanel(gtx layout.Context, panels []panel, focus int) int {
	for i := range panels {
		for {
			if _, ok := gtx.Event(pointer.Filter{Target: &panels[i], Kinds: pointer.Press}); !ok {
				break
			}
			focus = i
		}
	}
	return focus
}

// layout fills the constraints with the panel, with detail in muted text at the right of the
// tab strip and body inside the outline
func (pn *panel) layout(
	gtx layout.Context,
	theme *material.Theme,
	focused bool,
	detail string,
	body layout.Widget,
) layout.Dimensions {
	size := gtx.Constraints.Max
	tabHeight, padding := gtx.Dp(panelTabHeight), gtx.Dp(panelTabPadding)
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, pn)

	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: image.Pt(max(size.X-2*padding, 0), tabHeight)}
	title := textLabel(theme, smallTextSize, pn.title, colorText)
	title.Font.Weight = font.SemiBold
	macro := op.Record(gtx.Ops)
	titleDims := title.Layout(textGtx)
	titleCall := macro.Stop()
	tabWidth := titleDims.Size.X + 2*padding

	border := colorBorder
	if focused {
		border = colorAccent
	}
	drawPanelOutline(gtx, size, image.Pt(tabWidth, tabHeight), border)
	offset := op.Offset(image.Pt(padding, (tabHeight-titleDims.Size.Y)/2)).Push(gtx.Ops)
	titleCall.Add(gtx.Ops)
	offset.Pop()

	if detail != "" {
		textGtx.Constraints.Max.X = max(size.X-tabWidth-2*padding, 0)
		textGtx.Constraints.Min.X = textGtx.Constraints.Max.X
		label := textLabel(theme, smallTextSize, detail, colorTextMuted)
		label.Alignment = text.End
		macro := op.Record(gtx.Ops)
		dims := label.Layout(textGtx)
		call := macro.Stop()
		offset := op.Offset(image.Pt(tabWidth+padding, (tabHeight-dims.Size.Y)/2)).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
	}

	line := gtx.Dp(borderWidth)
	bodySize := image.Pt(max(size.X-2*line, 0), max(size.Y-tabHeight-line, 0))
	defer op.Offset(image.Pt(line, tabHeight)).Push(gtx.Ops).Pop()
	radius := max(gtx.Dp(panelRadius)-line, 0)
	defer clip.RRect{Rect: image.Rectangle{Max: bodySize}, SE: radius, SW: radius}.Push(gtx.Ops).Pop()
	gtx.Constraints = layout.Exact(bodySize)
	body(gtx)
	return layout.Dimensions{Size: size}
}

// drawPanelOutline fills the panel shape, a tab of size tab above a body that fills size, and
// strokes its outline
func drawPanelOutline(gtx layout.Context, size, tab image.Point, border color.NRGBA) {
	width := float32(gtx.Dp(borderWidth))
	radius := float32(gtx.Dp(panelRadius))
	// Stroke centers on the path, so inset it half a stroke to keep the outline inside size
	half := width / 2
	left, top := half, half
	right, bottom := float32(size.X)-half, float32(size.Y)-half
	tabRight, bodyTop := float32(tab.X)-half, float32(tab.Y)-half

	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(left, bottom-radius))
	path.LineTo(f32.Pt(left, top+radius))
	path.QuadTo(f32.Pt(left, top), f32.Pt(left+radius, top))
	path.LineTo(f32.Pt(tabRight-radius, top))
	path.QuadTo(f32.Pt(tabRight, top), f32.Pt(tabRight, top+radius))
	path.LineTo(f32.Pt(tabRight, bodyTop-radius))
	path.QuadTo(f32.Pt(tabRight, bodyTop), f32.Pt(tabRight+radius, bodyTop))
	path.LineTo(f32.Pt(right-radius, bodyTop))
	path.QuadTo(f32.Pt(right, bodyTop), f32.Pt(right, bodyTop+radius))
	path.LineTo(f32.Pt(right, bottom-radius))
	path.QuadTo(f32.Pt(right, bottom), f32.Pt(right-radius, bottom))
	path.LineTo(f32.Pt(left+radius, bottom))
	path.QuadTo(f32.Pt(left, bottom), f32.Pt(left, bottom-radius))
	path.Close()
	outline := path.End()
	paint.FillShape(gtx.Ops, colorPanel, clip.Outline{Path: outline}.Op())
	paint.FillShape(gtx.Ops, border, clip.Stroke{Path: outline, Width: width}.Op())
}
