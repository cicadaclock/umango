package factorwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/cicadaclock/umango/internal/pkg/ui/app_theme"
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
		color = app_theme.ColorFactorBackground
	case RedFactor:
		color = app_theme.ColorFactorRed
	case GreenFactor:
		color = app_theme.ColorFactorGreen
	case BlueFactor:
		color = app_theme.ColorFactorBlue
	case WhiteFactor:
		color = app_theme.ColorFactorWhite
	case RaceFactor:
		color = app_theme.ColorFactorRace
	default:
		color = app_theme.ColorFactorWhite
	}
	return color
}

func (t FactorType) Color() color.Color {
	return theme.Color(t.ColorName())
}
