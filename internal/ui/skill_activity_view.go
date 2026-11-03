package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/racesim"
)

const (
	skillStripHeight   = unit.Dp(18)
	skillBarInset      = unit.Dp(4)
	skillHeaderHeight  = unit.Dp(22)
	skillNoteHeight    = unit.Dp(18)
	skillNotePaddingY  = unit.Dp(3)
	skillKeyWidth      = unit.Dp(96)
	skillKeyHeight     = unit.Dp(8)
	skillKeyGap        = unit.Dp(6)
	skillSpurtLineSize = unit.Dp(1)
	// Texels of the heat key
	skillKeyTexels = 64
	// Bins on each side of a bin that smooth its heat
	skillHeatSmoothingRadius = 2
	// Height left for the skill list when a runner is selected and the window is short
	skillListMinimumHeight = skillHeaderHeight + 4*skillStripHeight

	skillPanelTitle       = "Skills"
	skillFiresInText      = "Rate"
	skillKeyLowText       = "1%"
	skillKeyHighText      = "100%"
	skillKeyText          = "of races have the skill active at that point"
	skillAlwaysText       = "Always active"
	skillNeverFiredFormat = "%d learned %s never fired."
	selectRunnerText      = "Select a row to show where its skills activate"
)

var (
	// Stops of the heat ramp, from the fewest races to the most
	skillHeatStops = [...]color.NRGBA{
		{R: 44, G: 17, B: 72, A: 217},
		{R: 120, G: 35, B: 128, A: 242},
		{R: 205, G: 64, B: 110, A: 255},
		{R: 246, G: 138, B: 88, A: 255},
		{R: 253, G: 231, B: 169, A: 255},
	}
	// Color of each count of races, transparent at 0
	skillHeatColors    = newSkillHeatColors()
	colorSkillBarLane  = rgb(0x232323)
	colorSkillSpurtBar = rgb(0xe8ad63)
)

// skillActivityRow is one skill strip of a skill activity view
type skillActivityRow struct {
	name string
	// Share of races in which the skill activated, as text
	firesIn string
	heat    paint.ImageOp
}

// skillActivityView holds the skills of one runner over the simulated races, built outside the
// window goroutine. It follows the skill rows of the race overview of uma.moe.
type skillActivityView struct {
	runnerName string
	rows       []skillActivityRow
	// Names of the skills without a time limit
	alwaysActive string
	// States how many learned skills never activated, or empty
	neverFired string
}

// newSkillActivityView returns the view of the skill activity of a runner
func newSkillActivityView(
	dataStore *data.DataStore,
	runner *racesim.Runner,
	activities []racesim.SkillActivity,
) skillActivityView {
	view := skillActivityView{runnerName: runner.Name}
	ids := make([]int, len(activities))
	fired := make(map[int]struct{}, len(activities))
	for i := range activities {
		ids[i] = activities[i].SkillId
		fired[ids[i]] = struct{}{}
	}
	names := dataStore.SkillNames(ids)
	var always []string
	for i := range activities {
		if activities[i].Passive {
			always = append(always, names[i])
			continue
		}
		view.rows = append(view.rows, skillActivityRow{
			name:    names[i],
			firesIn: fmt.Sprintf("%.0f%%", float64(activities[i].RaceCount)/racesPerTrial*100),
			heat:    newSkillHeatImage(&activities[i]),
		})
	}
	view.alwaysActive = strings.Join(always, ",  ")

	silent := make(map[int]struct{})
	for _, id := range runner.SkillIds {
		if _, ok := fired[id]; !ok {
			silent[id] = struct{}{}
		}
	}
	if count := len(silent); count > 0 {
		noun := "skills"
		if count == 1 {
			noun = "skill"
		}
		view.neverFired = fmt.Sprintf(skillNeverFiredFormat, count, noun)
	}
	return view
}

// newSkillHeatColors returns the color of each count of races. A share of 1% of races takes the
// start of the ramp and all races take its end, on a logarithmic scale, so a low overlap stays
// visible and each further overlap is hotter.
func newSkillHeatColors() [racesPerTrial + 1]color.RGBA {
	var colors [racesPerTrial + 1]color.RGBA
	for count := 1; count <= racesPerTrial; count++ {
		colors[count] = skillHeatRampColor(math.Log(float64(count)) / math.Log(racesPerTrial))
	}
	return colors
}

