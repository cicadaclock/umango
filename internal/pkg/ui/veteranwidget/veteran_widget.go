package veteranwidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/ui/factorwidget"
	"github.com/cicadaclock/umango/internal/pkg/veteran"
)

const (
	// Widget visuals
	factorCornerRadius           float32 = 32
	factorInternalPaddingDivisor float32 = 8
)

type VeteranWidget struct {
	widget.BaseWidget
	VeteranSlice *veteran.VeteranSlice
	dataStore    *data.DataStore
	widgets      []*factorwidget.FactorWidget
	list         *widget.List
}

func NewVeteranWidget(
	veteranSlice *veteran.VeteranSlice,
	dataStore *data.DataStore,
) *VeteranWidget {
	v := &VeteranWidget{
		VeteranSlice: veteranSlice,
		dataStore:    dataStore,
	}

	v.addFactorWidgets()

	v.list = widget.NewList(
		func() int {
			return veteranSlice.Len()
		},
		func() fyne.CanvasObject {
			// widget := factor_widget.NewFactorWidget(dataStore, 0, []int{}, []int{}, []int{})
			// return container.NewVBox(widget.Content)
			return container.NewVBox()
		},
		func(i widget.ListItemID, co fyne.CanvasObject) {
			c := co.(*fyne.Container)
			c.RemoveAll()
			c.Add(v.widgets[i].Content)
			c.Resize(v.widgets[i].MinSize())
		},
	)
	v.ExtendBaseWidget(v)
	return v
}

func (v *VeteranWidget) Resize(size fyne.Size) {
	for i, widget := range v.widgets {
		widget.Resize(size)
		v.list.SetItemHeight(i, widget.MinSize().Height)
	}
	v.BaseWidget.Resize(size)
}

func (v *VeteranWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(v.list)
}

func (v *VeteranWidget) addFactorWidgets() {
	// Create FactorWidget list
	v.widgets = make([]*factorwidget.FactorWidget, 0, v.VeteranSlice.Len())
	for i := range v.VeteranSlice.Len() {
		var factors, factorsP1, factorsP2 []int
		factors = v.VeteranSlice.FactorIdArray[i]
		for _, successionChara := range v.VeteranSlice.SuccessionCharaArray[i] {
			if successionChara.PositionId == factorwidget.ParentALegacy.Int() {
				factorsP1 = successionChara.FactorIdArray
			} else if successionChara.PositionId == factorwidget.ParentBLegacy.Int() {
				factorsP2 = successionChara.FactorIdArray
			}
		}

		// Set widget visuals and data
		widget := factorwidget.NewFactorWidget(v.dataStore, v.VeteranSlice.LocalVeteranId[i], factors, factorsP1, factorsP2)
		widget.SetPadding(-3, -3, 5, 5)
		widget.CornerRadius = factorCornerRadius
		widget.SetFactors(factors, factorsP1, factorsP2)
		v.widgets = append(v.widgets, widget)
	}
}
