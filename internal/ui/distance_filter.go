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

const filterGap = unit.Dp(6)

// distanceFilter is a row of toggle pills, one per distance type and labeled with its name, that
// select at most one distance type to filter by
type distanceFilter struct {
	// Indexed by races.DistanceType
	buttons [races.Dirt + 1]widget.Clickable
	// races.NoneDistance when no filter is active
	selected races.DistanceType
}

// update applies the clicks since the last frame: clicking a pill activates its distance type
// and deactivates the others, and clicking the active pill clears the filter
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

// layout draws the pills left to right, highlighting the active filter
func (f *distanceFilter) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	gap := gtx.Dp(filterGap)
	var size image.Point
	for distance := range races.DistanceTypeIter() {
		offset := op.Offset(image.Pt(size.X, 0)).Push(gtx.Ops)
		dims := layoutPill(gtx, theme, &f.buttons[distance], distance.String(), distance == f.selected)
		offset.Pop()
		size.X += dims.Size.X + gap
		size.Y = max(size.Y, dims.Size.Y)
	}
	size.X = max(size.X-gap, 0)
	return layout.Dimensions{Size: size}
}
