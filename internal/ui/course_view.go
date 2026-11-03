package ui

import (
	"cmp"
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/cicadaclock/umango/internal/racesim"
)

const (
	coursePaddingX      = unit.Dp(14)
	coursePaddingTop    = unit.Dp(4)
	coursePaddingBottom = unit.Dp(2)
	courseTitleHeight   = unit.Dp(24)
	courseHillHeight    = unit.Dp(20)
	courseGrassHeight   = unit.Dp(2)
	courseRowHeight     = unit.Dp(22)
	courseRowBandHeight = unit.Dp(2)
	courseRulerHeight   = unit.Dp(16)
	courseAxisHeight    = unit.Dp(16)
	// Widths left and right of the strips, which the skill rows share so that both have one
	// distance scale
	courseLabelWidth = unit.Dp(176)
	courseValueWidth = unit.Dp(56)
	courseTickWidth  = unit.Dp(1)
	courseTickHeight = unit.Dp(4)
	courseLabelInset = unit.Dp(2)
	courseMarkerGap  = unit.Dp(4)
	courseTitleSize  = unit.Sp(14)
	courseMarkerSize = unit.Sp(10)

	// A course divides into 24 sections
	courseSectionCount = 24
	// The minor ticks of the distance axis fall every 100 m, its labels every 500 m
	courseMinorTickMeters = 100
	courseAxisLabelMeters = 500
	// Share of the hill area that the lowest and the highest point of a course reach
	courseHillFloor, courseHillCeiling = 0.45, 0.95
	// Height difference in meters that fills the hill area. Smaller differences use this scale,
	// so that a slight slope does not look like a steep one.
	courseMinHillRange = 4.0
	gradientPerMeter   = 10000
	// Lines of the course title beside the strips
	courseTitleMaxLines = 4
	// Index in courseLegBounds of the start of the late race, where the last spurt can begin
	courseLateRaceLeg = 2
	// Index in courseView.rows of the row of straights and corners
	courseSectionRow = 1
	// A corner counts down to corner 4, the last corner before each straight
	cornersPerLap = 4

	courseLayoutMissingFormat = "The course view requires this missing file: %s"
	courseLayoutAbsentFormat  = "The course layout file has no course %d"
)

// Numbering of race_course_set values
const (
	turnRight, turnLeft, turnStraight = 1, 2, 4
	inoutInner, inoutOuter, inoutBoth = 2, 3, 4
	groundDirt                        = 2
)

var (
	colorCourseHill  = rgb(0x7fa84f)
	colorCourseGrass = rgb(0x567a30)

	colorCourseSlopeBase    = rgb(0x322c3a)
	colorCourseSlopeBand    = rgb(0x5d4f6e)
	colorCourseUphill       = [2]color.NRGBA{rgb(0x9a8350), rgb(0x8b7546)}
	colorCourseUphillBand   = rgb(0x6d5a2d)
	colorCourseDownhill     = [2]color.NRGBA{rgb(0x3b8f85), rgb(0x337f76)}
	colorCourseDownhillBand = rgb(0x24615a)

	colorCourseSectionBase = rgb(0x2c2c31)
	colorCourseSectionBand = rgb(0x56565e)
	colorCourseStraight    = [2]color.NRGBA{rgb(0x3f6f98), rgb(0x386489)}
	colorCourseStraightBnd = rgb(0x2b4f6d)
	colorCourseCorner      = [2]color.NRGBA{rgb(0xa8704a), rgb(0x976541)}
	colorCourseCornerBand  = rgb(0x764a2c)

	colorCourseLegs = [...]struct{ fill, band color.NRGBA }{
		{rgb(0x2c8c6c), rgb(0x1e6a51)},
		{rgb(0x8d8533), rgb(0x6e6723)},
		{rgb(0x9c6687), rgb(0x774a66)},
		{rgb(0x85507a), rgb(0x643a5b)},
	}
	courseLegNames = [...]string{"Opening leg", "Middle leg", "Final leg", "Last spurt"}
	// Start of each leg as a fraction of the course distance, and its end
	courseLegBounds = [...]float64{0, 1.0 / 6, 2.0 / 3, 5.0 / 6, 1}

	colorCourseRuler     = rgb(0x22272d)
	colorCourseRulerMark = rgb(0x6f8796)
	colorCourseRulerText = rgb(0xa3b6c3)
	colorCourseMarker    = rgb(0xd8d8d8)
)

