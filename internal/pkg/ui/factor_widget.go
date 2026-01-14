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
)

type FactorWidget struct {
	widget.BaseWidget
	Factor       string
	Level        int
	FactorType   data.FactorType
	CornerRadius float32
}

func NewFactorWidget(factor string, level int, factorType data.FactorType, cornerRadius float32) *FactorWidget {
	sparkWidget := FactorWidget{
		Factor:       factor,
		Level:        level,
		FactorType:   factorType,
		CornerRadius: cornerRadius,
	}
	sparkWidget.ExtendBaseWidget(&sparkWidget)
	return &sparkWidget
}

func (fw *FactorWidget) CreateRenderer() fyne.WidgetRenderer {
	richText := createFactorRichText(fw.Factor, fw.Level, fw.FactorType)
	rect := canvas.NewRectangle(getColor(fw.FactorType))
	rect.CornerRadius = fw.CornerRadius
	rect.FillColor = theme.Color(ColorNameFactorBackground)
	rect.StrokeColor = getColor(fw.FactorType)
	rect.StrokeWidth = 2.0

	richTextContainer := container.New(layout.NewCustomPaddedLayout(theme.Padding(), theme.Padding(), theme.Padding()*2, theme.Padding()*2), container.NewCenter(richText))

	factorView := container.NewHBox(container.NewStack(rect, richTextContainer))
	return widget.NewSimpleRenderer(factorView)
}

// Return color of the factor's type
func getColor(factorType data.FactorType) color.Color {
	var colorName fyne.ThemeColorName
	switch factorType {
	case data.FactorTypeBlue:
		colorName = ColorNameFactorBlue
	case data.FactorTypeRed:
		colorName = ColorNameFactorRed
	case data.FactorTypeGreen:
		colorName = ColorNameFactorGreen
	case data.FactorTypeWhite:
		colorName = ColorNameFactorWhite
	case data.FactorTypeRace:
		colorName = ColorNameFactorWhite
	}
	return theme.Color(colorName)
}

// Convert factor to richtext representation for color/font
func createFactorRichText(factor string, level int, factorType data.FactorType) *widget.RichText {
	var colorName fyne.ThemeColorName
	switch factorType {
	case data.FactorTypeBlue:
		colorName = ColorNameFactorBlue
	case data.FactorTypeRed:
		colorName = ColorNameFactorRed
	case data.FactorTypeGreen:
		colorName = ColorNameFactorGreen
	case data.FactorTypeWhite:
		colorName = ColorNameFactorWhite
	case data.FactorTypeRace:
		colorName = ColorNameFactorWhite
	}

	sparkTextSegment := widget.TextSegment{
		Text: factor + buildStars(" ", level),
		Style: widget.RichTextStyle{
			Inline:    true,
			ColorName: colorName,
			SizeName:  fyne.ThemeSizeName(FontSizeVeteranWidget),
		},
	}
	return widget.NewRichText(&sparkTextSegment)
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
