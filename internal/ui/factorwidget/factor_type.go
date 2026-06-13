package factorwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/cicadaclock/umango/internal/ui/apptheme"
)

// FactorType represents how a factor_widget should appear
type FactorType int

const (
	NoneFactor FactorType = iota
	BlueFactor
	RedFactor
	GreenFactor
	WhiteFactor
	RaceFactor
	SizeFactor
)

func (ft FactorType) Int() int {
	return int(ft)
}

func (t FactorType) ColorName() fyne.ThemeColorName {
	var color fyne.ThemeColorName
	switch t {
	case NoneFactor:
		color = apptheme.ColorFactorBackground
	case RedFactor:
		color = apptheme.ColorFactorRed
	case GreenFactor:
		color = apptheme.ColorFactorGreen
	case BlueFactor:
		color = apptheme.ColorFactorBlue
	case WhiteFactor:
		color = apptheme.ColorFactorWhite
	case RaceFactor:
		color = apptheme.ColorFactorRace
	default:
		color = apptheme.ColorFactorWhite
	}
	return color
}

func (t FactorType) Color() color.Color {
	return theme.Color(t.ColorName())
}
