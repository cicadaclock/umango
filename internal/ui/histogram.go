package ui

import (
	"image"
	"math"
	"slices"
	"strconv"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/cicadaclock/umango/internal/races"
)

const (
	scoreBucketWidth  = 5000
	scoreRangePadding = 10000
	maxAxisIntervals  = 8

	histogramTitle   = "Score vs. Frequency"
	scoreAxisTitle   = "Score"
	countAxisTitle   = "Frequency"
	plotMarginLeft   = unit.Dp(64)
	plotMarginRight  = unit.Dp(24)
	plotMarginTop    = unit.Dp(32)
	plotMarginBottom = unit.Dp(52)
	axisLineWidth    = unit.Dp(1)
	tickLength       = unit.Dp(4)
	// Distance from the left edge to the center of the vertical axis title
	countTitleCenter = unit.Dp(12)
)

// histogramAxis spans 0 to span with tick labels rendered once
type histogramAxis struct {
	span   int
	ticks  []int
	labels []string
}

// newHistogramAxis places ticks at a 1-2-5 step that splits span into at most maxAxisIntervals
func newHistogramAxis(span int) histogramAxis {
	step, magnitude := 1, 1
	for span > step*maxAxisIntervals {
		switch step / magnitude {
		case 1:
			step = 2 * magnitude
		case 2:
			step = 5 * magnitude
		default:
			magnitude *= 10
			step = magnitude
		}
	}
	tickCount := span/step + 1
	a := histogramAxis{
		span:   span,
		ticks:  make([]int, 0, tickCount),
		labels: make([]string, 0, tickCount),
	}
	for v := 0; v <= span; v += step {
		a.ticks = append(a.ticks, v)
		a.labels = append(a.labels, strconv.Itoa(v))
	}
	return a
}

// offset maps v to pixels along an axis drawn length pixels long
func (a histogramAxis) offset(v, length int) int {
	return v * length / a.span
}

// labelStride returns how many ticks apart labels must be so that labels labelSize pixels long
// do not overlap on an axis drawn length pixels long
func (a histogramAxis) labelStride(length, labelSize int) int {
	tickSpacing := a.offset(a.ticks[min(1, len(a.ticks)-1)], length)
	if tickSpacing <= 0 {
		return len(a.ticks)
	}
	return max(1, (labelSize+tickSpacing-1)/tickSpacing)
}

// scoreHistogram holds the score buckets of every summarized veteran on axes shared by all of
// them, so bar heights stay comparable when switching veterans
type scoreHistogram struct {
	// Bucket lower bounds and frequencies by summary index
	bucketScores [][]int
	bucketCounts [][]int
	scoreAxis    histogramAxis
	countAxis    histogramAxis
}

func newScoreHistogram(summary races.TTUmaSummary) scoreHistogram {
	h := scoreHistogram{
		bucketScores: make([][]int, summary.Len()),
		bucketCounts: make([][]int, summary.Len()),
	}
	maxScore, maxCount := 0, 0
	for i, scores := range summary.Scores {
		if scores.Len() == 0 {
			continue
		}
		h.bucketScores[i], h.bucketCounts[i] = scores.HistogramCoords(scoreBucketWidth)
		maxScore = max(maxScore, scores.Max())
		maxCount = max(maxCount, slices.Max(h.bucketCounts[i]))
	}
	h.scoreAxis = newHistogramAxis(maxScore + scoreRangePadding)
	h.countAxis = newHistogramAxis(maxCount + 1)
	return h
}

