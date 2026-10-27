package ui

import (
	"fmt"
	"image"
	"math"
	"slices"
	"strconv"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
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

	scoreAxisTitle   = "Score"
	countAxisTitle   = "Frequency"
	plotMarginLeft   = unit.Dp(56)
	plotMarginRight  = unit.Dp(20)
	plotMarginTop    = unit.Dp(20)
	plotMarginBottom = unit.Dp(46)
	tickLength       = unit.Dp(4)
	barGap           = unit.Dp(1)
	barRadius        = unit.Dp(2)
	tooltipPaddingX  = unit.Dp(8)
	tooltipPaddingY  = unit.Dp(6)
	tooltipLift      = unit.Dp(6)
	// Distance from the left edge to the center of the vertical axis title
	countTitleCenter = unit.Dp(14)
	// Distance from the bottom edge to the bottom of the horizontal axis title
	scoreTitleInset = unit.Dp(4)
)

var (
	// Bars deepen from a bright orange at the top of the plot to a dark one at the baseline, so
	// their heights also read as color
	colorBarTop  = rgb(0xff7a30)
	colorBarBase = rgb(0xb83800)
	// Backdrop of the plot column of the hovered bucket
	colorHoverColumn = rgb(0x1c1c1c)
)

// histogramAxis spans 0 to span with tick labels rendered once
type histogramAxis struct {
	span   int
	ticks  []int
	labels []string
}

// newHistogramAxis places ticks at a 1-2-5 step that splits span into at most maxAxisIntervals,
// labeled by format
func newHistogramAxis(span int, format func(int) string) histogramAxis {
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
		a.labels = append(a.labels, format(v))
	}
	return a
}

// formatThousands writes a multiple of 1000 in thousands, as "75k"
func formatThousands(v int) string {
	if v == 0 {
		return "0"
	}
	return strconv.Itoa(v/1000) + "k"
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
	// Pointer position in the histogram, meaningful while hovering the plot
	pointer  f32.Point
	hovering bool
	// Tooltip lines of the bucket starting at tooltipScore of the veteran at tooltipIndex,
	// formatted again only when the pointer reaches another bucket
	tooltipIndex, tooltipScore int
	tooltipCount, tooltipRange string
}

