package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	// Font sizes
	FontSizeVeteranWidget fyne.ThemeSizeName = "veteranWidget"
	// Colors
	ColorFactorBlue       fyne.ThemeColorName = "factorBlue"
	ColorFactorRed        fyne.ThemeColorName = "factorRed"
	ColorFactorGreen      fyne.ThemeColorName = "factorGreen"
	ColorFactorWhite      fyne.ThemeColorName = "factorWhite"
	ColorFactorBackground fyne.ThemeColorName = "factorBackground"
)

var (
	colorFactorBlue       = color.NRGBA{R: 0x37, G: 0xb7, B: 0xf4, A: 0xff}
	colorFactorRed        = color.NRGBA{R: 0xff, G: 0x76, B: 0xb2, A: 0xff}
	colorFactorGreen      = color.NRGBA{R: 0xad, G: 0xe2, B: 0x60, A: 0xff}
	colorFactorWhite      = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	colorFactorBackground = color.NRGBA{R: 0x1f, G: 0x1f, B: 0x1f, A: 0xff}
	colorBackground       = color.NRGBA{R: 0xdd, G: 0xdd, B: 0xdd, A: 0xff}
)

type myTheme struct{}

var _ fyne.Theme = (*myTheme)(nil)

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case ColorFactorBlue:
		return colorFactorBlue
	case ColorFactorRed:
		return colorFactorRed
	case ColorFactorGreen:
		return colorFactorGreen
	case ColorFactorWhite:
		return colorFactorWhite
	case ColorFactorBackground:
		return colorFactorBackground
	case theme.ColorNameBackground:
		return colorBackground
	}
	theme.DefaultTheme()

	return theme.DefaultTheme().Color(name, variant)
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	// if name == theme.IconNameHome {
	// 	return fyne.NewStaticResource("myHome", homeBytes)
	// }

	return theme.DefaultTheme().Icon(name)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	font, _ := fyne.LoadResourceFromPath("internal/font/Inter-VariableFont_opsz,wght.ttf")
	return font
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case FontSizeVeteranWidget:
		return 16.0
	}
	return theme.DefaultTheme().Size(name)
}
