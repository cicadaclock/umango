package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/veteran"
)

type VeteranWidget struct {
	widget.BaseWidget
	Veteran   veteran.Veteran
	dataStore *data.DataStore
}

const (
	factorCornerRadius float32 = 32
)

func NewVeteranWidget(dataStore *data.DataStore, veteran veteran.Veteran) *VeteranWidget {
	veteranWidget := VeteranWidget{
		Veteran:   veteran,
		dataStore: dataStore,
	}
	veteranWidget.ExtendBaseWidget(&veteranWidget)
	return &veteranWidget
}

func (item *VeteranWidget) CreateRenderer() fyne.WidgetRenderer {
	temp := createSingleVeteranView(item.dataStore, item.Veteran.FactorIdArray)
	// temp2 := container.NewHScroll(temp)
	return widget.NewSimpleRenderer(temp)
}

// Creates view of all factors
func createSingleVeteranView(dataStore *data.DataStore, factors []int) *fyne.Container {
	var blueFactors, redFactors, greenFactors, raceFactors, whiteFactors []*FactorWidget
	for _, factorId := range factors {
		factor := dataStore.FactorNames[factorId]
		factorType := data.FactorType(dataStore.FactorType[factorId])
		factorWidget := NewFactorWidget(factor, factorId%100, factorType, factorCornerRadius)
		switch factorType {
		case data.FactorTypeBlue:
			blueFactors = append(blueFactors, factorWidget)
		case data.FactorTypeRed:
			redFactors = append(redFactors, factorWidget)
		case data.FactorTypeGreen:
			greenFactors = append(greenFactors, factorWidget)
		case data.FactorTypeRace:
			raceFactors = append(raceFactors, factorWidget)
		case data.FactorTypeWhite:
			whiteFactors = append(whiteFactors, factorWidget)
		}
	}

	rgbSparkView := container.NewVBox()
	whiteSparkView := container.New(layout.NewRowWrapLayout())
	for _, factorWidget := range blueFactors {
		rgbSparkView.Add(factorWidget)
	}
	for _, factorWidget := range redFactors {
		rgbSparkView.Add(factorWidget)
	}
	for _, factorWidget := range greenFactors {
		rgbSparkView.Add(factorWidget)
	}
	for _, factorWidget := range raceFactors {
		whiteSparkView.Add(factorWidget)
	}
	for _, factorWidget := range whiteFactors {
		whiteSparkView.Add(factorWidget)
	}

	veteranView := container.NewBorder(nil, nil, rgbSparkView, nil, whiteSparkView)
	return veteranView
}
