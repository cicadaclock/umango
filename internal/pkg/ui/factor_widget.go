package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/veteran"
)

type FactorWidget struct {
	widget.BaseWidget
	Factor          string
	Level           int
	FactorType      data.FactorType
	CornerRadius    float32
	InternalPadding float32
	LegacyName      veteran.LegacyName
}

func NewFactorWidget(factorName string, level int, factorType data.FactorType, cornerRadius float32, internalPadding float32, legacyName veteran.LegacyName) *FactorWidget {
	sparkWidget := FactorWidget{
		Factor:          factorName,
		Level:           level,
		FactorType:      factorType,
		CornerRadius:    cornerRadius,
		InternalPadding: internalPadding,
		LegacyName:      legacyName,
	}
	sparkWidget.ExtendBaseWidget(&sparkWidget)
	return &sparkWidget
}

func (fw *FactorWidget) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(getColor(fw.FactorType))
	rect.CornerRadius = fw.CornerRadius
	rect.FillColor = theme.Color(ColorFactorBackground)
	rect.StrokeColor = getColor(fw.FactorType)
	rect.StrokeWidth = 2.0

	richText := createFactorRichText(fw.Factor, fw.Level, fw.FactorType, fw.LegacyName)
	richTextContainer := container.New(layout.NewCustomPaddedLayout(0, 0, fw.InternalPadding*2, fw.InternalPadding*2), container.NewCenter(richText))

	factorView := container.NewHBox(container.NewStack(rect, richTextContainer))
	return widget.NewSimpleRenderer(factorView)
}

// Return color name of the factor's type
func getColorName(factorType data.FactorType) fyne.ThemeColorName {
	var colorName fyne.ThemeColorName
	switch factorType {
	case data.FactorTypeRed:
		colorName = ColorFactorRed
	case data.FactorTypeGreen:
		colorName = ColorFactorGreen
	case data.FactorTypeBlue:
		colorName = ColorFactorBlue
	case data.FactorTypeRace:
		colorName = ColorFactorWhite
	case data.FactorTypeWhite:
		colorName = ColorFactorWhite
	}
	return colorName
}

// Return color of the factor's type
func getColor(factorType data.FactorType) color.Color {
	return theme.Color(getColorName(factorType))
}

// Convert factor to richtext representation for color/font
func createFactorRichText(factorName string, level int, factorType data.FactorType, legacyName veteran.LegacyName) *widget.RichText {
	var colorName, starsColorName fyne.ThemeColorName
	switch factorType {
	case data.FactorTypeBlue:
		colorName = ColorFactorBlue
	case data.FactorTypeRed:
		colorName = ColorFactorRed
	case data.FactorTypeGreen:
		colorName = ColorFactorGreen
	case data.FactorTypeWhite:
		colorName = ColorFactorWhite
	case data.FactorTypeRace:
		colorName = ColorFactorWhite
	}
	if legacyName == veteran.LegacyNameMain {
		starsColorName = ColorFactorMain
	} else {
		starsColorName = colorName
	}

	sparkTextSegment := widget.TextSegment{
		Text: factorName,
		Style: widget.RichTextStyle{
			Inline:    true,
			ColorName: colorName,
			SizeName:  fyne.ThemeSizeName(FontSizeVeteranWidget),
		},
	}

	starsTextSegment := widget.TextSegment{
		Text: buildStars(" ", level),
		Style: widget.RichTextStyle{
			Inline:    true,
			ColorName: starsColorName,
			SizeName:  fyne.ThemeSizeName(FontSizeVeteranWidget),
		},
	}
	return widget.NewRichText(&sparkTextSegment, &starsTextSegment)
}

// Append stars based on factor level
func buildStars(separator string, level int) string {
	var stars strings.Builder
	stars.WriteString(separator)
	for range level {
		_, _ = stars.WriteString("★")
	}
	return stars.String()
}

type FactorWidgets struct {
	widget.BaseWidget

	dataStore       *data.DataStore
	CornerRadius    float32
	InternalPadding float32

	LegacyNames []veteran.LegacyName
	Factors     []int
}

func NewFactorWidgets(dataStore *data.DataStore, cornerRadius, internalPadding float32, legacyNames []veteran.LegacyName, factors []int) *FactorWidgets {
	factorWidgets := FactorWidgets{
		dataStore:       dataStore,
		CornerRadius:    cornerRadius,
		InternalPadding: internalPadding,
		LegacyNames:     legacyNames,
		Factors:         factors,
	}
	factorWidgets.ExtendBaseWidget(&factorWidgets)
	return &factorWidgets
}

func (fw *FactorWidgets) CreateRenderer() fyne.WidgetRenderer {
	rects := createRects(fw.Factors, fw.CornerRadius, fw.dataStore)
	richTexts := createFactorRichTexts(fw.Factors, fw.LegacyNames, fw.dataStore)
	richTextContainers := createRichTextContainers(richTexts, fw.InternalPadding)
	factors := createFactorContainers(rects, richTextContainers)
	// TODO: Render factors together like in veteran_widget.go
	return widget.NewSimpleRenderer(factors)
}

// Convert factor to richtext representation for color/font
func createFactorRichTexts(factors []int, legacyNames []veteran.LegacyName, dataStore *data.DataStore) []*widget.RichText {
	richTexts := make([]*widget.RichText, 0, len(factors))

	var colorName, starsColorName fyne.ThemeColorName

	for i, factorId := range factors {
		factorType := data.FactorType(dataStore.FactorType[factorId])
		colorName = getColorName(factorType)
		if legacyNames[i] == veteran.LegacyNameMain {
			starsColorName = ColorFactorMain
		} else {
			starsColorName = colorName
		}
		sparkTextSegment := widget.TextSegment{
			Text: dataStore.FactorNames[factorId],
			Style: widget.RichTextStyle{
				Inline:    true,
				ColorName: colorName,
				SizeName:  fyne.ThemeSizeName(FontSizeVeteranWidget),
			},
		}
		starsTextSegment := widget.TextSegment{
			Text: buildStars(" ", factorId%100),
			Style: widget.RichTextStyle{
				Inline:    true,
				ColorName: starsColorName,
				SizeName:  fyne.ThemeSizeName(FontSizeVeteranWidget),
			},
		}
		richTexts = append(richTexts, widget.NewRichText(&sparkTextSegment, &starsTextSegment))
	}

	return richTexts
}

func createRects(factors []int, cornerRadius float32, dataStore *data.DataStore) []*canvas.Rectangle {
	rects := make([]*canvas.Rectangle, 0, len(factors))
	for _, factorId := range factors {
		factorType := data.FactorType(dataStore.FactorType[factorId])
		rect := canvas.NewRectangle(getColor(factorType))
		rect.CornerRadius = cornerRadius
		rect.FillColor = theme.Color(ColorFactorBackground)
		rect.StrokeColor = getColor(factorType)
		rect.StrokeWidth = 2.0
		rects = append(rects, rect)
	}
	return rects
}

func createRichTextContainers(richTexts []*widget.RichText, internalPadding float32) []*fyne.Container {
	c := make([]*fyne.Container, 0, len(richTexts))
	for _, text := range richTexts {
		c = append(c, container.New(layout.NewCustomPaddedLayout(0, 0, internalPadding*2, internalPadding*2), container.NewCenter(text)))
	}
	return c
}

func createFactorContainers(rects []*canvas.Rectangle, richTextContainers []*fyne.Container) []*fyne.Container {
	c := make([]*fyne.Container, 0, len(rects))
	for i := range len(rects) {
		c = append(c, container.NewHBox(container.NewStack(rects[i], richTextContainers[i])))
	}
	return c
}
