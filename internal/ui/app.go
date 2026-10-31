// Package ui draws the Umango window with the Gio immediate mode toolkit.
package ui

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/logo"
	"github.com/cicadaclock/umango/internal/races"
	fontapi "github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"golang.org/x/sync/errgroup"
)

const (
	fontPath                        = "assets/font/Inter-VariableFont_opsz,wght.ttf"
	windowWidth, windowHeight       = unit.Dp(1280), unit.Dp(720)
	minWindowWidth, minWindowHeight = unit.Dp(960), unit.Dp(600)
	topBarHeight                    = unit.Dp(40)
	logoSize                        = unit.Dp(26)
	// Space between the logo and the wordmark
	logoGap      = unit.Dp(6)
	wordmarkSize = unit.Sp(16)
	wordmarkText = "UMANGO"
	// Cap height of Inter in ems
	interCapHeight = 0.727
	// Space between the wordmark and the first page tab
	wordmarkGap                   = unit.Dp(18)
	tabPadding                    = unit.Dp(10)
	tabIndicatorHeight            = unit.Dp(2)
	loadingIndicatorSize          = unit.Dp(28)
	loadingTextGap                = unit.Dp(14)
	tabTeamTrials, tabVeteranList = 0, 1
	tabCount                      = 2
	loadingText                   = "Loading master.mdb and saved races…"
	loadingErrorPrefix            = "Failed to load master.mdb:\n"
)

var (
	tabTitles = [tabCount]string{tabTeamTrials: "TT Chart", tabVeteranList: "Veteran List"}

	// Weight instances of the Inter variable font that the full shaper adds to the regular face
	interInstances = [...]struct {
		weight font.Weight
		wght   float32
	}{
		{font.SemiBold, 600},
		{font.ExtraBold, 800},
	}
	// The wordmark's gradient runs from this amber into the accent
	colorWordmarkStart = rgb(0xffa630)
)

// App opens the Umango window and loads master.mdb and the saved races in the background.
//
// App must be called from the main goroutine, as it hands the main thread to the window system.
// Once the window opens App never returns: the process exits when the window closes, or logs
// fatally when the window fails.
//
// Inputs:
// - assets: Embedded file system holding the Inter font
//
// Returns:
// - err: Failure reading or parsing the font, before the window opens
func App(assets embed.FS) error {
	fontData, err := assets.ReadFile(fontPath)
	if err != nil {
		return fmt.Errorf("read assets: %w", err)
	}
	shaper, err := newShaper(fontData, false)
	if err != nil {
		return err
	}
	go func() {
		if err := run(fontData, shaper); err != nil {
			log.Fatalf("run window: %v", err)
		}
		os.Exit(0)
	}()
	app.Main()
	return nil
}

// newShaper parses fontData, the Inter variable font, into a text shaper
//
// The full shaper adds semibold and extra bold instances of Inter and falls back to the system
// fonts for glyphs Inter lacks. Each costs milliseconds, so the first frames use a regular-only
// shaper while the loader builds the full one in the background.
func newShaper(fontData []byte, full bool) (*text.Shaper, error) {
	regular, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	faces := []font.FontFace{{Font: regular.Font(), Face: regular}}
	if !full {
		return text.NewShaper(text.NoSystemFonts(), text.WithCollection(faces)), nil
	}
	for _, instance := range interInstances {
		// Gio identifies faces by their parsed font, so each instance needs its own parse
		face, err := opentype.Parse(fontData)
		if err != nil {
			return nil, fmt.Errorf("parse font: %w", err)
		}
		instanceFont := face.Font()
		instanceFont.Weight = instance.weight
		faces = append(faces, font.FontFace{Font: instanceFont, Face: variedFace{
			face:       face,
			variations: []fontapi.Variation{{Tag: ot.MustNewTag("wght"), Value: instance.wght}},
		}})
	}
	return text.NewShaper(text.WithCollection(faces)), nil
}

// variedFace serves a variable font at fixed variation axis values
type variedFace struct {
	face       opentype.Face
	variations []fontapi.Variation
}

