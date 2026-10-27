package ui

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/veteran"
)

const (
	// Content width bounds as percentages of the page width
	contentMinPercent = 60
	contentMaxPercent = 80

	veteranPadding   = unit.Dp(10)
	chipGap          = unit.Dp(4)
	chipPaddingX     = unit.Dp(13)
	chipPaddingY     = unit.Dp(5)
	parentAPosition  = 10
	parentBPosition  = 20
	factorRowCount   = 5
	maxStarSuffixLen = 9
	// A star suffix is a prefix of this string, so suffixes never allocate
	starSuffixes = " ★★★★★★★★★"

	veteranPanelTitle  = "Veteran List"
	loadButtonText     = "Load veterans file…"
	emptyListText      = "Load a veterans file to list the factors of each veteran"
	loadButtonPaddingX = unit.Dp(12)
	loadButtonPaddingY = unit.Dp(6)
	loadErrorGap       = unit.Dp(10)
)

// factorType is a succession_factor factor_type, which selects a factor chip's color and row
type factorType uint8

const (
	noneFactor factorType = iota
	blueFactor
	redFactor
	greenFactor
	whiteFactor
	raceFactor
)

var (
	colorFactorBackground = rgb(0x232323)
	colorFactorOwnStars   = color.NRGBA{R: 0xff, G: 0xcf, B: 0x33, A: 0xff}
	colorFactorParentStar = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	factorColors = [...]color.NRGBA{
		blueFactor:  {R: 0x37, G: 0xb7, B: 0xf4, A: 0xff},
		redFactor:   {R: 0xff, G: 0x76, B: 0xb2, A: 0xff},
		greenFactor: {R: 0xad, G: 0xe2, B: 0x60, A: 0xff},
		whiteFactor: {R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		raceFactor:  {R: 0xc8, G: 0xa2, B: 0xc8, A: 0xff},
	}
	// Display row of each factor type; -1 hides the factor
	factorRows = [...]int8{
		noneFactor:  -1,
		blueFactor:  0,
		redFactor:   1,
		greenFactor: 2,
		raceFactor:  3,
		whiteFactor: 4,
	}
)

func factorRow(kind int) int {
	if kind < 0 || kind >= len(factorRows) {
		return -1
	}
	return int(factorRows[kind])
}

func starSuffix(level int) string {
	return starSuffixes[:len(" ")+len("★")*min(max(level, 0), maxStarSuffixLen)]
}

// factorChips holds the factor chips of every loaded veteran in display order
//
// The chips of veteran v in display row r end at rowEnds[v*factorRowCount+r] and start where
// the previous row ends.
type factorChips struct {
	names     []string
	stars     []string
	kinds     []factorType
	inherited []bool
	rowEnds   []int
}

// newFactorChips maps the factors of every veteran and its two parents in one batch, then
// groups each veteran's chips by display row, keeping its own factors before its parents'
func newFactorChips(dataStore *data.DataStore, vets *veteran.VeteranSlice) factorChips {
	n := vets.Len()
	var ids []int
	ownEnds := make([]int, n)
	veteranEnds := make([]int, n)
	for v := range n {
		ids = append(ids, vets.FactorIdArray[v]...)
		ownEnds[v] = len(ids)
		for _, position := range [...]int{parentAPosition, parentBPosition} {
			for _, parent := range vets.SuccessionCharaArray[v] {
				if parent.PositionId == position {
					ids = append(ids, parent.FactorIdArray...)
				}
			}
		}
		veteranEnds[v] = len(ids)
	}
	names := dataStore.FactorNames(ids)
	kinds := dataStore.FactorClass(ids)
	levels := dataStore.FactorLevels(ids)

	chips := factorChips{
		names:     make([]string, 0, len(ids)),
		stars:     make([]string, 0, len(ids)),
		kinds:     make([]factorType, 0, len(ids)),
		inherited: make([]bool, 0, len(ids)),
		rowEnds:   make([]int, 0, n*factorRowCount),
	}
	start := 0
	for v := range n {
		for row := range factorRowCount {
			for i := start; i < veteranEnds[v]; i++ {
				if factorRow(kinds[i]) != row {
					continue
				}
				chips.names = append(chips.names, names[i])
				chips.stars = append(chips.stars, starSuffix(levels[i]))
				chips.kinds = append(chips.kinds, factorType(kinds[i]))
				chips.inherited = append(chips.inherited, i >= ownEnds[v])
			}
			chips.rowEnds = append(chips.rowEnds, len(chips.names))
		}
		start = veteranEnds[v]
	}
	return chips
}

func (c *factorChips) veteranCount() int {
	return len(c.rowEnds) / factorRowCount
}

func (c *factorChips) row(veteran, row int) (start, end int) {
	i := veteran*factorRowCount + row
	if i > 0 {
		start = c.rowEnds[i-1]
	}
	return start, c.rowEnds[i]
}

// veteranLoad carries a veterans file chosen in the file picker into the window goroutine
type veteranLoad struct {
	chips factorChips
	err   error
}

// veteranPage lists the factors of every veteran in a veterans file the user picks
type veteranPage struct {
	dataStore  *data.DataStore
	panel      panel
	loadButton widget.Clickable
	choosing   bool
	loads      chan veteranLoad
	loadError  string
	chips      factorChips
	// Loaded veteran count, formatted when a file loads
	veteranCount string
	list         widget.List
	// Content width in pixels, stepped between contentMinPercent and contentMaxPercent
	width int
}

func newVeteranPage(dataStore *data.DataStore) *veteranPage {
	return &veteranPage{
		dataStore: dataStore,
		panel:     panel{title: veteranPanelTitle},
		loads:     make(chan veteranLoad, 1),
		list:      widget.List{List: layout.List{Axis: layout.Vertical}},
	}
}

func (p *veteranPage) update(gtx layout.Context, picker *explorer.Explorer, invalidate func()) {
	if p.loadButton.Clicked(gtx) && !p.choosing {
		p.choosing = true
		go p.chooseFile(picker, invalidate)
	}
	select {
	case result := <-p.loads:
		p.choosing = false
		switch {
		case errors.Is(result.err, explorer.ErrUserDecline):
		case result.err != nil:
			log.Printf("load veterans: %v", result.err)
			p.loadError = result.err.Error()
		default:
			p.loadError = ""
			p.chips = result.chips
			p.veteranCount = fmt.Sprintf("%d veterans", p.chips.veteranCount())
			p.list.Position = layout.Position{}
		}
	default:
	}
}

// chooseFile blocks on the file picker, then decodes and maps the chosen veterans file off the
// window goroutine
func (p *veteranPage) chooseFile(picker *explorer.Explorer, invalidate func()) {
	defer invalidate()
	file, err := picker.ChooseFile(".json")
	if err != nil {
		p.loads <- veteranLoad{err: err}
		return
	}
	defer file.Close()
	vets, err := veteran.Decode(file)
	if err != nil {
		p.loads <- veteranLoad{err: fmt.Errorf("load veterans: %w", err)}
		return
	}
	p.loads <- veteranLoad{chips: newFactorChips(p.dataStore, vets)}
}

func (p *veteranPage) layout(
	gtx layout.Context,
	theme *material.Theme,
	picker *explorer.Explorer,
	invalidate func(),
) layout.Dimensions {
	p.update(gtx, picker, invalidate)
	return p.panel.layout(gtx, theme, true, p.veteranCount,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutToolbar(gtx, theme)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if p.chips.veteranCount() == 0 {
						return layoutCenteredText(gtx, theme, emptyListText, colorTextMuted)
					}
					return p.layoutList(gtx, theme)
				}),
			)
		},
	)
}

