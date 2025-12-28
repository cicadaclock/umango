package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/veteran"
)

type VeteranWidget struct {
	widget.BaseWidget
	Veteran   veteran.Veteran
	dataStore *data.DataStore
}

func NewVeteranView(dataStore *data.DataStore, veteran veteran.Veteran) *VeteranWidget {
	veteranView := VeteranWidget{
		Veteran:   veteran,
		dataStore: dataStore,
	}

	return &veteranView
}

func (item *VeteranWidget) CreateRenderer() fyne.WidgetRenderer {
	temp := createSingleVeteranView(item.dataStore, item.Veteran.FactorIdArray)
	temp2 := container.NewHScroll(temp)
	return widget.NewSimpleRenderer(temp2)
}

func createSingleVeteranView(dataStore *data.DataStore, factors []int) *fyne.Container {
	sparkView := container.NewHBox()
	for _, factorId := range factors {
		spark := dataStore.FactorNames[factorId]
		var level strings.Builder
		for range factorId % 100 {
			_, _ = level.WriteString("★")
		}
		text := createFactorRichText(spark, level.String())
		sparkView.Add(text)
	}
	return sparkView
}

func createFactorRichText(spark, level string) *widget.RichText {
	sparkTextSegment := widget.TextSegment{
		Text: spark,
		Style: widget.RichTextStyle{
			Inline:    true,
			ColorName: ColorNameFactorBlue,
		},
	}
	levelTextSegment := widget.TextSegment{
		Text: level,
		Style: widget.RichTextStyle{
			Inline:    true,
			ColorName: ColorNameFactorPink,
		},
	}
	return widget.NewRichText(&sparkTextSegment, &levelTextSegment)
}
