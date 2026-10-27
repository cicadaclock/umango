// Package ui draws the Umango window with the Gio immediate mode toolkit.
package ui

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/opentype"
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
	"github.com/cicadaclock/umango/internal/races"
	fontapi "github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"golang.org/x/sync/errgroup"
)

const (
	fontPath                      = "assets/font/Inter-VariableFont_opsz,wght.ttf"
	interExtraBoldWeight          = 800
	windowWidth, windowHeight     = unit.Dp(1280), unit.Dp(720)
	textSize                      = unit.Sp(14)
	tabPadding                    = unit.Dp(10)
	tabIndicatorHeight            = unit.Dp(2)
	loadingIndicatorSize          = unit.Dp(32)
	tabTeamTrials, tabVeteranList = 0, 1
	tabCount                      = 2
	loadingText                   = "Loading master.mdb and saved races..."
	loadingErrorPrefix            = "Failed to load master.mdb:\n"
)

var (
	colorBackground = color.NRGBA{R: 0xdd, G: 0xdd, B: 0xdd, A: 0xff}
	colorForeground = color.NRGBA{R: 0x56, G: 0x56, B: 0x56, A: 0xff}
	colorPrimary    = color.NRGBA{R: 0x29, G: 0x6f, B: 0xf6, A: 0xff}
	colorOnPrimary  = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	colorSelection  = color.NRGBA{R: 0x29, G: 0x6f, B: 0xf6, A: 0x40}
	colorHover      = color.NRGBA{R: 0x29, G: 0x6f, B: 0xf6, A: 0x1a}
	colorHeader     = color.NRGBA{R: 0xf5, G: 0xf5, B: 0xf5, A: 0xff}
	colorSeparator  = color.NRGBA{R: 0xe3, G: 0xe3, B: 0xe3, A: 0xff}
	colorDivider    = color.NRGBA{R: 0xc4, G: 0xc4, B: 0xc4, A: 0xff}
	colorError      = color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff}

	tabTitles = [tabCount]string{tabTeamTrials: "TT Chart", tabVeteranList: "Veteran List"}
)

// App opens the Umango window and loads master.mdb and the saved races in the background.
//
// App must be called from the main goroutine, as it hands the main thread to the window system.
// Once the window opens App never returns: the process exits when the window closes, or logs
// fatally when the window fails.
//
// Inputs:
// - assets: Embedded file system holding the Inter font and the extracted game icons
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
		if err := run(assets, fontData, shaper); err != nil {
			log.Fatalf("run window: %v", err)
		}
		os.Exit(0)
	}()
	app.Main()
	return nil
}

// newShaper parses fontData, the Inter variable font, into a text shaper
//
// The full shaper adds an extra bold instance of Inter and falls back to the system fonts for
// glyphs Inter lacks. Each costs milliseconds, so the first frames use a regular-only shaper
// while the loader builds the full one in the background.
func newShaper(fontData []byte, full bool) (*text.Shaper, error) {
	regular, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	faces := []font.FontFace{{Font: regular.Font(), Face: regular}}
	if !full {
		return text.NewShaper(text.NoSystemFonts(), text.WithCollection(faces)), nil
	}
	// Gio identifies faces by their parsed font, so the bold instance needs its own parse
	bold, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	boldFont := bold.Font()
	boldFont.Weight = font.ExtraBold
	boldFace := variedFace{face: bold, variations: []fontapi.Variation{
		{Tag: ot.MustNewTag("wght"), Value: interExtraBoldWeight},
	}}
	faces = append(faces, font.FontFace{Font: boldFont, Face: boldFace})
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
	err        error
}

