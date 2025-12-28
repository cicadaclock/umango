package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	ColorNameFactorBlue  fyne.ThemeColorName = "factorBlue"
	ColorNameFactorPink  fyne.ThemeColorName = "factorPink"
	ColorNameFactorGreen fyne.ThemeColorName = "factorGreen"
	ColorNameFactorWhite fyne.ThemeColorName = "factorWhite"
)

var (
	colorFactorBlue = color.NRGBA{R: 0x00, G: 0x6c, B: 0xff, A: 0xff}
	colorFactorPink = color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff}
)

type myTheme struct{}

var _ fyne.Theme = (*myTheme)(nil)

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == ColorNameFactorBlue {
		return colorFactorBlue
	} else if name == ColorNameFactorPink {
		return colorFactorPink
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	// if name == theme.IconNameHome {
	// 	return fyne.NewStaticResource("myHome", homeBytes)
	// }

	return theme.DefaultTheme().Icon(name)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
