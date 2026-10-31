package ui

import (
	"fmt"
	"image"
	"image/color"
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
	// Shares of races are counted in tenths of a percent, so share axes reuse the integer axis
	permillePerPercent = 10

	scoreAxisTitle   = "Score"
	countAxisTitle   = "Frequency"
	shareAxisTitle   = "Share of Races"
	plotMarginLeft   = unit.Dp(56)
	plotMarginRight  = unit.Dp(20)
	plotMarginTop    = unit.Dp(20)
	plotMarginBottom = unit.Dp(46)
	tickLength       = unit.Dp(4)
	barGap           = unit.Dp(1)
	barRadius        = unit.Dp(2)
	outlineWidth     = unit.Dp(2)
	tooltipPaddingX  = unit.Dp(8)
	tooltipPaddingY  = unit.Dp(6)
	tooltipLift      = unit.Dp(6)
	swatchSize       = unit.Dp(8)
	swatchGap        = unit.Dp(6)
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

// formatPermille writes a share in tenths of a percent as a percentage, as "12.5%"
func formatPermille(v int) string {
	if v%permillePerPercent == 0 {
		return strconv.Itoa(v/permillePerPercent) + "%"
	}
	return strconv.FormatFloat(float64(v)/permillePerPercent, 'f', 1, 64) + "%"
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

// scoreHistogram draws the score buckets of the selected veteran, alone or against a compared
// veteran, on a score axis shared by every summarized veteran
//
// A lone veteran's bars count races on an axis shared by every veteran, so bar heights stay
// comparable when switching veterans. While comparing, the selected veteran's bars and the
// compared veteran's stepped outline give shares of their own races on an axis fitted to the
// pair, so veterans with different race counts compare fairly.
type scoreHistogram struct {
	// Bucket lower bounds, frequencies, and frequencies in tenths of a percent of the veteran's
	// races, by summary index
	bucketScores [][]int
	bucketCounts [][]int
	bucketShares [][]int
	scoreAxis    histogramAxis
	countAxis    histogramAxis
	// Fitted to the compared pair by show
	shareAxis histogramAxis
	// Summary indices of the selected and the compared veteran, -1 when none
	selected, compared int
	// Pointer position in the histogram, meaningful while hovering the plot
	pointer  f32.Point
	hovering bool
	// Tooltip lines of the bucket and veterans at tooltipKey, formatted again only when the
	// pointer reaches another bucket or other veterans are shown
	tooltipKey   tooltipKey
	tooltipLines [3]string
}

// tooltipKey identifies the bucket a tooltip describes and the veterans it describes
type tooltipKey struct {
	selected, compared, score int
}

func newScoreHistogram(summary races.TTUmaSummary) scoreHistogram {
	h := scoreHistogram{
		bucketScores: make([][]int, summary.Len()),
		bucketCounts: make([][]int, summary.Len()),
		bucketShares: make([][]int, summary.Len()),
		selected:     -1,
		compared:     -1,
		// No bucket starts below zero, so no tooltip matches until the first is formatted
		tooltipKey: tooltipKey{score: -1},
	}
	maxScore, maxCount := 0, 0
	for i, scores := range summary.Scores {
		if scores.Len() == 0 {
			continue
		}
		h.bucketScores[i], h.bucketCounts[i] = scores.HistogramCoords(scoreBucketWidth)
		h.bucketShares[i] = make([]int, len(h.bucketCounts[i]))
		for bucket, count := range h.bucketCounts[i] {
			h.bucketShares[i][bucket] = count * 100 * permillePerPercent / scores.Len()
		}
		maxScore = max(maxScore, scores.Max())
		maxCount = max(maxCount, slices.Max(h.bucketCounts[i]))
	}
	h.scoreAxis = newHistogramAxis(maxScore+scoreRangePadding, formatThousands)
	h.countAxis = newHistogramAxis(maxCount+1, strconv.Itoa)
	return h
}

// show sets the veterans to draw, fitting the share axis to a compared pair
//
// Inputs:
// - selected: Summary index of the veteran whose bars to draw, -1 to draw only the axes
// - compared: Summary index of the veteran to outline against selected, -1 for none
func (h *scoreHistogram) show(selected, compared int) {
	h.selected, h.compared = selected, compared
	if compared < 0 {
		return
	}
	maxShare := max(slices.Max(h.bucketShares[selected]), slices.Max(h.bucketShares[compared]))
	h.shareAxis = newHistogramAxis(maxShare+permillePerPercent, formatPermille)
}

// heights returns the bar heights by summary index and the axis they rise along: race counts
// for a lone veteran, and shares of races while comparing
func (h *scoreHistogram) heights() ([][]int, *histogramAxis) {
	if h.compared >= 0 {
		return h.bucketShares, &h.shareAxis
	}
	return h.bucketCounts, &h.countAxis
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

// layout draws the axes and the shown veterans' buckets, or only the axes when no veteran is
// selected, with a readout of the bucket under the pointer
func (h *scoreHistogram) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
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

	hovered := h.hoveredScore(plot)
	if hovered >= 0 {
		fillRect(gtx, colorHoverColumn, h.columnRect(plot, hovered))
	}

	_, heightAxis := h.heights()
	line, tick := gtx.Dp(borderWidth), gtx.Dp(tickLength)
	// The last label is the widest, and every label is as tall as any other
	widestLabel := measureLabel(gtx, axisLabel(theme, h.scoreAxis.labels[len(h.scoreAxis.labels)-1]))
	heightStride := heightAxis.labelStride(plot.Dy(), 2*widestLabel.Y)
	for i, height := range heightAxis.ticks {
		if i%heightStride != 0 {
			continue
		}
		y := plot.Max.Y - heightAxis.offset(height, plot.Dy())
		if height > 0 {
			fillRect(gtx, colorGridLine, image.Rect(plot.Min.X, y, plot.Max.X, y+line))
		}
		drawLabel(gtx, axisLabel(theme, heightAxis.labels[i]), image.Pt(plot.Min.X-2*tick, y), 1, 0.5)
	}

	if h.selected >= 0 {
		h.drawBars(gtx, plot, hovered)
	}
	if h.compared >= 0 {
		h.drawOutline(gtx, plot)
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
	heightTitle := countAxisTitle
	if h.compared >= 0 {
		heightTitle = shareAxisTitle
	}
	countTitle := image.Pt(gtx.Dp(countTitleCenter), (plot.Min.Y+plot.Max.Y)/2)
	drawVerticalLabel(gtx, axisTitle(theme, heightTitle), countTitle)
	if hovered >= 0 {
		h.drawTooltip(gtx, theme, plot, hovered)
	}

	area := clip.Rect(plot).Push(gtx.Ops)
	event.Op(gtx.Ops, h)
	area.Pop()
	return layout.Dimensions{Size: size}
}

// hoveredScore returns the lower bound of the bucket under the pointer when a shown veteran has
// races in it, or -1
func (h *scoreHistogram) hoveredScore(plot image.Rectangle) int {
	if !h.hovering || h.selected < 0 {
		return -1
	}
	score := int(h.pointer.X-float32(plot.Min.X)) * h.scoreAxis.span / plot.Dx()
	score = score / scoreBucketWidth * scoreBucketWidth
	if _, found := h.bucketAt(h.selected, score); found {
		return score
	}
	if h.compared >= 0 {
		if _, found := h.bucketAt(h.compared, score); found {
			return score
		}
	}
	return -1
}

// bucketAt returns the position among the buckets of the veteran at summary index of the one
// starting at score, and whether the veteran has races there
func (h *scoreHistogram) bucketAt(index, score int) (int, bool) {
	return slices.BinarySearch(h.bucketScores[index], score)
}

// columnRect returns the full height plot column over the bucket starting at score
func (h *scoreHistogram) columnRect(plot image.Rectangle, score int) image.Rectangle {
	return image.Rect(
		plot.Min.X+h.scoreAxis.offset(score, plot.Dx()), plot.Min.Y,
		plot.Min.X+h.scoreAxis.offset(score+scoreBucketWidth, plot.Dx()), plot.Max.Y,
	)
}

// barRect returns the plot area of the veteran's bucket at position bucket
func (h *scoreHistogram) barRect(plot image.Rectangle, index, bucket int) image.Rectangle {
	heights, axis := h.heights()
	bar := h.columnRect(plot, h.bucketScores[index][bucket])
	bar.Min.Y = plot.Max.Y - axis.offset(heights[index][bucket], plot.Dy())
	return bar
}

// drawBars fills the selected veteran's buckets as one path shaded by a gradient down the plot,
// then fills its bar in the bucket starting at hovered, if any, in the bright accent
func (h *scoreHistogram) drawBars(gtx layout.Context, plot image.Rectangle, hovered int) {
	gap, radius := float32(gtx.Dp(barGap)), float32(gtx.Dp(barRadius))
	var path clip.Path
	path.Begin(gtx.Ops)
	for bucket := range h.bucketScores[h.selected] {
		appendBar(&path, h.barRect(plot, h.selected, bucket), gap, radius)
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
	bucket, found := h.bucketAt(h.selected, hovered)
	if !found {
		return
	}
	path.Begin(gtx.Ops)
	appendBar(&path, h.barRect(plot, h.selected, bucket), gap, radius)
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

// drawOutline traces the compared veteran's buckets in the compare color as a stepped line that
// drops to the baseline wherever its buckets break off
//
// The line's segments are all horizontal or vertical, so they are filled as pixel aligned
// rectangles, which stay crisp where a stroke would antialias and can leak coverage.
func (h *scoreHistogram) drawOutline(gtx layout.Context, plot image.Rectangle) {
	width := gtx.Dp(outlineWidth)
	center := image.Pt(width/2, width/2)
	segment := func(from, to image.Point) {
		r := image.Rectangle{Min: from, Max: to}.Canon()
		fillRect(gtx, colorCompare, image.Rectangle{
			Min: r.Min.Sub(center),
			Max: r.Max.Sub(center).Add(image.Pt(width, width)),
		})
	}
	scores := h.bucketScores[h.compared]
	edgeTop := plot.Max.Y
	for bucket, score := range scores {
		bar := h.barRect(plot, h.compared, bucket)
		if bucket > 0 && scores[bucket-1] != score-scoreBucketWidth {
			edgeTop = plot.Max.Y
		}
		segment(image.Pt(bar.Min.X, edgeTop), bar.Min)
		segment(bar.Min, image.Pt(bar.Max.X, bar.Min.Y))
		edgeTop = bar.Min.Y
		if bucket == len(scores)-1 || scores[bucket+1] != score+scoreBucketWidth {
			segment(image.Pt(bar.Max.X, bar.Min.Y), bar.Max)
		}
	}
}

// drawTooltip describes the bucket starting at score in a box above its tallest bar, kept within
// the plot's width
//
// A lone veteran's box gives its race count in the bucket, the score range, and its chance of
// scoring at least the bucket's lower bound. A compared pair's box gives the score range over a
// line per veteran, keyed by a swatch of its color, with its count, share and chance.
func (h *scoreHistogram) drawTooltip(
	gtx layout.Context,
	theme *material.Theme,
	plot image.Rectangle,
	score int,
) {
	if key := (tooltipKey{h.selected, h.compared, score}); key != h.tooltipKey {
		h.tooltipKey = key
		h.formatTooltip(score)
	}
	var swatches [2]color.NRGBA
	if h.compared >= 0 {
		swatches = [2]color.NRGBA{colorAccent, colorCompare}
	}
	title := textLabel(theme, smallTextSize, h.tooltipLines[0], colorText)
	title.Font.Weight = font.SemiBold
	macro := op.Record(gtx.Ops)
	textDims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(title.Layout),
		layout.Rigid(tooltipLine(theme, h.tooltipLines[1], swatches[0])),
		layout.Rigid(tooltipLine(theme, h.tooltipLines[2], swatches[1])),
	)
	text := macro.Stop()

	top := plot.Max.Y
	for _, index := range [...]int{h.selected, h.compared} {
		if index < 0 {
			continue
		}
		if bucket, found := h.bucketAt(index, score); found {
			top = min(top, h.barRect(plot, index, bucket).Min.Y)
		}
	}
	paddingX, paddingY := gtx.Dp(tooltipPaddingX), gtx.Dp(tooltipPaddingY)
	box := image.Rectangle{Max: textDims.Size.Add(image.Pt(2*paddingX, 2*paddingY))}
	column := h.columnRect(plot, score)
	origin := image.Pt(
		min(max((column.Min.X+column.Max.X-box.Dx())/2, plot.Min.X), plot.Max.X-box.Dx()),
		max(top-gtx.Dp(tooltipLift)-box.Dy(), 0),
	)
	defer op.Offset(origin).Push(gtx.Ops).Pop()
	radius := gtx.Dp(controlRadius)
	paint.FillShape(gtx.Ops, colorTooltip, clip.UniformRRect(box, radius).Op(gtx.Ops))
	strokeRRect(gtx, colorBorderStrong, box, radius)
	defer op.Offset(image.Pt(paddingX, paddingY)).Push(gtx.Ops).Pop()
	text.Add(gtx.Ops)
}

// tooltipLine returns a widget drawing s as a line of tooltip text, led by a swatch of the color
// swatch unless it is transparent
func tooltipLine(theme *material.Theme, s string, swatch color.NRGBA) layout.Widget {
	label := axisLabel(theme, s)
	if swatch.A == 0 {
		return label.Layout
	}
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				side := gtx.Dp(swatchSize)
				square := image.Rectangle{Max: image.Pt(side, side)}
				paint.FillShape(gtx.Ops, swatch, clip.UniformRRect(square, gtx.Dp(barRadius)).Op(gtx.Ops))
				return layout.Dimensions{Size: square.Max}
			}),
			layout.Rigid(layout.Spacer{Width: swatchGap}.Layout),
			layout.Rigid(label.Layout),
		)
	}
}