// load reads master.mdb, the saved races and the distance type icons concurrently and
// aggregates them into the TT page
func load(assets fs.FS, fontData []byte) loadResult {
	var (
		result        loadResult
		resultSet     races.TeamTrialResultSet
		distanceIcons distanceTypeIcons
	)
	var g errgroup.Group
	g.Go(func() (err error) { result.shaper, err = newShaper(fontData, true); return })
	g.Go(func() (err error) { result.dataStore, err = data.Init(); return })
	g.Go(func() (err error) { resultSet, err = loadTeamTrialResults(); return })
	g.Go(func() (err error) { distanceIcons, err = loadDistanceTypeIcons(assets); return })
	if err := g.Wait(); err != nil {
		return loadResult{err: err}
	}
	result.teamTrials = newTeamTrialsPage(result.dataStore, resultSet, distanceIcons)
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
}

// run drives the window event loop until the window closes
func run(assets fs.FS, fontData []byte, shaper *text.Shaper) error {
	var window app.Window
	window.Option(app.Title("Umango"), app.Size(windowWidth, windowHeight))

	theme := material.NewTheme()
	theme.Shaper = shaper
	theme.TextSize = textSize
	theme.Palette = material.Palette{
		Bg:         colorBackground,
		Fg:         colorForeground,
		ContrastBg: colorPrimary,
		ContrastFg: colorOnPrimary,
	}
	s := screen{theme: theme, window: &window, explorer: explorer.NewExplorer(&window)}

	loads := make(chan loadResult, 1)
	go func() {
		loads <- load(assets, fontData)
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
			default:
			}
			gtx := app.NewContext(&ops, e)
			s.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

// applyLoad installs the background load results, or its error message on failure
func (s *screen) applyLoad(result loadResult) {
	if result.err != nil {
		log.Printf("load data: %v", result.err)
		s.loadError = loadingErrorPrefix + result.err.Error()
		return
	}
	s.theme.Shaper = result.shaper
	s.teamTrials = result.teamTrials
	s.veterans = newVeteranPage(result.dataStore)
}

func (s *screen) layout(gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, s.theme.Bg)
	switch {
	case s.loadError != "":
		return layoutCenteredText(gtx, s.theme, s.loadError)
	case s.teamTrials == nil:
		return layoutLoading(gtx, s.theme)
	}
	for i := range s.tabs {
		if s.tabs[i].Clicked(gtx) {
			s.activeTab = i
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(s.layoutTabBar),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if s.activeTab == tabVeteranList {
				return s.veterans.layout(gtx, s.theme, s.explorer, s.window.Invalidate)
			}
			return s.teamTrials.layout(gtx, s.theme)
		}),
	)
}

// layoutTabBar draws the tab titles left to right, underlining the active tab
func (s *screen) layoutTabBar(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	var size image.Point
	for i := range s.tabs {
		offset := op.Offset(image.Pt(size.X, 0)).Push(gtx.Ops)
		dims := material.Clickable(gtx, &s.tabs[i], func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(s.theme, tabTitles[i])
			if i == s.activeTab {
				label.Color = colorPrimary
			}
			dims := layout.UniformInset(tabPadding).Layout(gtx, label.Layout)
			if i == s.activeTab {
				indicator := image.Rectangle{
					Min: image.Pt(0, dims.Size.Y-gtx.Dp(tabIndicatorHeight)),
					Max: dims.Size,
				}
				paint.FillShape(gtx.Ops, colorPrimary, clip.Rect(indicator).Op())
			}
			return dims
		})
		offset.Pop()
		size.X += dims.Size.X
		size.Y = max(size.Y, dims.Size.Y)
	}
	return layout.Dimensions{Size: size}
}

func layoutLoading(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(material.Body1(theme, loadingText).Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				size := gtx.Dp(loadingIndicatorSize)
				gtx.Constraints = layout.Exact(image.Pt(size, size))
				return material.Loader(theme).Layout(gtx)
			}),
		)
	})
}

func layoutCenteredText(gtx layout.Context, theme *material.Theme, s string) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Body1(theme, s)
		label.Alignment = text.Middle
		return label.Layout(gtx)
	})
}
