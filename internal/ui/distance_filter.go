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

// Space between adjacent pills
const filterGap = unit.Dp(6)

// distanceFilter is a row of pills, one per distance category, that selects at most one
// category to filter the vet table by
type distanceFilter struct {
	// Indexed by races.DistanceType
	buttons [races.Dirt + 1]widget.Clickable
	// races.NoneDistance when no filter is active
	selected races.DistanceType
}

// update applies the pill clicks since the last frame and reports whether the selection
// changed. Clicking the active pill clears the filter.
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

// layout draws the pills from left to right, highlighting the active pill
func (f *distanceFilter) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	gap := gtx.Dp(filterGap)
	var size image.Point
	for distance := range races.DistanceTypeIter() {
		var highlight highlightColors
		if distance == f.selected {
			highlight = selectionHighlight
		}
		offset := op.Offset(image.Pt(size.X, 0)).Push(gtx.Ops)
		dims := layoutPill(gtx, theme, &f.buttons[distance], distance.String(), highlight)
		offset.Pop()
		size.X += dims.Size.X + gap
		size.Y = max(size.Y, dims.Size.Y)
	}
	size.X = max(size.X-gap, 0)
	return layout.Dimensions{Size: size}
}