// formatTooltip formats the tooltip lines describing the bucket starting at score
func (h *scoreHistogram) formatTooltip(score int) {
	scoreRange := formatThousands(score) + " – " + formatThousands(score+scoreBucketWidth)
	if h.compared < 0 {
		inBucket, atLeast, total := h.bucketStats(h.selected, score)
		h.tooltipLines = [3]string{
			formatRaceCount(inBucket), scoreRange, formatScoreChance(atLeast, total, score),
		}
		return
	}
	h.tooltipLines = [3]string{
		scoreRange, h.formatBucketShare(h.selected, score), h.formatBucketShare(h.compared, score),
	}
}

// formatBucketShare writes the races of the veteran at summary index in the bucket starting at
// score, as a count and a share of its races, beside its chance of scoring at least score, as
// "3 races (12.5%) · 34.2% at ≥ 70k"
func (h *scoreHistogram) formatBucketShare(index, score int) string {
	inBucket, atLeast, total := h.bucketStats(index, score)
	return fmt.Sprintf("%s (%.1f%%) · %s",
		formatRaceCount(inBucket),
		100*float64(inBucket)/float64(total),
		formatScoreChance(atLeast, total, score),
	)
}

// bucketStats counts the races of the veteran at summary index that scored in the bucket
// starting at score, that scored at least score, and in total
func (h *scoreHistogram) bucketStats(index, score int) (inBucket, atLeast, total int) {
	for bucket, count := range h.bucketCounts[index] {
		bucketScore := h.bucketScores[index][bucket]
		total += count
		if bucketScore >= score {
			atLeast += count
		}
		if bucketScore == score {
			inBucket = count
		}
	}
	return inBucket, atLeast, total
}

// formatRaceCount writes count as a number of races
func formatRaceCount(count int) string {
	if count == 1 {
		return "1 race"
	}
	return fmt.Sprintf("%d races", count)
}

// formatScoreChance writes the share of total races that scored at least score, atLeast of them,
// as "34.2% at ≥ 70k"
func formatScoreChance(atLeast, total, score int) string {
	return fmt.Sprintf("%.1f%% at ≥ %s", 100*float64(atLeast)/float64(total), formatThousands(score))
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
