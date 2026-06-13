package centersteppedlayout

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// Resizes the object in steps
type HSteppedLayout struct {
	minPct float32
	maxPct float32
}

func NewHStepped(minPct, maxPct float32, objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(NewHSteppedLayout(minPct, maxPct), objects...)
}

func NewHSteppedLayout(minPct, maxPct float32) fyne.Layout {
	return HSteppedLayout{minPct, maxPct}
}

func (l HSteppedLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		minSize = minSize.Max(child.MinSize())
	}
	return minSize
}

func (l HSteppedLayout) Size(objects []fyne.CanvasObject) fyne.Size {
	size := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		size = size.Max(child.Size())
	}
	return size
}

func (l HSteppedLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	minWidth := size.Width * l.minPct
	maxWidth := size.Width * l.maxPct
	width := l.Size(objects).Width
	pct := width / size.Width

	// Conditionally resize width
	if pct < l.minPct { // Increasing in size, snap to max
		topLeft := abs(size.Width-maxWidth) / 2
		newSize := fyne.NewSize(maxWidth, size.Height)
		resizeAndMove(objects, newSize, fyne.NewPos(topLeft, 0))
	} else if pct > l.maxPct { // Decreasing in size, snap to min
		topLeft := abs(size.Width-minWidth) / 2
		newSize := fyne.NewSize(minWidth, size.Height)
		resizeAndMove(objects, newSize, fyne.NewPos(topLeft, 0))
	} else { // Always resize height if no conditions met
		topLeft := abs(width-size.Width) / 2
		newSize := fyne.NewSize(width, size.Height)
		resizeAndMove(objects, newSize, fyne.NewPos(topLeft, 0))
	}
}

func abs(x float32) float32 {
	return math.Float32frombits(math.Float32bits(x) &^ (1 << 31))
}

func resizeAndMove(objects []fyne.CanvasObject, size fyne.Size, topLeft fyne.Position) {
	for _, child := range objects {
		child.Resize(size)
		child.Move(topLeft)
	}
}
