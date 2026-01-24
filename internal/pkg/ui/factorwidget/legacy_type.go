package factorwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/cicadaclock/umango/internal/pkg/ui/app_theme"
)

// LegacyType represents the parent that each factor in factor_widget belongs to
type LegacyType int

const (
	MainLegacy    LegacyType = 0
	ParentALegacy LegacyType = 10
	ParentBLegacy LegacyType = 20
)

// Returns the integer representation of the legacy name
func (l LegacyType) Int() int {
	return int(l)
}

func (l LegacyType) ColorName() fyne.ThemeColorName {
	var color fyne.ThemeColorName
	switch l {
	case MainLegacy:
		color = app_theme.ColorFactorMain
	case ParentALegacy:
		color = app_theme.ColorFactorWhite
	case ParentBLegacy:
		color = app_theme.ColorFactorWhite
	default:
		color = app_theme.ColorFactorWhite
	}
	return color
}

func (l LegacyType) Color() color.Color {
	return theme.Color(l.ColorName())
}
