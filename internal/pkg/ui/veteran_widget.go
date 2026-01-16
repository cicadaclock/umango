package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/veteran"
)

const (
	// Widget visuals
	factorCornerRadius           float32 = 32
	factorInternalPaddingDivisor float32 = 8
)

type VeteranWidget struct {
	widget.BaseWidget
	Veteran   veteran.Veteran
	dataStore *data.DataStore

	filterLegacy veteran.LegacyName
	filterFactor data.FactorType
}

func NewVeteranWidget(dataStore *data.DataStore, veteran veteran.Veteran) *VeteranWidget {
	veteranWidget := VeteranWidget{
		Veteran:   veteran,
		dataStore: dataStore,
	}
	veteranWidget.ExtendBaseWidget(&veteranWidget)
	return &veteranWidget
}

func (v *VeteranWidget) CreateRenderer() fyne.WidgetRenderer {
	view := container.NewVBox()
	veteranView, err := createVeteranView(v)
	if err != nil {
		errMsg := fmt.Sprintf("could not render veteran id=%d: %v", v.Veteran.LocalVeteranId, err)
		view.Add(canvas.NewText(errMsg, color.Black))
	} else {
		view.Add(veteranView)
	}
	return widget.NewSimpleRenderer(view)
}

// Creates view of all factors
func createVeteranView(v *VeteranWidget) (*fyne.Container, error) {
	mainFactors := v.Veteran.FactorIdArray
	p1Factors := v.Veteran.LegacyFactor(veteran.LegacyNameParentL)
	p2Factors := v.Veteran.LegacyFactor(veteran.LegacyNameParentR)
	if len(mainFactors) == 0 {
		return nil, fmt.Errorf("veteran has no factors")
	}
	if len(p1Factors) == 0 {
		return nil, fmt.Errorf("parent 1 has no factors")
	}
	if len(p2Factors) == 0 {
		return nil, fmt.Errorf("parent 2 has no factors")
	}

	factorWidgets := VeteranFactorWidgets{}
	factorWidgets.CreateFactorWidgets(mainFactors, v.dataStore, veteran.LegacyNameMain)
	factorWidgets.CreateFactorWidgets(p1Factors, v.dataStore, veteran.LegacyNameParentL)
	factorWidgets.CreateFactorWidgets(p2Factors, v.dataStore, veteran.LegacyNameParentR)

	rSparkView := container.NewHBox()
	gSparkView := container.NewHBox()
	bSparkView := container.NewHBox()
	whiteSparkView := container.New(layout.NewRowWrapLayout())
	addWidgetsToContainer(rSparkView, factorWidgets.RedFactorWidgets)
	addWidgetsToContainer(gSparkView, factorWidgets.GreenFactorWidgets)
	addWidgetsToContainer(bSparkView, factorWidgets.BlueFactorWidgets)
	addWidgetsToContainer(whiteSparkView, factorWidgets.RaceFactorWidgets)
	addWidgetsToContainer(whiteSparkView, factorWidgets.WhiteFactorWidgets)

	rbSparkView := container.NewHBox(bSparkView, rSparkView)
	veteranView := container.NewVBox(rbSparkView, gSparkView, whiteSparkView)
	return veteranView, nil
}

type VeteranFactorWidgets struct {
	RedFactorWidgets   []*FactorWidget
	GreenFactorWidgets []*FactorWidget
	BlueFactorWidgets  []*FactorWidget
	RaceFactorWidgets  []*FactorWidget
	WhiteFactorWidgets []*FactorWidget
}

func (v *VeteranFactorWidgets) CreateFactorWidgets(factorIds []int, dataStore *data.DataStore, legacyName veteran.LegacyName) {
	for _, factorId := range factorIds {
		factorName := dataStore.FactorNames[factorId]
		factorType := data.FactorType(dataStore.FactorType[factorId])
		factorWidget := NewFactorWidget(
			factorName,
			factorId%100,
			factorType,
			factorCornerRadius,
			theme.Size(FontSizeVeteranWidget)/factorInternalPaddingDivisor,
			legacyName,
		)
		switch factorType {
		case data.FactorTypeRed:
			v.RedFactorWidgets = append(v.RedFactorWidgets, factorWidget)
		case data.FactorTypeGreen:
			v.GreenFactorWidgets = append(v.GreenFactorWidgets, factorWidget)
		case data.FactorTypeBlue:
			v.BlueFactorWidgets = append(v.BlueFactorWidgets, factorWidget)
		case data.FactorTypeRace:
			v.RaceFactorWidgets = append(v.RaceFactorWidgets, factorWidget)
		case data.FactorTypeWhite:
			v.WhiteFactorWidgets = append(v.WhiteFactorWidgets, factorWidget)
		}
	}
}

func addWidgetsToContainer(c *fyne.Container, fw []*FactorWidget) {
	for _, w := range fw {
		c.Add(w)
	}
}
