package factorwidget

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
)

type FactorWidget struct {
	widget.BaseWidget

	// Visuals

	VeteranId     int                // The ID of the veteran
	Alignment     fyne.TextAlign     // The alignment of the text
	Wrapping      fyne.TextWrap      // The wrapping of the text
	SizeName      fyne.ThemeSizeName // The theme size name for the text size of the factor widget
	StrokeWidth   float32            // The size of the rect outline
	CornerRadius  float32            // The radius of the rect container
	TopPadding    float32            // The top padding of the rect container
	BottomPadding float32            // The bottom padding of the rect container
	LeftPadding   float32            // The left padding of the rect container
	RightPadding  float32            // The right padding of the rect container
	SuffixType    SuffixType         // SuffixType informs how the factor widget should display factor suffixes

	// Data

	Factors   []int // Factors is the array of factor names
	FactorsP1 []int // FactorsP1 is the array of parent 1's factor names
	FactorsP2 []int // FactorsP1 is the array of parent 2's factor names

	// Internal data

	dataStore      *data.DataStore // dataStore has transformations for our data
	factorSuffixes []string        // factorSuffixes is the array of stars (or other suffixes) appended to each factor
	redFactors     *fyne.Container // redFactors is the view of all red factors
	greenFactors   *fyne.Container // greenFactors is the view of all green factors
	blueFactors    *fyne.Container // blueFactors is the view of all blue factors
	whiteFactors   *fyne.Container // whiteFactors is the view of all white factors
	raceFactors    *fyne.Container // raceFactors is the view of all race factors
	Content        *fyne.Container // content is the combined view of all the factor containers
}

func NewFactorWidget(
	dataStore *data.DataStore,
	veteranId int,
	factors, factorsP1, factorsP2 []int,
) *FactorWidget {
	return NewFactorWidgetWithStyle(
		dataStore,
		veteranId,
		factors, factorsP1, factorsP2,
		StarSuffix,
		fyne.TextAlignCenter,
		theme.Padding(), theme.Padding(), theme.Padding(), theme.Padding(),
	)
}

// NewFactorWidgetWithStyle creates a new label widget with the set text content
func NewFactorWidgetWithStyle(
	dataStore *data.DataStore,
	veteranId int,
	factors, factorsP1, factorsP2 []int,
	suffixType SuffixType,
	alignment fyne.TextAlign,
	topPadding, bottomPadding, leftPadding, rightPadding float32,
) *FactorWidget {
	f := &FactorWidget{
		dataStore:     dataStore,
		VeteranId:     veteranId,
		Factors:       factors,
		FactorsP1:     factorsP1,
		FactorsP2:     factorsP2,
		SuffixType:    suffixType,
		Alignment:     alignment,
		TopPadding:    topPadding,
		BottomPadding: bottomPadding,
		LeftPadding:   leftPadding,
		RightPadding:  rightPadding,
		redFactors:    container.NewHBox(),
		greenFactors:  container.NewHBox(),
		blueFactors:   container.NewHBox(),
		whiteFactors:  container.New(layout.NewRowWrapLayout()),
		raceFactors:   container.New(layout.NewRowWrapLayout()),
	}
	// rgbContainer := container.NewVBox(f.blueFactors, f.redFactors, f.greenFactors)
	// wContainer := container.NewVBox(f.raceFactors, f.whiteFactors)
	// f.Content = container.NewBorder(rgbContainer, nil, nil, nil, wContainer)
	f.Content = container.NewVBox(f.blueFactors, f.redFactors, f.greenFactors, f.raceFactors, f.whiteFactors)

	// For the empty case, we append a single empty rich text widget for all
	// factor containers so that MinSize() is accurate when this widget is empty
	if len(factors) == 0 {
		f.setEmptyFactors()
	} else {
		f.addFactors()
	}

	f.ExtendBaseWidget(f)
	return f
}

func (f *FactorWidget) MinSize() fyne.Size {
	return f.Content.MinSize()
}