// layoutToolbar draws the load button and the last load error above the veteran list
func (p *veteranPage) layoutToolbar(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	inset := layout.Inset{
		Top:    toolbarPadding,
		Bottom: toolbarPadding,
		Left:   veteranPadding,
		Right:  veteranPadding,
	}
	dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if p.choosing {
					gtx = gtx.Disabled()
				}
				button := material.Button(theme, &p.loadButton, loadButtonText)
				button.TextSize = smallTextSize
				button.Font.Weight = font.SemiBold
				button.CornerRadius = controlRadius
				button.Inset = layout.Inset{
					Top:    loadButtonPaddingY,
					Bottom: loadButtonPaddingY,
					Left:   loadButtonPaddingX,
					Right:  loadButtonPaddingX,
				}
				return button.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: loadErrorGap}.Layout),
			layout.Flexed(1, textLabel(theme, smallTextSize, p.loadError, colorError).Layout),
		)
	})
	line := gtx.Dp(borderWidth)
	fillRect(gtx, colorBorder, image.Rect(0, dims.Size.Y-line, dims.Size.X, dims.Size.Y))
	return dims
}

// layoutList centers the veteran list in the stepped content width
func (p *veteranPage) layoutList(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	available := gtx.Constraints.Max.X
	p.width = steppedWidth(p.width, available)
	defer op.Offset(image.Pt((available-p.width)/2, 0)).Push(gtx.Ops).Pop()
	gtx.Constraints.Min.X = p.width
	gtx.Constraints.Max.X = p.width
	list := material.List(theme, &p.list)
	styleScrollbar(&list.ScrollbarStyle)
	dims := list.Layout(gtx, p.chips.veteranCount(),
		func(gtx layout.Context, v int) layout.Dimensions {
			return p.layoutVeteran(gtx, theme, v)
		},
	)
	dims.Size.X = available
	return dims
}

