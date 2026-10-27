package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/cicadaclock/umango/internal/races"
)

const (
	filterIconSize = unit.Dp(24)
	filterGap      = unit.Dp(2)
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

// layout draws the buttons left to right, outlining the active filter
func (f *distanceFilter) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	gap := gtx.Dp(filterGap)
	var size image.Point
	for distance := range races.DistanceTypeIter() {
		offset := op.Offset(image.Pt(size.X, 0)).Push(gtx.Ops)
		dims := layoutToolButton(gtx, &f.buttons[distance], distance == f.selected,
			func(gtx layout.Context) layout.Dimensions {
				return f.layoutButtonContent(gtx, theme, distance)
			},
		)
		offset.Pop()
		size.X += dims.Size.X + gap
		size.Y = max(size.Y, dims.Size.Y)
	}
	size.X = max(size.X-gap, 0)
	return layout.Dimensions{Size: size}
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
	return textLabel(theme, smallTextSize, distance.String(), colorText).Layout(gtx)
}
