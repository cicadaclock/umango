package ui

import (
	"cmp"
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/cicadaclock/umango/internal/races"
)

const (
	breakdownLineHeight   = unit.Dp(16)
	breakdownRowPadding   = unit.Dp(4)
	breakdownBarThickness = unit.Dp(8)
	breakdownColumnGap    = unit.Dp(8)
	// Widest average points and share that the value columns must fit
	breakdownPointsSample = "00000"
	breakdownShareSample  = "00.0%"
	// Percent of a veteran's score that a category must reach in some veteran to be listed
	minListedSharePercent = 1
)

var (
	// Bar color of each chosen veteran's line, by role
	breakdownBarColors = [chosenVetCount]color.NRGBA{
		selectedVet: colorAccent,
		comparedVet: colorCompare,
	}
	// Categories that the breakdown never lists
	unlistedCategories = [...]races.ScoreCategory{races.MarginScore, races.LongShotScore}
)

// scoreBreakdown lists the average points per race in each score category of the selected
// veteran, and of the compared veteran beneath it, with their share of the average score
type scoreBreakdown struct {
	// Points by category and race scores, by summary index
	breakdowns []races.ScoreBreakdown
	scores     []races.ScoreArray
	// Listed categories in display order
	categories []races.ScoreCategory
	// Summary indices of the shown veterans by role, -1 when none
	shown [chosenVetCount]int
	// Average points per race by role and category, and their text
	averages [chosenVetCount][races.ScoreCategoryCount]float64
	points   [chosenVetCount][races.ScoreCategoryCount]string
	shares   [chosenVetCount][races.ScoreCategoryCount]string
	// Largest average magnitude shown, which a full bar length represents
	fullBar float64
	list    widget.List
	// Column widths in pixels, measured at fitMetric
	fitMetric                           unit.Metric
	labelWidth, pointsWidth, shareWidth int
}

// newScoreBreakdown selects the categories to list from the breakdowns of every veteran
func newScoreBreakdown(summary races.TTUmaSummary) scoreBreakdown {
	var listed [races.ScoreCategoryCount]bool
	for i, breakdown := range summary.ScoreBreakdowns {
		if summary.Scores[i].Len() == 0 {
			continue
		}
		total := summary.Scores[i].Sum
		for category, points := range breakdown {
			listed[category] = listed[category] ||
				100*max(points, -points) >= minListedSharePercent*total
		}
	}
	for _, category := range unlistedCategories {
		listed[category] = false
	}
	categories := make([]races.ScoreCategory, 0, races.ScoreCategoryCount)
	for category := range races.ScoreCategoryCount {
		if listed[category] {
			categories = append(categories, category)
		}
	}
	return scoreBreakdown{
		breakdowns: summary.ScoreBreakdowns,
		scores:     summary.Scores,
		categories: categories,
		shown:      [chosenVetCount]int{-1, -1},
		list:       widget.List{List: layout.List{Axis: layout.Vertical}},
	}
}

// show sets the veterans to list, orders the categories by the points of the selected
// veteran, and formats the averages
func (b *scoreBreakdown) show(selected, compared int) {
	b.shown = [chosenVetCount]int{selected, compared}
	if selected >= 0 {
		breakdown := &b.breakdowns[selected]
		slices.SortFunc(b.categories, func(x, y races.ScoreCategory) int {
			return cmp.Or(cmp.Compare(breakdown[y], breakdown[x]), cmp.Compare(x, y))
		})
	}
	// A floor of one point keeps bar lengths finite for a veteran without points
	b.fullBar = 1
	for role, index := range b.shown {
		if index < 0 {
			continue
		}
		breakdown := &b.breakdowns[index]
		raceCount := float64(b.scores[index].Len())
		total := float64(b.scores[index].Sum)
		for _, category := range b.categories {
			points := float64(breakdown[category])
			average := points / raceCount
			b.averages[role][category] = average
			b.points[role][category] = strconv.Itoa(int(math.Round(average)))
			b.shares[role][category] = fmt.Sprintf("%.1f%%", 100*points/total)
			b.fullBar = max(b.fullBar, math.Abs(average))
		}
	}
}