// steppedWidth keeps width between contentMinPercent and contentMaxPercent of available. A
// width pushed outside that band snaps to the opposite bound, so the chip rows reflow at a few
// steps instead of on every resize.
func steppedWidth(width, available int) int {
	switch {
	case width*100 < available*contentMinPercent:
		return available * contentMaxPercent / 100
	case width*100 > available*contentMaxPercent:
		return available * contentMinPercent / 100
	}
	return width
}

// layoutVeteran draws veteran v's non-empty chip rows above a separator
func (p *veteranPage) layoutVeteran(gtx layout.Context, theme *material.Theme, v int) layout.Dimensions {
	padding, gap := gtx.Dp(veteranPadding), gtx.Dp(chipGap)
	width := gtx.Constraints.Max.X
	y := padding
	for row := range factorRowCount {
		start, end := p.chips.row(v, row)
		if start == end {
			continue
		}
		if y > padding {
			y += gap
		}
		offset := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
		y += p.layoutChipRow(gtx, theme, start, end)
		offset.Pop()
	}
	y += padding
	separator := image.Rect(0, y, width, y+gtx.Dp(borderWidth))
	fillRect(gtx, colorBorder, separator)
	return layout.Dimensions{Size: image.Pt(width, separator.Max.Y)}
}

// layoutChipRow places chips [start, end) left to right, wrapping at the maximum width, and
// returns the row height
func (p *veteranPage) layoutChipRow(gtx layout.Context, theme *material.Theme, start, end int) int {
	gap := gtx.Dp(chipGap)
	maxWidth := gtx.Constraints.Max.X
	gtx.Constraints.Min = image.Point{}
	var x, y, lineHeight int
	for i := start; i < end; i++ {
		macro := op.Record(gtx.Ops)
		size := p.layoutChip(gtx, theme, i)
		chip := macro.Stop()
		if x > 0 && x+size.X > maxWidth {
			x, y, lineHeight = 0, y+lineHeight+gap, 0
		}
		offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
		chip.Add(gtx.Ops)
		offset.Pop()
		x += size.X + gap
		lineHeight = max(lineHeight, size.Y)
	}
	return y + lineHeight
}

// layoutChip draws chip i as its name and star suffix on a rounded dark background
func (p *veteranPage) layoutChip(gtx layout.Context, theme *material.Theme, i int) image.Point {
	paddingX, paddingY := gtx.Dp(chipPaddingX), gtx.Dp(chipPaddingY)

	name := textLabel(theme, textSize, p.chips.names[i], factorColors[p.chips.kinds[i]])
	starColor := colorFactorOwnStars
	if p.chips.inherited[i] {
		starColor = colorFactorParentStar
	}
	stars := textLabel(theme, textSize, p.chips.stars[i], starColor)

	macro := op.Record(gtx.Ops)
	nameDims := name.Layout(gtx)
	starsOffset := op.Offset(image.Pt(nameDims.Size.X, 0)).Push(gtx.Ops)
	starsDims := stars.Layout(gtx)
	starsOffset.Pop()
	text := macro.Stop()

	size := image.Pt(
		nameDims.Size.X+starsDims.Size.X+2*paddingX,
		max(nameDims.Size.Y, starsDims.Size.Y)+2*paddingY,
	)
	background := clip.UniformRRect(image.Rectangle{Max: size}, size.Y/2)
	paint.FillShape(gtx.Ops, colorFactorBackground, background.Op(gtx.Ops))
	offset := op.Offset(image.Pt(paddingX, paddingY)).Push(gtx.Ops)
	text.Add(gtx.Ops)
	offset.Pop()
	return size
}
