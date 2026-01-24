package app_theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	// Font sizes
	FontSizeVeteranWidget fyne.ThemeSizeName = "veteranWidget"
	// Colors
	ColorFactorMain       fyne.ThemeColorName = "factorMain"
	ColorFactorBlue       fyne.ThemeColorName = "factorBlue"
	ColorFactorRed        fyne.ThemeColorName = "factorRed"
	ColorFactorGreen      fyne.ThemeColorName = "factorGreen"
	ColorFactorWhite      fyne.ThemeColorName = "factorWhite"
	ColorFactorRace       fyne.ThemeColorName = "factorRace"
	ColorFactorBackground fyne.ThemeColorName = "factorBackground"
)

var (
	colorFactorMain       = color.NRGBA{R: 0xff, G: 0xcf, B: 0x33, A: 0xff}
	colorFactorBlue       = color.NRGBA{R: 0x37, G: 0xb7, B: 0xf4, A: 0xff}
	colorFactorRed        = color.NRGBA{R: 0xff, G: 0x76, B: 0xb2, A: 0xff}
	colorFactorGreen      = color.NRGBA{R: 0xad, G: 0xe2, B: 0x60, A: 0xff}
	colorFactorWhite      = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	colorFactorRace       = color.NRGBA{R: 0xc8, G: 0xa2, B: 0xc8, A: 0xff}
	colorFactorBackground = color.NRGBA{R: 0x1f, G: 0x1f, B: 0x1f, A: 0xff}
	colorBackground       = color.NRGBA{R: 0xdd, G: 0xdd, B: 0xdd, A: 0xff}
)

type AppTheme struct{}

var _ fyne.Theme = (*AppTheme)(nil)

func (m AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case ColorFactorMain:
		return colorFactorMain
	case ColorFactorBlue:
		return colorFactorBlue
	case ColorFactorRed:
		return colorFactorRed
	case ColorFactorGreen:
		return colorFactorGreen
	case ColorFactorWhite:
		return colorFactorWhite
	case ColorFactorRace:
		return colorFactorRace
	case ColorFactorBackground:
		return colorFactorBackground
	case theme.ColorNameBackground:
		return colorBackground
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	font, _ := fyne.LoadResourceFromPath("internal/font/Inter-VariableFont_opsz,wght.ttf")
	return font
}

func (m AppTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case FontSizeVeteranWidget:
		return 14.0
	}
	return theme.DefaultTheme().Size(name)
}