// layout draws the axes and the buckets of the veteran at summary index, or only the axes
// when index is negative
func (h *scoreHistogram) layout(gtx layout.Context, theme *material.Theme, index int) layout.Dimensions {
	size := gtx.Constraints.Max
	plot := image.Rect(
		gtx.Dp(plotMarginLeft), gtx.Dp(plotMarginTop),
		size.X-gtx.Dp(plotMarginRight), size.Y-gtx.Dp(plotMarginBottom),
	)
	if plot.Dx() <= 0 || plot.Dy() <= 0 {
		return layout.Dimensions{Size: size}
	}
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	gtx.Constraints.Min = image.Point{}

	if index >= 0 {
		counts := h.bucketCounts[index]
		for i, score := range h.bucketScores[index] {
			bar := image.Rect(
				plot.Min.X+h.scoreAxis.offset(score, plot.Dx()),
				plot.Max.Y-h.countAxis.offset(counts[i], plot.Dy()),
				plot.Min.X+h.scoreAxis.offset(score+scoreBucketWidth, plot.Dx()),
				plot.Max.Y,
			)
			paint.FillShape(gtx.Ops, colorPrimary, clip.Rect(bar).Op())
		}
	}

	line, tick := gtx.Dp(axisLineWidth), gtx.Dp(tickLength)
	fillRect(gtx, theme, image.Rect(plot.Min.X-line, plot.Min.Y, plot.Min.X, plot.Max.Y+line))
	fillRect(gtx, theme, image.Rect(plot.Min.X, plot.Max.Y, plot.Max.X, plot.Max.Y+line))
	// The last label is the widest, and every label is as tall as any other
	widestLabel := measureLabel(gtx, theme, h.scoreAxis.labels[len(h.scoreAxis.labels)-1])
	scoreStride := h.scoreAxis.labelStride(plot.Dx(), widestLabel.X+2*tick)
	countStride := h.countAxis.labelStride(plot.Dy(), widestLabel.Y)
	for i, score := range h.scoreAxis.ticks {
		x := plot.Min.X + h.scoreAxis.offset(score, plot.Dx())
		fillRect(gtx, theme, image.Rect(x, plot.Max.Y, x+line, plot.Max.Y+tick))
		if i%scoreStride == 0 {
			drawLabel(gtx, theme, h.scoreAxis.labels[i], image.Pt(x, plot.Max.Y+tick), 0.5, 0)
		}
	}
	for i, count := range h.countAxis.ticks {
		y := plot.Max.Y - h.countAxis.offset(count, plot.Dy())
		fillRect(gtx, theme, image.Rect(plot.Min.X-tick, y, plot.Min.X, y+line))
		if i%countStride == 0 {
			drawLabel(gtx, theme, h.countAxis.labels[i], image.Pt(plot.Min.X-2*tick, y), 1, 0.5)
		}
	}
	drawLabel(gtx, theme, histogramTitle, image.Pt(size.X/2, plot.Min.Y/2), 0.5, 0.5)
	drawLabel(gtx, theme, scoreAxisTitle, image.Pt((plot.Min.X+plot.Max.X)/2, size.Y), 0.5, 1)
	drawVerticalLabel(gtx, theme, countAxisTitle, image.Pt(gtx.Dp(countTitleCenter), (plot.Min.Y+plot.Max.Y)/2))
	return layout.Dimensions{Size: size}
}

func fillRect(gtx layout.Context, theme *material.Theme, r image.Rectangle) {
	paint.FillShape(gtx.Ops, theme.Fg, clip.Rect(r).Op())
}

func measureLabel(gtx layout.Context, theme *material.Theme, s string) image.Point {
	macro := op.Record(gtx.Ops)
	dims := material.Body1(theme, s).Layout(gtx)
	macro.Stop()
	return dims.Size
}

// drawLabel draws s so that the point at fractions (alignX, alignY) of its bounds lands on anchor
func drawLabel(gtx layout.Context, theme *material.Theme, s string, anchor image.Point, alignX, alignY float32) {
	macro := op.Record(gtx.Ops)
	dims := material.Body1(theme, s).Layout(gtx)
	call := macro.Stop()
	origin := anchor.Sub(image.Pt(
		int(float32(dims.Size.X)*alignX),
		int(float32(dims.Size.Y)*alignY),
	))
	defer op.Offset(origin).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
}

// drawVerticalLabel draws s reading bottom to top, centered on center
func drawVerticalLabel(gtx layout.Context, theme *material.Theme, s string, center image.Point) {
	macro := op.Record(gtx.Ops)
	dims := material.Body1(theme, s).Layout(gtx)
	call := macro.Stop()
	// A quarter turn counterclockwise moves the label box above the origin, so shift it down by
	// its new height
	transform := f32.AffineId().
		Rotate(f32.Point{}, -math.Pi/2).
		Offset(f32.Pt(float32(center.X-dims.Size.Y/2), float32(center.Y+dims.Size.X/2)))
	defer op.Affine(transform).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
}