// courseSpan is a labeled stretch of a course row, in meters from the start
type courseSpan struct {
	start, end float64
	fill, band color.NRGBA
	// Label of the span, and the label that replaces it when the span is too narrow
	label, shortLabel string
}

// courseMarker is a distance label on a course row
type courseMarker struct {
	meters float64
	text   string
}

// courseRow is one horizontal strip of a course view
type courseRow struct {
	base, band color.NRGBA
	spans      []courseSpan
	markers    []courseMarker
	// Left and right of each label drawn in the current frame
	labelRanges [][2]int
}

// courseView holds the content of the loaded course, built outside the window goroutine. It draws
// the height profile, then the slopes, the straights and corners, the legs and the sections.
type courseView struct {
	title string
	// Why the course has no layout to draw, or empty
	note     string
	distance float64
	// Height profile as points in meters from the start and shares of the hill area
	hillMeters, hillShares []float64
	rows                   []courseRow
	sectionLabels          [courseSectionCount]string
	axisLabels             []courseMarker
}

// newCourseNote returns a view of a course with only its title and a note
func newCourseNote(title, note string, distance int) *courseView {
	return &courseView{title: title, note: note, distance: float64(distance)}
}

// newCourseView returns a view of a course with its layout
func newCourseView(title string, distance int, layout *racesim.CourseLayout) *courseView {
	v := &courseView{title: title, distance: float64(distance)}
	v.setHillProfile(layout.Slopes)
	v.rows = []courseRow{newSlopeRow(layout.Slopes), newSectionRow(layout), newLegRow(v.distance)}
	for i := range v.rows {
		v.rows[i].markers = rowMarkers(v.rows[i].spans, v.distance)
	}
	for i := range v.sectionLabels {
		v.sectionLabels[i] = strconv.Itoa(i + 1)
	}
	for meters := 0; meters <= distance; meters += courseAxisLabelMeters {
		v.axisLabels = append(v.axisLabels, courseMarker{float64(meters), groupedMeters(meters)})
	}
	return v
}

// courseTitle describes a course as its track, ground, distance, layout and turn direction
func courseTitle(trackName string, course *racesim.CourseSet) string {
	ground := "Turf"
	if course.Ground == groundDirt {
		ground = "Dirt"
	}
	title := fmt.Sprintf("%s %s %sm", trackName, ground, groupedMeters(course.Distance))
	switch course.Inout {
	case inoutInner:
		title += " (Inner)"
	case inoutOuter:
		title += " (Outer)"
	case inoutBoth:
		title += " (Outer to inner)"
	}
	switch course.Turn {
	case turnRight:
		title += " (Right)"
	case turnLeft:
		title += " (Left)"
	case turnStraight:
		title += " (Straight)"
	}
	return title
}

// groupedMeters returns a distance with its thousands separated
func groupedMeters(meters int) string {
	if meters < 1000 {
		return strconv.Itoa(meters)
	}
	return fmt.Sprintf("%d,%03d", meters/1000, meters%1000)
}

// setHillProfile accumulates the height of the course over its slopes and scales it to the hill
// area, keeping the same scale for every course that climbs less than courseMinHillRange
func (v *courseView) setHillProfile(slopes []racesim.CourseSlope) {
	ordered := slices.SortedFunc(slices.Values(slopes), func(a, b racesim.CourseSlope) int {
		return cmp.Compare(a.Start, b.Start)
	})
	v.hillMeters = append(v.hillMeters[:0], 0)
	heights := make([]float64, 1, 2*len(ordered)+2)
	height := 0.0
	for _, slope := range ordered {
		v.hillMeters = append(v.hillMeters, slope.Start)
		heights = append(heights, height)
		height += slope.Gradient / gradientPerMeter * (slope.End - slope.Start)
		v.hillMeters = append(v.hillMeters, slope.End)
		heights = append(heights, height)
	}
	v.hillMeters = append(v.hillMeters, v.distance)
	heights = append(heights, height)

	lowest, highest := slices.Min(heights), slices.Max(heights)
	scale := (courseHillCeiling - courseHillFloor) / max(highest-lowest, courseMinHillRange)
	v.hillShares = make([]float64, len(heights))
	for i, value := range heights {
		v.hillShares[i] = courseHillFloor + (value-lowest)*scale
	}
}