// layout draws one row per listed category in a scrolling list
func (b *scoreBreakdown) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	if b.shown[selectedVet] < 0 {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	if gtx.Metric != b.fitMetric {
		b.fitColumns(gtx, theme)
	}
	list := material.List(theme, &b.list)
	// The thumb fits in the right padding, so rows keep the panel's full width
	list.AnchorStrategy = material.Overlay
	styleScrollbar(&list.ScrollbarStyle)
	return list.Layout(gtx, len(b.categories), func(gtx layout.Context, row int) layout.Dimensions {
		return b.layoutRow(gtx, theme, b.categories[row])
	})
}

// fitColumns sizes the name column to the widest category name and the value columns to
// their samples
func (b *scoreBreakdown) fitColumns(gtx layout.Context, theme *material.Theme) {
	gtx.Constraints = layout.Constraints{Max: image.Pt(math.MaxInt32, math.MaxInt32)}
	b.fitMetric = gtx.Metric
	b.labelWidth = 0
	for _, category := range b.categories {
		label := textLabel(theme, smallTextSize, category.String(), colorText)
		b.labelWidth = max(b.labelWidth, measureLabel(gtx, label).X)
	}
	b.pointsWidth = measureLabel(gtx, textLabel(theme, smallTextSize, breakdownPointsSample, colorText)).X
	b.shareWidth = measureLabel(gtx, textLabel(theme, smallTextSize, breakdownShareSample, colorText)).X
}

// layoutRow draws a category name beside one line per shown veteran of bar, average points
// and share
func (b *scoreBreakdown) layoutRow(
	gtx layout.Context,
	theme *material.Theme,
	category races.ScoreCategory,
) layout.Dimensions {
	lineCount := 1
	if b.shown[comparedVet] >= 0 {
		lineCount = 2
	}
	lineHeight, padding := gtx.Dp(breakdownLineHeight), gtx.Dp(breakdownRowPadding)
	size := image.Pt(gtx.Constraints.Max.X, lineCount*lineHeight+2*padding)
	gtx.Constraints.Min = image.Point{}

	edge, gap := gtx.Dp(tableCellPadding), gtx.Dp(breakdownColumnGap)
	name := textLabel(theme, smallTextSize, category.String(), colorText)
	drawLabel(gtx, name, image.Pt(edge, size.Y/2), 0, 0.5)
	pointsRight := size.X - edge - b.shareWidth - gap
	track := image.Rect(edge+b.labelWidth+gap, 0, pointsRight-b.pointsWidth-gap, 0)
	thickness, radius := gtx.Dp(breakdownBarThickness), gtx.Dp(barRadius)
	for role := range lineCount {
		center := padding + role*lineHeight + lineHeight/2
		track.Min.Y, track.Max.Y = center-thickness/2, center-thickness/2+thickness
		average := b.averages[role][category]
		bar := track
		bar.Max.X = bar.Min.X + int(math.Abs(average)/b.fullBar*float64(track.Dx()))
		barColor := breakdownBarColors[role]
		if average < 0 {
			barColor = colorError
		}
		paint.FillShape(gtx.Ops, colorGridLine, clip.UniformRRect(track, radius).Op(gtx.Ops))
		paint.FillShape(gtx.Ops, barColor, clip.UniformRRect(bar, min(radius, bar.Dx()/2)).Op(gtx.Ops))
		points := textLabel(theme, smallTextSize, b.points[role][category], colorText)
		drawLabel(gtx, points, image.Pt(pointsRight, center), 1, 0.5)
		share := textLabel(theme, smallTextSize, b.shares[role][category], colorTextMuted)
		drawLabel(gtx, share, image.Pt(size.X-edge, center), 1, 0.5)
	}
	return layout.Dimensions{Size: size}
}