func (f *FactorWidget) Resize(size fyne.Size) {
	f.Content.Resize(size)
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (f *FactorWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(f.Content)
}

func (f *FactorWidget) SetPadding(topPadding, bottomPadding, leftPadding, rightPadding float32) {
	f.TopPadding = topPadding
	f.BottomPadding = bottomPadding
	f.LeftPadding = leftPadding
	f.RightPadding = rightPadding
}

func (f *FactorWidget) SetFactors(factors, factorsP1, factorsP2 []int) {
	f.clearFactorContainers()
	f.Factors = factors
	f.FactorsP1 = factorsP1
	f.FactorsP2 = factorsP2
	f.addFactors()
}

func (f *FactorWidget) clearFactorContainers() {
	f.redFactors.RemoveAll()
	f.greenFactors.RemoveAll()
	f.blueFactors.RemoveAll()
	f.whiteFactors.RemoveAll()
	f.raceFactors.RemoveAll()
}

func (f *FactorWidget) addFactors() {
	f.addFactorsFromLegacyType(MainLegacy)
	f.addFactorsFromLegacyType(ParentALegacy)
	f.addFactorsFromLegacyType(ParentBLegacy)
}

func (f *FactorWidget) addFactorsFromLegacyType(legacyType LegacyType) {
	var factors []int
	switch legacyType {
	case MainLegacy:
		factors = f.Factors
	case ParentALegacy:
		factors = f.FactorsP1
	case ParentBLegacy:
		factors = f.FactorsP2
	}

	if len(factors) == 0 {
		return
	}

	factorNames := f.dataStore.MapFactorNames(factors)
	factorTypes := f.dataStore.MapFactorTypes(factors)
	factorLevels := f.dataStore.FactorLevels(factors)
	factorSuffixes := newSuffixes(factorLevels, f.SuffixType)
	factorTexts := newRichTexts(factorNames, factorSuffixes, factorTypes, legacyType, f.SizeName)
	factorRects := newRects(factorTypes, f.StrokeWidth, f.CornerRadius)
	factorStacks := newStacks(factorTexts, factorRects, f.TopPadding, f.BottomPadding, f.LeftPadding, f.RightPadding)

	for i, stack := range factorStacks {
		switch FactorType(factorTypes[i]) {
		case RedFactor:
			f.redFactors.Add(stack)
		case GreenFactor:
			f.greenFactors.Add(stack)
		case BlueFactor:
			f.blueFactors.Add(stack)
		case WhiteFactor:
			f.whiteFactors.Add(stack)
		case RaceFactor:
			f.raceFactors.Add(stack)
		}
	}
}

// Adds "empty" factor containers for MinSize calculations
func (f *FactorWidget) setEmptyFactors() {
	text := widget.NewRichText(
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				Inline:   true,
				SizeName: fyne.ThemeSizeName(f.SizeName),
			},
			Text: "",
		},
	)
	f.redFactors.Add(text)
	f.greenFactors.Add(text)
	f.blueFactors.Add(text)
	f.whiteFactors.Add(text)
	f.raceFactors.Add(text)
}

// Adds all factor containers to content
func (f *FactorWidget) setContent() {
	f.Content.Add(f.blueFactors)
	f.Content.Add(f.redFactors)
	f.Content.Add(f.greenFactors)
	f.Content.Add(f.raceFactors)
	f.Content.Add(f.whiteFactors)
}

func newSuffixes(factorLevels []int, suffixType SuffixType) []string {
	result := make([]string, 0, len(factorLevels))
	if suffixType == StarSuffix {
		for _, factorLevel := range factorLevels {
			result = append(result, strings.Repeat("★", factorLevel))
		}
	}
	return result
}

// Returns rich texts with color according to type and legacy
func newRichTexts(factorNames, factorSuffixes []string, factorTypes []int, legacyType LegacyType, sizeName fyne.ThemeSizeName) []*widget.RichText {
	r := make([]*widget.RichText, 0, len(factorNames))

	suffixBuffer := " "
	var colorName, starsColorName fyne.ThemeColorName
	for i := range len(factorNames) {
		colorName = FactorType(factorTypes[i]).ColorName()
		starsColorName = legacyType.ColorName()
		r = append(r,
			widget.NewRichText(
				&widget.TextSegment{
					Style: widget.RichTextStyle{
						Inline:    true,
						ColorName: colorName,
						SizeName:  fyne.ThemeSizeName(sizeName),
					},
					Text: factorNames[i],
				},
				&widget.TextSegment{
					Style: widget.RichTextStyle{
						Inline:    true,
						ColorName: starsColorName,
						SizeName:  fyne.ThemeSizeName(sizeName),
					},
					Text: suffixBuffer + factorSuffixes[i],
				},
			),
		)
	}
	return r
}

// Returns rectangles with colors according to type
func newRects(types []int, strokeWidth, cornerRadius float32) []*canvas.Rectangle {
	rects := make([]*canvas.Rectangle, 0, len(types))
	for i := range len(types) {
		rect := canvas.NewRectangle(NoneFactor.Color())
		rect.CornerRadius = cornerRadius
		rect.StrokeColor = FactorType(types[i]).Color()
		rect.StrokeWidth = strokeWidth
		rects = append(rects, rect)
	}
	return rects
}

// Returns containers with texts stacked over rects
func newStacks(texts []*widget.RichText, rects []*canvas.Rectangle, topPadding, bottomPadding, leftPadding, rightPadding float32) []*fyne.Container {
	sparks := make([]*fyne.Container, 0, len(texts))
	for i := range len(texts) {
		sparks = append(sparks,
			container.NewStack(
				rects[i],
				container.NewCenter(container.New(layout.NewCustomPaddedLayout(topPadding, bottomPadding, leftPadding, rightPadding), texts[i])),
			),
		)
	}
	return sparks
}