// newSlopeRow returns the row of uphill and downhill spans
func newSlopeRow(slopes []racesim.CourseSlope) courseRow {
	row := courseRow{base: colorCourseSlopeBase, band: colorCourseSlopeBand}
	var uphills, downhills int
	for _, slope := range slopes {
		span := courseSpan{start: slope.Start, end: slope.End}
		if slope.Gradient > 0 {
			span.fill, span.band = colorCourseUphill[uphills%2], colorCourseUphillBand
			span.label, span.shortLabel = "Uphill", "Up"
			uphills++
		} else {
			span.fill, span.band = colorCourseDownhill[downhills%2], colorCourseDownhillBand
			span.label, span.shortLabel = "Downhill", "Dn"
			downhills++
		}
		row.spans = append(row.spans, span)
	}
	sortSpans(row.spans)
	return row
}

// newSectionRow returns the row of straight and corner spans
func newSectionRow(layout *racesim.CourseLayout) courseRow {
	row := courseRow{base: colorCourseSectionBase, band: colorCourseSectionBand}
	for i, straight := range layout.Straights {
		row.spans = append(row.spans, courseSpan{
			start:      straight.Start,
			end:        straight.End,
			fill:       colorCourseStraight[i%2],
			band:       colorCourseStraightBnd,
			label:      "Straight",
			shortLabel: "St",
		})
	}
	for i, corner := range layout.Corners {
		number := strconv.Itoa(cornersPerLap - (len(layout.Corners)-1-i)%cornersPerLap)
		row.spans = append(row.spans, courseSpan{
			start:      corner.Start,
			end:        corner.End,
			fill:       colorCourseCorner[i%2],
			band:       colorCourseCornerBand,
			label:      "Corner " + number,
			shortLabel: "C" + number,
		})
	}
	sortSpans(row.spans)
	return row
}

// newLegRow returns the row of the four legs of a race
func newLegRow(distance float64) courseRow {
	row := courseRow{spans: make([]courseSpan, len(courseLegNames))}
	for i := range row.spans {
		row.spans[i] = courseSpan{
			start: courseLegBounds[i] * distance,
			end:   courseLegBounds[i+1] * distance,
			fill:  colorCourseLegs[i].fill,
			band:  colorCourseLegs[i].band,
			label: courseLegNames[i],
		}
	}
	return row
}

// sortSpans orders spans by their start
func sortSpans(spans []courseSpan) {
	slices.SortFunc(spans, func(a, b courseSpan) int { return cmp.Compare(a.start, b.start) })
}

// rowMarkers returns a distance marker for each span boundary inside the course
func rowMarkers(spans []courseSpan, distance float64) []courseMarker {
	var markers []courseMarker
	add := func(meters float64) {
		if meters <= 0 || meters >= distance {
			return
		}
		if len(markers) > 0 && markers[len(markers)-1].meters == meters {
			return
		}
		markers = append(markers, courseMarker{meters, strconv.Itoa(int(math.Round(meters))) + "m"})
	}
	for i := range spans {
		add(spans[i].start)
		add(spans[i].end)
	}
	return markers
}

// courseTrackBounds returns the left and right of the distance scale that the course strips and
// the skill rows share
func courseTrackBounds(gtx layout.Context) (left, right int) {
	left = gtx.Dp(courseLabelWidth)
	return left, max(gtx.Constraints.Max.X-gtx.Dp(courseValueWidth), left+1)
}

// sectionMarkers returns the boundaries of the straights and corners, none without a layout
func (v *courseView) sectionMarkers() []courseMarker {
	if v.note != "" {
		return nil
	}
	return v.rows[courseSectionRow].markers
}