// skillHeatRampColor returns the alpha premultiplied color at a position from 0 to 1 of the ramp
func skillHeatRampColor(position float64) color.RGBA {
	scaled := position * float64(len(skillHeatStops)-1)
	stop := min(int(scaled), len(skillHeatStops)-2)
	share := scaled - float64(stop)
	from, to := skillHeatStops[stop], skillHeatStops[stop+1]
	mix := func(a, b uint8) uint8 { return uint8(float64(a) + (float64(b)-float64(a))*share) }
	return color.RGBAModel.Convert(color.NRGBA{
		R: mix(from.R, to.R),
		G: mix(from.G, to.G),
		B: mix(from.B, to.B),
		A: mix(from.A, to.A),
	}).(color.RGBA)
}

// newSkillHeatImage returns the heat profile of a skill as an image with one texel per bin. Each
// bin takes the mean of its neighbors weighted by nearness, rounded up, so that the scattered
// activations of a short skill read as a faint stretch and a single activation stays visible.
func newSkillHeatImage(activity *racesim.SkillActivity) paint.ImageOp {
	img := image.NewRGBA(image.Rect(0, 0, racesim.SkillHeatBins, 1))
	for bin := range activity.Heat {
		sum, weightSum := 0, 0
		first := max(bin-skillHeatSmoothingRadius, 0)
		last := min(bin+skillHeatSmoothingRadius, racesim.SkillHeatBins-1)
		for neighbor := first; neighbor <= last; neighbor++ {
			distance := bin - neighbor
			if distance < 0 {
				distance = -distance
			}
			weight := skillHeatSmoothingRadius + 1 - distance
			sum += weight * int(activity.Heat[neighbor])
			weightSum += weight
		}
		img.SetRGBA(bin, 0, skillHeatColors[(sum+weightSum-1)/weightSum])
	}
	return paint.NewImageOp(img)
}

// newSkillKeyImage returns the heat ramp as an image
func newSkillKeyImage() paint.ImageOp {
	img := image.NewRGBA(image.Rect(0, 0, skillKeyTexels, 1))
	for x := range skillKeyTexels {
		img.SetRGBA(x, 0, skillHeatRampColor(float64(x)/(skillKeyTexels-1)))
	}
	return paint.NewImageOp(img)
}

// layoutStretchedImage fills a rectangle with an image, with rounded ends
func layoutStretchedImage(gtx layout.Context, img paint.ImageOp, bounds image.Rectangle) {
	radius := bounds.Dy() / 2
	defer clip.UniformRRect(bounds, radius).Push(gtx.Ops).Pop()
	size := img.Size()
	scale := f32.Pt(float32(bounds.Dx())/float32(size.X), float32(bounds.Dy())/float32(size.Y))
	transform := f32.Affine2D{}.Scale(f32.Point{}, scale).Offset(layout.FPt(bounds.Min))
	defer op.Affine(transform).Push(gtx.Ops).Pop()
	img.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

// layout draws the header above the scrolling skill rows
func (v *skillActivityView) layout(
	gtx layout.Context,
	theme *material.Theme,
	list *widget.List,
	course *courseView,
	key paint.ImageOp,
) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return v.layoutHeader(gtx, theme, key)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			style := material.List(theme, list)
			style.AnchorStrategy = material.Overlay
			styleScrollbar(&style.ScrollbarStyle)
			return style.Layout(gtx, len(v.rows)+2, func(gtx layout.Context, index int) layout.Dimensions {
				switch {
				case index < len(v.rows):
					return v.rows[index].layout(gtx, theme, course)
				case index == len(v.rows):
					return layoutSkillNote(gtx, theme, skillAlwaysText, v.alwaysActive)
				}
				return layoutSkillNote(gtx, theme, "", v.neverFired)
			})
		}),
	)
}

