package ui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/cicadaclock/umango/internal/races"
)

// Proportions of the game's running style icons, as fractions of the icon's diameter
const (
	runningStyleRimRatio   = 0.05
	runningStyleCapRatio   = 0.24
	runningStyleWidthRatio = 0.82
	// Cap height of Inter in ems
	interCapHeight = 0.727
)

var (
	colorRunningStyleFill  = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	colorRunningStyleLabel = color.NRGBA{R: 0x79, G: 0x3e, B: 0x14, A: 0xff}

	// Label and rim color of each running style's icon, indexed by races.RunStyle
	runningStyleIcons = [races.End + 1]struct {
		label string
		rim   color.NRGBA
	}{
		races.Front: {"FRONT", color.NRGBA{R: 0x50, G: 0x9b, B: 0xe7, A: 0xff}},
		races.Pace:  {"PACE", color.NRGBA{R: 0xa4, G: 0xdb, B: 0x21, A: 0xff}},
		races.Late:  {"LATE", color.NRGBA{R: 0xff, G: 0xde, B: 0x31, A: 0xff}},
		races.End:   {"END", color.NRGBA{R: 0xff, G: 0xb3, B: 0x36, A: 0xff}},
	}
)

// layoutRunningStyleIcon draws the game's running style icon from vector shapes and text: the
// style's name in heavy capitals on a white disc with a colored rim
//
// The disc fills the square given by the exact constraints, so the icon stays sharp at any size.
// races.NoneStyle draws nothing.
func layoutRunningStyleIcon(gtx layout.Context, theme *material.Theme, style races.RunStyle) {
	icon := runningStyleIcons[style]
	if icon.label == "" {
		return
	}
	side := gtx.Constraints.Max.X
	diameter := float32(side)
	paint.FillShape(gtx.Ops, icon.rim, clip.Ellipse{Max: image.Pt(side, side)}.Op(gtx.Ops))
	rim := max(1, int(diameter*runningStyleRimRatio+0.5))
	inner := image.Rect(rim, rim, side-rim, side-rim)
	paint.FillShape(gtx.Ops, colorRunningStyleFill, clip.Ellipse(inner).Op(gtx.Ops))

	capHeight := diameter * runningStyleCapRatio
	label := material.Label(theme, unit.Sp(capHeight/interCapHeight/gtx.Metric.PxPerSp), icon.label)
	label.Font.Weight = font.ExtraBold
	label.Color = colorRunningStyleLabel
	label.MaxLines = 1
	gtx.Constraints = layout.Constraints{Max: image.Pt(2*side, side)}
	macro := op.Record(gtx.Ops)
	dims := label.Layout(gtx)
	call := macro.Stop()

	// Condense wide labels horizontally, as the game's font does, so every label keeps the same
	// cap height
	width := float32(dims.Size.X)
	scaleX := min(1, diameter*runningStyleWidthRatio/width)
	ascent := float32(dims.Size.Y - dims.Baseline)
	origin := f32.Pt((diameter-width*scaleX)/2, (diameter+capHeight)/2-ascent)
	transform := f32.AffineId().Scale(f32.Point{}, f32.Pt(scaleX, 1)).Offset(origin)
	defer op.Affine(transform).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
}