// Face returns a face of the font at the variation axis values, for use by a single goroutine
func (f variedFace) Face() *fontapi.Face {
	face := f.face.Face()
	face.SetVariations(f.variations)
	return face
}

// loadResult carries the background load into the window goroutine
type loadResult struct {
	shaper     *text.Shaper
	dataStore  *data.DataStore
	teamTrials *teamTrialsPage
	// Directory the saved races loaded from
	resultsDir string
	err        error
}

// load reads master.mdb and the saved races concurrently and aggregates them into the TT page
func load(fontData []byte) loadResult {
	var (
		result    loadResult
		resultSet races.TeamTrialResultSet
	)
	var g errgroup.Group
	g.Go(func() (err error) { result.shaper, err = newShaper(fontData, true); return })
	g.Go(func() (err error) { result.dataStore, err = data.Init(); return })
	g.Go(func() (err error) { result.resultsDir, resultSet, err = loadTeamTrialResults(); return })
	if err := g.Wait(); err != nil {
		return loadResult{err: err}
	}
	result.teamTrials = newTeamTrialsPage(result.dataStore, resultSet)
	return result
}

// screen holds the window state between frames
type screen struct {
	theme      *material.Theme
	window     *app.Window
	explorer   *explorer.Explorer
	loadError  string
	tabs       [tabCount]widget.Clickable
	activeTab  int
	teamTrials *teamTrialsPage
	veterans   *veteranPage
	// Reloads of the TT page's data from the results watcher, at most one waiting
	reloads chan teamTrialsData
}

// run drives the window event loop until the window closes
func run(fontData []byte, shaper *text.Shaper) error {
	var window app.Window
	window.Option(
		app.Title("Umango"),
		app.Size(windowWidth, windowHeight),
		app.MinSize(minWindowWidth, minWindowHeight),
	)
	s := screen{
		theme:    newTheme(shaper),
		window:   &window,
		explorer: explorer.NewExplorer(&window),
		reloads:  make(chan teamTrialsData, 1),
	}

	loads := make(chan loadResult, 1)
	go func() {
		loads <- load(fontData)
		window.Invalidate()
	}()

	var ops op.Ops
	for {
		e := window.Event()
		s.explorer.ListenEvents(e)
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			select {
			case result := <-loads:
				s.applyLoad(result)
			case loaded := <-s.reloads:
				s.teamTrials.applyData(loaded)
			default:
			}
			gtx := app.NewContext(&ops, e)
			s.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

// applyLoad installs the background load results and starts watching for new races, or shows
// the load's error message on failure
func (s *screen) applyLoad(result loadResult) {
	if result.err != nil {
		log.Printf("load data: %v", result.err)
		s.loadError = loadingErrorPrefix + result.err.Error()
		return
	}
	s.theme.Shaper = result.shaper
	s.teamTrials = result.teamTrials
	s.veterans = newVeteranPage(result.dataStore)
	go watchTeamTrialResults(result.resultsDir, result.dataStore, s.deliverReload)
}

// deliverReload hands the window goroutine a reload of the TT page's data, replacing one it has
// not applied yet, and wakes it to apply it
func (s *screen) deliverReload(loaded teamTrialsData) {
	select {
	case <-s.reloads:
	default:
	}
	s.reloads <- loaded
	s.window.Invalidate()
}

func (s *screen) layout(gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, colorWindow)
	switch {
	case s.loadError != "":
		return layoutCenteredText(gtx, s.theme, s.loadError, colorError)
	case s.teamTrials == nil:
		return layoutLoading(gtx, s.theme)
	}
	for i := range s.tabs {
		if s.tabs[i].Clicked(gtx) {
			s.activeTab = i
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(s.layoutTopBar),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(gutterWidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if s.activeTab == tabVeteranList {
					return s.veterans.layout(gtx, s.theme, s.explorer, s.window.Invalidate)
				}
				return s.teamTrials.layout(gtx, s.theme)
			})
		}),
	)
}