// layoutHeader draws the runner, the key of the heat ramp and the title of the rate column
func (v *skillActivityView) layoutHeader(
	gtx layout.Context,
	theme *material.Theme,
	key paint.ImageOp,
) layout.Dimensions {
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(skillHeaderHeight))
	left, right := courseTrackBounds(gtx)
	line := gtx.Dp(borderWidth)
	fillRect(gtx, colorBorder, image.Rect(0, 0, size.X, line))
	fillRect(gtx, colorBorder, image.Rect(0, size.Y-line, size.X, size.Y))

	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: image.Pt(left-gtx.Dp(coursePaddingX)-gtx.Dp(skillKeyGap), size.Y)}
	title := textLabel(theme, smallTextSize, skillPanelTitle+"  "+v.runnerName, colorText)
	title.Font.Weight = font.SemiBold
	macro := op.Record(gtx.Ops)
	titleDims := title.Layout(textGtx)
	titleCall := macro.Stop()
	placeCourseText(gtx, titleCall, image.Pt(gtx.Dp(coursePaddingX), (size.Y-titleDims.Size.Y)/2))

	gap := gtx.Dp(skillKeyGap)
	x := left
	place := func(s string) {
		call, textSize := recordCourseText(gtx, theme, courseMarkerSize, s, colorTextMuted, font.Normal)
		placeCourseText(gtx, call, image.Pt(x, (size.Y-textSize.Y)/2))
		x += textSize.X + gap
	}
	place(skillKeyLowText)
	keyWidth, keyHeight := gtx.Dp(skillKeyWidth), gtx.Dp(skillKeyHeight)
	keyTop := (size.Y - keyHeight) / 2
	layoutStretchedImage(gtx, key, image.Rect(x, keyTop, x+keyWidth, keyTop+keyHeight))
	x += keyWidth + gap
	place(skillKeyHighText)
	place(skillKeyText)

	call, textSize := recordCourseText(gtx, theme, courseMarkerSize, skillFiresInText, colorTextMuted, font.Normal)
	placeCourseText(gtx, call, image.Pt(right+gtx.Dp(courseValueWidth)-gtx.Dp(coursePaddingX)-textSize.X, (size.Y-textSize.Y)/2))
	return layout.Dimensions{Size: size}
}

// layoutSkillNote draws a line of text under the skill rows, or nothing when the text is empty
func layoutSkillNote(gtx layout.Context, theme *material.Theme, heading, s string) layout.Dimensions {
	if s == "" {
		return layout.Dimensions{}
	}
	left, right := courseTrackBounds(gtx)
	paddingY := gtx.Dp(skillNotePaddingY)
	if heading != "" {
		call, _ := recordCourseText(gtx, theme, smallTextSize, heading, colorTextMuted, font.Normal)
		placeCourseText(gtx, call, image.Pt(gtx.Dp(coursePaddingX), paddingY))
	}
	label := textLabel(theme, smallTextSize, s, colorText)
	label.MaxLines = 0
	label.WrapPolicy = text.WrapWords
	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: image.Pt(right-left, gtx.Constraints.Max.Y)}
	macro := op.Record(gtx.Ops)
	dims := label.Layout(textGtx)
	call := macro.Stop()
	placeCourseText(gtx, call, image.Pt(left, paddingY))
	height := max(dims.Size.Y+2*paddingY, gtx.Dp(skillNoteHeight))
	return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, height)}
}

// layout draws the skill name, its strip on the distance scale of the course, and its rate
func (row *skillActivityRow) layout(
	gtx layout.Context,
	theme *material.Theme,
	course *courseView,
) layout.Dimensions {
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(skillStripHeight))
	left, right := courseTrackBounds(gtx)
	paddingX := gtx.Dp(coursePaddingX)

	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: image.Pt(left-paddingX-gtx.Dp(skillKeyGap), size.Y)}
	macro := op.Record(gtx.Ops)
	nameDims := textLabel(theme, smallTextSize, row.name, colorTextMuted).Layout(textGtx)
	nameCall := macro.Stop()
	placeCourseText(gtx, nameCall, image.Pt(paddingX, (size.Y-nameDims.Size.Y)/2))

	line := gtx.Dp(borderWidth)
	fillRect(gtx, colorColumnRule, image.Rect(left, size.Y-line, right, size.Y))
	scale := float64(right-left) / course.distance
	for _, section := range course.sectionMarkers() {
		x := left + int(section.meters*scale)
		fillRect(gtx, colorBorder, image.Rect(x, 0, x+line, size.Y))
	}
	inset := gtx.Dp(skillBarInset)
	bar := image.Rect(left, inset, right, size.Y-inset)
	paint.FillShape(gtx.Ops, colorSkillBarLane, clip.UniformRRect(bar, bar.Dy()/2).Op(gtx.Ops))
	layoutStretchedImage(gtx, row.heat, bar)
	spurt := left + int(courseLegBounds[courseLateRaceLeg]*course.distance*scale)
	fillRect(gtx, colorSkillSpurtBar, image.Rect(spurt, 0, spurt+gtx.Dp(skillSpurtLineSize), size.Y))

	call, textSize := recordCourseText(gtx, theme, courseMarkerSize, row.firesIn, colorTextMuted, font.Normal)
	placeCourseText(gtx, call, image.Pt(right+gtx.Dp(courseValueWidth)-paddingX-textSize.X, (size.Y-textSize.Y)/2))
	return layout.Dimensions{Size: size}
}
