package ui

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/cicadaclock/umango/internal/races"
)

const (
	filterIconSize = unit.Dp(32)
	// Gap between a button's edge and its content, and between neighboring buttons
	filterPadding = unit.Dp(4)
	filterRadius  = unit.Dp(6)
	filterBorder  = unit.Dp(2)
)

// distanceFilter is a bar of toggle buttons, one per distance type, that select at most one
// distance type to filter by
type distanceFilter struct {
	icons distanceTypeIcons
	// Indexed by races.DistanceType
	buttons [races.Dirt + 1]widget.Clickable
	// races.NoneDistance when no filter is active
	selected races.DistanceType
}

// update applies the clicks since the last frame: clicking a button activates its distance type
// and deactivates the others, and clicking the active button clears the filter
//
// Returns:
// - changed: Whether the selected distance type changed
func (f *distanceFilter) update(gtx layout.Context) (changed bool) {
	for distance := range races.DistanceTypeIter() {
		if !f.buttons[distance].Clicked(gtx) {
			continue
		}
		if f.selected == distance {
			f.selected = races.NoneDistance
		} else {
			f.selected = distance
		}
		changed = true
	}
	return changed
}

// layout draws the buttons left to right
func (f *distanceFilter) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	gap := gtx.Dp(filterPadding)
	size := image.Pt(gap, 0)
	for distance := range races.DistanceTypeIter() {
		offset := op.Offset(image.Pt(size.X, gap)).Push(gtx.Ops)
		dims := f.layoutButton(gtx, theme, distance)
		offset.Pop()
		size.X += dims.Size.X + gap
		size.Y = max(size.Y, dims.Size.Y)
	}
	size.Y += 2 * gap
	return layout.Dimensions{Size: size}
}

// layoutButton draws distance's button on a rounded background that marks the active filter
// and the hovered button
func (f *distanceFilter) layoutButton(
	gtx layout.Context,
	theme *material.Theme,
	distance races.DistanceType,
) layout.Dimensions {
	button := &f.buttons[distance]
	return button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		padding := gtx.Dp(filterPadding)
		macro := op.Record(gtx.Ops)
		offset := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
		content := f.layoutButtonContent(gtx, theme, distance)
		offset.Pop()
		call := macro.Stop()

		bounds := image.Rectangle{Max: content.Size.Add(image.Pt(2*padding, 2*padding))}
		radius := gtx.Dp(filterRadius)
		switch {
		case distance == f.selected:
			paint.FillShape(gtx.Ops, colorSelection, clip.UniformRRect(bounds, radius).Op(gtx.Ops))
			border := gtx.Dp(filterBorder)
			outline := clip.UniformRRect(bounds.Inset(border/2), radius).Path(gtx.Ops)
			paint.FillShape(gtx.Ops, colorPrimary, clip.Stroke{Path: outline, Width: float32(border)}.Op())
		case button.Hovered():
			paint.FillShape(gtx.Ops, colorHover, clip.UniformRRect(bounds, radius).Op(gtx.Ops))
		}
		call.Add(gtx.Ops)
		defer clip.Rect(bounds).Push(gtx.Ops).Pop()
		pointer.CursorPointer.Add(gtx.Ops)
		return layout.Dimensions{Size: bounds.Max}
	})
}

// layoutButtonContent draws distance's racing shoe icon, or its name when the icon was not
// extracted
func (f *distanceFilter) layoutButtonContent(
	gtx layout.Context,
	theme *material.Theme,
	distance races.DistanceType,
) layout.Dimensions {
	if icon := f.icons[distance]; iconLoaded(icon) {
		side := gtx.Dp(filterIconSize)
		gtx.Constraints = layout.Exact(image.Pt(side, side))
		return widget.Image{Src: icon, Fit: widget.Contain}.Layout(gtx)
	}
	gtx.Constraints.Min = image.Point{}
	label := material.Body1(theme, distance.String())
	label.MaxLines = 1
	return label.Layout(gtx)
}