// layout draws the title beside the course strips, or above the note when there is no layout
func (v *courseView) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	width := gtx.Constraints.Max.X
	border := gtx.Dp(borderWidth)
	fillRect(gtx, colorBorder, image.Rect(0, 0, width, border))
	y := border + gtx.Dp(coursePaddingTop)

	if v.note != "" {
		title, titleSize := recordCourseText(gtx, theme, courseTitleSize, v.title, colorText, font.SemiBold)
		titleHeight := gtx.Dp(courseTitleHeight)
		placeCourseText(gtx, title, image.Pt((width-titleSize.X)/2, y+(titleHeight-titleSize.Y)/2))
		y += titleHeight
		note, noteSize := recordCourseText(gtx, theme, smallTextSize, v.note, colorTextMuted, font.Normal)
		placeCourseText(gtx, note, image.Pt((width-noteSize.X)/2, y))
		return layout.Dimensions{Size: image.Pt(width, y+noteSize.Y+gtx.Dp(coursePaddingBottom))}
	}

	left, right := courseTrackBounds(gtx)
	offset := op.Offset(image.Pt(left, y)).Push(gtx.Ops)
	stripHeight := v.layoutStrips(gtx, theme, right-left)
	offset.Pop()

	paddingX := gtx.Dp(coursePaddingX)
	title := textLabel(theme, textSize, v.title, colorText)
	title.Font.Weight = font.SemiBold
	title.MaxLines = courseTitleMaxLines
	title.WrapPolicy = text.WrapWords
	titleGtx := gtx
	titleGtx.Constraints = layout.Constraints{Max: image.Pt(left-2*paddingX, stripHeight)}
	macro := op.Record(gtx.Ops)
	titleDims := title.Layout(titleGtx)
	titleCall := macro.Stop()
	placeCourseText(gtx, titleCall, image.Pt(paddingX, y+(stripHeight-titleDims.Size.Y)/2))
	return layout.Dimensions{Size: image.Pt(width, y+stripHeight+gtx.Dp(coursePaddingBottom))}
}

// layoutStrips draws the hill area, the rows, the section ruler and the distance axis at the
// current origin, and returns their height
func (v *courseView) layoutStrips(gtx layout.Context, theme *material.Theme, width int) int {
	scale := float64(width) / v.distance
	y := v.layoutHills(gtx, width, scale)
	rowHeight := gtx.Dp(courseRowHeight)
	for i := range v.rows {
		offset := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
		v.rows[i].layout(gtx, theme, width, rowHeight, scale)
		offset.Pop()
		y += rowHeight
	}
	offset := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
	y += v.layoutRuler(gtx, theme, width)
	offset.Pop()
	offset = op.Offset(image.Pt(0, y)).Push(gtx.Ops)
	y += v.layoutAxis(gtx, theme, width, scale)
	offset.Pop()
	return y
}

// layoutHills fills the height profile above a grass line and returns their height
func (v *courseView) layoutHills(gtx layout.Context, width int, scale float64) int {
	height := gtx.Dp(courseHillHeight)
	bottom := float32(height)
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(0, bottom))
	for i, meters := range v.hillMeters {
		path.LineTo(f32.Pt(float32(meters*scale), bottom*(1-float32(v.hillShares[i]))))
	}
	path.LineTo(f32.Pt(float32(width), bottom))
	path.Close()
	paint.FillShape(gtx.Ops, colorCourseHill, clip.Outline{Path: path.End()}.Op())
	grass := gtx.Dp(courseGrassHeight)
	fillRect(gtx, colorCourseGrass, image.Rect(0, height, width, height+grass))
	return height + grass
}

// layout draws the row base, its spans with their labels, and its markers
func (row *courseRow) layout(
	gtx layout.Context,
	theme *material.Theme,
	width, height int,
	scale float64,
) {
	band := gtx.Dp(courseRowBandHeight)
	if row.base != (color.NRGBA{}) {
		fillRect(gtx, row.base, image.Rect(0, 0, width, height-band))
		fillRect(gtx, row.band, image.Rect(0, height-band, width, height))
	}
	inset := gtx.Dp(courseLabelInset)
	row.labelRanges = row.labelRanges[:0]
	for i := range row.spans {
		span := &row.spans[i]
		left, right := int(span.start*scale), int(span.end*scale)
		fillRect(gtx, span.fill, image.Rect(left, 0, right, height-band))
		fillRect(gtx, span.band, image.Rect(left, height-band, right, height))
		call, size := recordCourseText(gtx, theme, smallTextSize, span.label, colorTextBright, font.SemiBold)
		if size.X+2*inset > right-left && span.shortLabel != "" {
			call, size = recordCourseText(gtx, theme, smallTextSize, span.shortLabel, colorTextBright, font.SemiBold)
		}
		if size.X+2*inset <= right-left {
			x := (left + right - size.X) / 2
			placeCourseText(gtx, call, image.Pt(x, inset))
			row.labelRanges = append(row.labelRanges, [2]int{x, x + size.X})
		}
	}
	row.layoutMarkers(gtx, theme, width, height-band, scale)
}