func newScoreHistogram(summary races.TTUmaSummary) scoreHistogram {
	h := scoreHistogram{
		bucketScores: make([][]int, summary.Len()),
		bucketCounts: make([][]int, summary.Len()),
		tooltipIndex: -1,
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
	h.scoreAxis = newHistogramAxis(maxScore+scoreRangePadding, formatThousands)
	h.countAxis = newHistogramAxis(maxCount+1, strconv.Itoa)
	return h
}

// update tracks the pointer over the plot
func (h *scoreHistogram) update(gtx layout.Context) {
	for {
		e, ok := gtx.Event(pointer.Filter{
			Target: h,
			Kinds:  pointer.Enter | pointer.Move | pointer.Leave,
		})
		if !ok {
			return
		}
		if e, ok := e.(pointer.Event); ok {
			h.hovering = e.Kind == pointer.Enter || e.Kind == pointer.Move
			h.pointer = e.Position
		}
	}
}

// layout draws the axes and the buckets of the veteran at summary index, or only the axes
// when index is negative, with a readout of the bucket under the pointer
func (h *scoreHistogram) layout(
	gtx layout.Context,
	theme *material.Theme,
	index int,
) layout.Dimensions {
	h.update(gtx)
	size := gtx.Constraints.Max
	plot := image.Rect(
		gtx.Dp(plotMarginLeft), gtx.Dp(plotMarginTop),
		size.X-gtx.Dp(plotMarginRight), size.Y-gtx.Dp(plotMarginBottom),
	)
	if plot.Dx() <= 0 || plot.Dy() <= 0 {
		return layout.Dimensions{Size: size}
	}
	gtx.Constraints.Min = image.Point{}

	hovered := -1
	if index >= 0 {
		hovered = h.hoveredBucket(plot, index)
	}
	if hovered >= 0 {
		bar := h.barRect(plot, index, hovered)
		fillRect(gtx, colorHoverColumn, image.Rect(bar.Min.X, plot.Min.Y, bar.Max.X, plot.Max.Y))
	}

	line, tick := gtx.Dp(borderWidth), gtx.Dp(tickLength)
	// The last label is the widest, and every label is as tall as any other
	widestLabel := measureLabel(gtx, axisLabel(theme, h.scoreAxis.labels[len(h.scoreAxis.labels)-1]))
	countStride := h.countAxis.labelStride(plot.Dy(), 2*widestLabel.Y)
	for i, count := range h.countAxis.ticks {
		if i%countStride != 0 {
			continue
		}
		y := plot.Max.Y - h.countAxis.offset(count, plot.Dy())
		if count > 0 {
			fillRect(gtx, colorGridLine, image.Rect(plot.Min.X, y, plot.Max.X, y+line))
		}
		drawLabel(gtx, axisLabel(theme, h.countAxis.labels[i]), image.Pt(plot.Min.X-2*tick, y), 1, 0.5)
	}

	if index >= 0 {
		h.drawBars(gtx, plot, index, hovered)
	}

	fillRect(gtx, colorBorderStrong, image.Rect(plot.Min.X, plot.Max.Y, plot.Max.X, plot.Max.Y+line))
	scoreStride := h.scoreAxis.labelStride(plot.Dx(), widestLabel.X+4*tick)
	for i, score := range h.scoreAxis.ticks {
		x := plot.Min.X + h.scoreAxis.offset(score, plot.Dx())
		fillRect(gtx, colorBorderStrong, image.Rect(x, plot.Max.Y, x+line, plot.Max.Y+tick))
		if i%scoreStride == 0 {
			drawLabel(gtx, axisLabel(theme, h.scoreAxis.labels[i]), image.Pt(x, plot.Max.Y+tick), 0.5, 0)
		}
	}
	scoreTitle := image.Pt((plot.Min.X+plot.Max.X)/2, size.Y-gtx.Dp(scoreTitleInset))
	drawLabel(gtx, axisTitle(theme, scoreAxisTitle), scoreTitle, 0.5, 1)
	countTitle := image.Pt(gtx.Dp(countTitleCenter), (plot.Min.Y+plot.Max.Y)/2)
	drawVerticalLabel(gtx, axisTitle(theme, countAxisTitle), countTitle)
	if hovered >= 0 {
		h.drawTooltip(gtx, theme, plot, index, hovered)
	}

	area := clip.Rect(plot).Push(gtx.Ops)
	event.Op(gtx.Ops, h)
	area.Pop()
	return layout.Dimensions{Size: size}
}

// hoveredBucket returns the position among the veteran's buckets of the one under the pointer,
// or -1 when the pointer is not over a bar's column
func (h *scoreHistogram) hoveredBucket(plot image.Rectangle, index int) int {
	if !h.hovering {
		return -1
	}
	score := int(h.pointer.X-float32(plot.Min.X)) * h.scoreAxis.span / plot.Dx()
	return slices.Index(h.bucketScores[index], score/scoreBucketWidth*scoreBucketWidth)
}

// barRect returns the plot area of the veteran's bucket at position bucket
func (h *scoreHistogram) barRect(plot image.Rectangle, index, bucket int) image.Rectangle {
	score := h.bucketScores[index][bucket]
	return image.Rect(
		plot.Min.X+h.scoreAxis.offset(score, plot.Dx()),
		plot.Max.Y-h.countAxis.offset(h.bucketCounts[index][bucket], plot.Dy()),
		plot.Min.X+h.scoreAxis.offset(score+scoreBucketWidth, plot.Dx()),
		plot.Max.Y,
	)
}

// drawBars fills the veteran's buckets as one path shaded by a gradient down the plot, then
// fills the hovered bar, when not negative, in the bright accent
func (h *scoreHistogram) drawBars(gtx layout.Context, plot image.Rectangle, index, hovered int) {
	gap, radius := float32(gtx.Dp(barGap)), float32(gtx.Dp(barRadius))
	var path clip.Path
	path.Begin(gtx.Ops)
	for bucket := range h.bucketScores[index] {
		appendBar(&path, h.barRect(plot, index, bucket), gap, radius)
	}
	bars := clip.Outline{Path: path.End()}.Op().Push(gtx.Ops)
	paint.LinearGradientOp{
		Stop1:  f32.Pt(0, float32(plot.Min.Y)),
		Color1: colorBarTop,
		Stop2:  f32.Pt(0, float32(plot.Max.Y)),
		Color2: colorBarBase,
	}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	bars.Pop()
	if hovered < 0 {
		return
	}
	path.Begin(gtx.Ops)
	appendBar(&path, h.barRect(plot, index, hovered), gap, radius)
	paint.FillShape(gtx.Ops, colorAccentBright, clip.Outline{Path: path.End()}.Op())
}

// appendBar adds bar to path as a closed outline with its top corners rounded by radius, its
// right edge pulled in by gap to keep neighboring bars apart
func appendBar(path *clip.Path, bar image.Rectangle, gap, radius float32) {
	left, top, bottom := float32(bar.Min.X), float32(bar.Min.Y), float32(bar.Max.Y)
	right := max(float32(bar.Max.X)-gap, left+1)
	radius = min(radius, (right-left)/2, bottom-top)
	path.MoveTo(f32.Pt(left, bottom))
	path.LineTo(f32.Pt(left, top+radius))
	path.QuadTo(f32.Pt(left, top), f32.Pt(left+radius, top))
	path.LineTo(f32.Pt(right-radius, top))
	path.QuadTo(f32.Pt(right, top), f32.Pt(right, top+radius))
	path.LineTo(f32.Pt(right, bottom))
	path.Close()
}

// drawTooltip draws the race count and score range of the veteran's bucket at position bucket
// in a box above its bar, kept within the plot's width
func (h *scoreHistogram) drawTooltip(
	gtx layout.Context,
	theme *material.Theme,
	plot image.Rectangle,
	index, bucket int,
) {
	score := h.bucketScores[index][bucket]
	if index != h.tooltipIndex || score != h.tooltipScore {
		h.tooltipIndex, h.tooltipScore = index, score
		h.tooltipCount = formatRaceCount(h.bucketCounts[index][bucket])
		h.tooltipRange = formatThousands(score) + " – " + formatThousands(score+scoreBucketWidth)
	}
	count := textLabel(theme, smallTextSize, h.tooltipCount, colorText)
	count.Font.Weight = font.SemiBold
	macro := op.Record(gtx.Ops)
	countDims := count.Layout(gtx)
	offset := op.Offset(image.Pt(0, countDims.Size.Y)).Push(gtx.Ops)
	rangeDims := axisLabel(theme, h.tooltipRange).Layout(gtx)
	offset.Pop()
	text := macro.Stop()

	paddingX, paddingY := gtx.Dp(tooltipPaddingX), gtx.Dp(tooltipPaddingY)
	box := image.Rectangle{Max: image.Pt(
		max(countDims.Size.X, rangeDims.Size.X)+2*paddingX,
		countDims.Size.Y+rangeDims.Size.Y+2*paddingY,
	)}
	bar := h.barRect(plot, index, bucket)
	origin := image.Pt(
		min(max((bar.Min.X+bar.Max.X-box.Dx())/2, plot.Min.X), plot.Max.X-box.Dx()),
		max(bar.Min.Y-gtx.Dp(tooltipLift)-box.Dy(), 0),
	)
	defer op.Offset(origin).Push(gtx.Ops).Pop()
	radius := gtx.Dp(controlRadius)
	paint.FillShape(gtx.Ops, colorTooltip, clip.UniformRRect(box, radius).Op(gtx.Ops))
	strokeRRect(gtx, colorBorderStrong, box, radius)
	defer op.Offset(image.Pt(paddingX, paddingY)).Push(gtx.Ops).Pop()
	text.Add(gtx.Ops)
}

// formatRaceCount writes count as a number of races
func formatRaceCount(count int) string {
	if count == 1 {
		return "1 race"
	}
	return fmt.Sprintf("%d races", count)
}

func axisLabel(theme *material.Theme, s string) material.LabelStyle {
	return textLabel(theme, captionTextSize, s, colorTextMuted)
}

func axisTitle(theme *material.Theme, s string) material.LabelStyle {
	label := textLabel(theme, captionTextSize, s, colorTextFaint)
	label.Font.Weight = font.SemiBold
	return label
}

// drawLabel draws label so that the point at fractions (alignX, alignY) of its bounds lands on
// anchor
func drawLabel(
	gtx layout.Context,
	label material.LabelStyle,
	anchor image.Point,
	alignX, alignY float32,
) {
	macro := op.Record(gtx.Ops)
	dims := label.Layout(gtx)
	call := macro.Stop()
	origin := anchor.Sub(image.Pt(
		int(float32(dims.Size.X)*alignX),
		int(float32(dims.Size.Y)*alignY),
	))
	defer op.Offset(origin).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
}

// drawVerticalLabel draws label reading bottom to top, centered on center
func drawVerticalLabel(gtx layout.Context, label material.LabelStyle, center image.Point) {
	macro := op.Record(gtx.Ops)
	dims := label.Layout(gtx)
	call := macro.Stop()
	// A quarter turn counterclockwise moves the label box above the origin, so shift it down by
	// its new height
	transform := f32.AffineId().
		Rotate(f32.Point{}, -math.Pi/2).
		Offset(f32.Pt(float32(center.X-dims.Size.Y/2), float32(center.Y+dims.Size.X/2)))
	defer op.Affine(transform).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
}
