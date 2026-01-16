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

// Return color of the factor's type
func getColor(factorType data.FactorType) color.Color {
	var colorName fyne.ThemeColorName
	switch factorType {
	case data.FactorTypeRed:
		colorName = ColorFactorRed
	case data.FactorTypeGreen:
		colorName = ColorFactorGreen
	case data.FactorTypeBlue:
		colorName = ColorFactorBlue
	case data.FactorTypeWhite:
		colorName = ColorFactorWhite
	case data.FactorTypeRace:
		colorName = ColorFactorWhite
	}
	return theme.Color(colorName)
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
