package ui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Text sizes and shared dimensions of the dark theme
const (
	textSize        = unit.Sp(13)
	smallTextSize   = unit.Sp(12)
	captionTextSize = unit.Sp(11)
	// Space between panels and around the window edge
	gutterWidth   = unit.Dp(6)
	borderWidth   = unit.Dp(1)
	controlRadius = unit.Dp(4)
	// Space between the edge of a pill and its label
	pillPaddingX   = unit.Dp(10)
	pillPaddingY   = unit.Dp(4)
	scrollbarWidth = unit.Dp(6)
)

// Dark palette with one orange accent for focus, selection and data. Tints are opaque, because
// Gio blends translucent colors in linear light, which brightens them over dark panels.
var (
	colorWindow       = rgb(0x0f0f0f)
	colorPanel        = rgb(0x161616)
	colorStripe       = rgb(0x1b1b1b)
	colorHover        = rgb(0x262626)
	colorPressed      = rgb(0x303030)
	colorBorder       = rgb(0x2a2a2a)
	colorBorderStrong = rgb(0x3a3a3a)
	colorColumnRule   = rgb(0x101010)
	colorGridLine     = rgb(0x202020)
	colorTooltip      = rgb(0x202020)
	colorScrollbar    = rgb(0x3a3a3a)
	colorScrollHover  = rgb(0x555555)
	colorText         = rgb(0xe6e6e6)
	colorTextBright   = rgb(0xffffff)
	colorTextMuted    = rgb(0x8a8a8a)
	colorTextFaint    = rgb(0x5e5e5e)
	colorAccent       = rgb(0xfc5000)
	colorAccentBright = rgb(0xffa066)
	colorSelection    = rgb(0x3d2316)
	colorError        = rgb(0xff5c5c)
	// Blue marks the compared veteran, against the orange of the selected veteran
	colorCompare          = rgb(0x4ea8ff)
	colorCompareSelection = rgb(0x1f2e3f)
)

// highlightColors are the colors of an active control or a chosen table row; the zero value
// is no highlight
type highlightColors struct {
	// Background
	fill color.NRGBA
	// Outline or edge marker
	accent color.NRGBA
}

var (
	selectionHighlight  = highlightColors{fill: colorSelection, accent: colorAccent}
	comparisonHighlight = highlightColors{fill: colorCompareSelection, accent: colorCompare}
)

// rgb converts a 0xRRGGBB literal to an opaque color
func rgb(hex uint32) color.NRGBA {
	return color.NRGBA{R: uint8(hex >> 16), G: uint8(hex >> 8), B: uint8(hex), A: 0xff}
}

// newTheme returns the dark theme, drawing text with shaper
func newTheme(shaper *text.Shaper) *material.Theme {
	theme := material.NewTheme()
	theme.Shaper = shaper
	theme.TextSize = textSize
	theme.Palette = material.Palette{
		Bg:         colorWindow,
		Fg:         colorText,
		ContrastBg: colorAccent,
		ContrastFg: colorTextBright,
	}
	return theme
}

// textLabel returns a single line label of s that ends in an ellipsis when it overflows
func textLabel(theme *material.Theme, size unit.Sp, s string, c color.NRGBA) material.LabelStyle {
	label := material.Label(theme, size, s)
	label.Color = c
	label.MaxLines = 1
	label.WrapPolicy = text.WrapGraphemes
	return label
}

// measureLabel returns the size of label within the constraints, without drawing it
func measureLabel(gtx layout.Context, label material.LabelStyle) image.Point {
	macro := op.Record(gtx.Ops)
	dims := label.Layout(gtx)
	macro.Stop()
	return dims.Size
}

// styleScrollbar sets s to a narrow thumb without a track
func styleScrollbar(s *material.ScrollbarStyle) {
	s.Indicator.Color = colorScrollbar
	s.Indicator.HoverColor = colorScrollHover
	s.Indicator.MinorWidth = scrollbarWidth
	s.Indicator.CornerRadius = scrollbarWidth / 2
}

// fillRect fills r with a solid color
func fillRect(gtx layout.Context, c color.NRGBA, r image.Rectangle) {
	paint.FillShape(gtx.Ops, c, clip.Rect(r).Op())
}

// strokeRRect draws a borderWidth outline inside r, offset half a stroke so it is aligned to
// whole pixels
func strokeRRect(gtx layout.Context, c color.NRGBA, r image.Rectangle, radius int) {
	width := gtx.Dp(borderWidth)
	half := float32(width) / 2
	defer op.Affine(f32.AffineId().Offset(f32.Pt(half, half))).Push(gtx.Ops).Pop()
	r.Max = r.Max.Sub(image.Pt(width, width))
	path := clip.UniformRRect(r, radius).Path(gtx.Ops)
	paint.FillShape(gtx.Ops, c, clip.Stroke{Path: path, Width: float32(width)}.Op())
}

// layoutPill draws label in an outlined capsule button. A nonzero highlight marks the pill as
// active.
func layoutPill(
	gtx layout.Context,
	theme *material.Theme,
	button *widget.Clickable,
	label string,
	highlight highlightColors,
) layout.Dimensions {
	active := highlight != highlightColors{}
	return button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		c := colorTextMuted
		if active || button.Hovered() {
			c = colorText
		}
		padding := image.Pt(gtx.Dp(pillPaddingX), gtx.Dp(pillPaddingY))
		gtx.Constraints.Min = image.Point{}
		gtx.Constraints.Max.X = max(gtx.Constraints.Max.X-2*padding.X, 0)
		macro := op.Record(gtx.Ops)
		offset := op.Offset(padding).Push(gtx.Ops)
		dims := textLabel(theme, smallTextSize, label, c).Layout(gtx)
		offset.Pop()
		call := macro.Stop()

		bounds := image.Rectangle{Max: dims.Size.Add(padding.Mul(2))}
		radius := bounds.Dy() / 2
		border := colorBorderStrong
		switch {
		case active:
			paint.FillShape(gtx.Ops, highlight.fill, clip.UniformRRect(bounds, radius).Op(gtx.Ops))
			border = highlight.accent
		case button.Pressed():
			paint.FillShape(gtx.Ops, colorPressed, clip.UniformRRect(bounds, radius).Op(gtx.Ops))
		case button.Hovered():
			paint.FillShape(gtx.Ops, colorHover, clip.UniformRRect(bounds, radius).Op(gtx.Ops))
		}
		strokeRRect(gtx, border, bounds, radius)
		call.Add(gtx.Ops)
		defer clip.Rect(bounds).Push(gtx.Ops).Pop()
		pointer.CursorPointer.Add(gtx.Ops)
		return layout.Dimensions{Size: bounds.Max}
	})
}
