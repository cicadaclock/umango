package factorwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/cicadaclock/umango/internal/ui/apptheme"
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
		color = apptheme.ColorFactorMain
	case ParentALegacy:
		color = apptheme.ColorFactorWhite
	case ParentBLegacy:
		color = apptheme.ColorFactorWhite
	default:
		color = apptheme.ColorFactorWhite
	}
	return color
}

func (l LegacyType) Color() color.Color {
	return theme.Color(l.ColorName())
}