// layoutMarkers draws the distance markers along the bottom of the row, leaving out each marker
// that would touch the one before it or a span label
func (row *courseRow) layoutMarkers(
	gtx layout.Context,
	theme *material.Theme,
	width, bottom int,
	scale float64,
) {
	gap := gtx.Dp(courseMarkerGap)
	nextFree := 0
	for i := range row.markers {
		marker := &row.markers[i]
		call, size := recordCourseText(gtx, theme, courseMarkerSize, marker.text, colorCourseMarker, font.Normal)
		x := max(min(int(marker.meters*scale)-size.X/2, width-size.X), 0)
		if x < nextFree || row.touchesLabel(x-gap, x+size.X+gap) {
			continue
		}
		placeCourseText(gtx, call, image.Pt(x, bottom-size.Y))
		nextFree = x + size.X + gap
	}
}

// touchesLabel reports whether a horizontal range overlaps a label drawn in the current frame
func (row *courseRow) touchesLabel(left, right int) bool {
	for _, label := range row.labelRanges {
		if left < label[1] && label[0] < right {
			return true
		}
	}
	return false
}

// layoutRuler draws the numbered sections and returns its height
func (v *courseView) layoutRuler(gtx layout.Context, theme *material.Theme, width int) int {
	height := gtx.Dp(courseRulerHeight)
	fillRect(gtx, colorCourseRuler, image.Rect(0, 0, width, height))
	tick, tickHeight := gtx.Dp(courseTickWidth), gtx.Dp(courseTickHeight)
	for i := range courseSectionCount + 1 {
		x := min(i*width/courseSectionCount, width-tick)
		fillRect(gtx, colorCourseRulerMark, image.Rect(x, height-tickHeight, x+tick, height))
	}
	for i := range v.sectionLabels {
		call, size := recordCourseText(gtx, theme, courseMarkerSize, v.sectionLabels[i], colorCourseRulerText, font.Normal)
		center := (2*i + 1) * width / (2 * courseSectionCount)
		placeCourseText(gtx, call, image.Pt(center-size.X/2, (height-tickHeight-size.Y)/2))
	}
	fillRect(gtx, colorCourseRulerMark, image.Rect(0, height-tick, width, height))
	return height
}

// layoutAxis draws the distance ticks and their labels and returns its height
func (v *courseView) layoutAxis(gtx layout.Context, theme *material.Theme, width int, scale float64) int {
	tick, tickHeight := gtx.Dp(courseTickWidth), gtx.Dp(courseTickHeight)
	for meters := 0.0; meters <= v.distance; meters += courseMinorTickMeters {
		x := min(int(meters*scale), width-tick)
		fillRect(gtx, colorCourseRulerMark, image.Rect(x, 0, x+tick, tickHeight))
	}
	for i := range v.axisLabels {
		label := &v.axisLabels[i]
		call, size := recordCourseText(gtx, theme, courseMarkerSize, label.text, colorTextMuted, font.Normal)
		x := max(min(int(label.meters*scale)-size.X/2, width-size.X), 0)
		placeCourseText(gtx, call, image.Pt(x, tickHeight))
	}
	return gtx.Dp(courseAxisHeight)
}

// recordCourseText records a single line of text at any width, for drawing where its size decides
func recordCourseText(
	gtx layout.Context,
	theme *material.Theme,
	size unit.Sp,
	s string,
	c color.NRGBA,
	weight font.Weight,
) (op.CallOp, image.Point) {
	gtx.Constraints = layout.Constraints{Max: image.Pt(unboundedTextWidth, gtx.Constraints.Max.Y)}
	label := textLabel(theme, size, s, c)
	label.Font.Weight = weight
	label.Alignment = text.Start
	macro := op.Record(gtx.Ops)
	dims := label.Layout(gtx)
	return macro.Stop(), dims.Size
}

// placeCourseText draws recorded text with its top left corner at a point
func placeCourseText(gtx layout.Context, call op.CallOp, at image.Point) {
	defer op.Offset(at).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
}