// layoutTopBar draws the logo, the wordmark and the page tabs left to right above a divider
func (s *screen) layoutTopBar(gtx layout.Context) layout.Dimensions {
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(topBarHeight))
	gtx.Constraints = layout.Constraints{Max: size}
	line := gtx.Dp(borderWidth)
	fillRect(gtx, colorBorder, image.Rect(0, size.Y-line, size.X, size.Y))

	x, logoSide := gtx.Dp(gutterWidth+panelTabPadding), gtx.Dp(logoSize)
	offset := op.Offset(image.Pt(x, (size.Y-logoSide)/2)).Push(gtx.Ops)
	logo.Draw(gtx.Ops, float32(logoSide))
	offset.Pop()
	x += logoSide + gtx.Dp(logoGap)
	offset = op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
	x += layoutWordmark(gtx, s.theme).Size.X + gtx.Dp(wordmarkGap)
	offset.Pop()
	for i := range s.tabs {
		offset := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
		x += s.layoutPageTab(gtx, i).Size.X
		offset.Pop()
	}
	return layout.Dimensions{Size: size}
}

// layoutWordmark draws the application name in heavy capitals shaded by the accent gradient,
// its capitals centered in the height of the maximum constraints
func layoutWordmark(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	height := gtx.Constraints.Max.Y
	gtx.Constraints.Min = image.Point{}
	wordmark := widget.Label{MaxLines: 1}
	face := font.Font{Typeface: theme.Face, Weight: font.ExtraBold}
	// The gradient spans the measured width of the text
	macro := op.Record(gtx.Ops)
	dims := wordmark.Layout(gtx, theme.Shaper, face, wordmarkSize, wordmarkText, op.CallOp{})
	macro.Stop()
	macro = op.Record(gtx.Ops)
	paint.LinearGradientOp{
		Stop1:  f32.Pt(0, 0),
		Color1: colorWordmarkStart,
		Stop2:  f32.Pt(float32(dims.Size.X), 0),
		Color2: colorAccent,
	}.Add(gtx.Ops)
	gradient := macro.Stop()

	capHeight := int(interCapHeight * float32(gtx.Sp(wordmarkSize)))
	ascent := dims.Size.Y - dims.Baseline
	defer op.Offset(image.Pt(0, (height+capHeight)/2-ascent)).Push(gtx.Ops).Pop()
	wordmark.Layout(gtx, theme.Shaper, face, wordmarkSize, wordmarkText, gradient)
	return layout.Dimensions{Size: image.Pt(dims.Size.X, height)}
}

// layoutPageTab draws page i's title, brightened while hovered or active, and underlines the
// active page in the accent
func (s *screen) layoutPageTab(gtx layout.Context, i int) layout.Dimensions {
	tab := &s.tabs[i]
	return tab.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		active := i == s.activeTab
		c := colorTextMuted
		if active || tab.Hovered() {
			c = colorText
		}
		label := textLabel(s.theme, textSize, tabTitles[i], c)
		label.Font.Weight = font.SemiBold
		height, padding := gtx.Constraints.Max.Y, gtx.Dp(tabPadding)
		gtx.Constraints.Min = image.Point{}
		macro := op.Record(gtx.Ops)
		dims := label.Layout(gtx)
		call := macro.Stop()

		size := image.Pt(dims.Size.X+2*padding, height)
		offset := op.Offset(image.Pt(padding, (height-dims.Size.Y)/2)).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
		if active {
			indicator := image.Rect(padding, height-gtx.Dp(tabIndicatorHeight), size.X-padding, height)
			fillRect(gtx, colorAccent, indicator)
		}
		defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
		pointer.CursorPointer.Add(gtx.Ops)
		return layout.Dimensions{Size: size}
	})
}

func layoutLoading(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				size := gtx.Dp(loadingIndicatorSize)
				gtx.Constraints = layout.Exact(image.Pt(size, size))
				return material.Loader(theme).Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: loadingTextGap}.Layout),
			layout.Rigid(textLabel(theme, smallTextSize, loadingText, colorTextMuted).Layout),
		)
	})
}

// layoutCenteredText draws s centered in the constraints, wrapping at line breaks
func layoutCenteredText(
	gtx layout.Context,
	theme *material.Theme,
	s string,
	c color.NRGBA,
) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(theme, textSize, s)
		label.Color = c
		label.Alignment = text.Middle
		return label.Layout(gtx)
	})
}
