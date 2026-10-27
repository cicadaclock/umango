package ui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"slices"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/races"
)

const (
	// The sidebar keeps a fixed width so its histogram and skill table never reflow
	sidebarWidth = unit.Dp(400)
	// Initial share and bounds of the sidebar height given to the histogram
	histogramRatio, minHistogramRatio, maxHistogramRatio = 0.45, 0.2, 0.8

	vetRowHeight         = unit.Dp(32)
	skillRowHeight       = unit.Dp(28)
	runningStyleIconSize = unit.Dp(26)
	distanceIconSize     = unit.Dp(20)
	fieldedMarkSize      = unit.Dp(14)
	toolbarPadding       = unit.Dp(6)
	toolbarLabelGap      = unit.Dp(8)
	splitGripLength      = unit.Dp(32)
	splitGripThickness   = unit.Dp(2)
	distanceFilterLabel  = "Distance"
)

// Panels of the TT page, indexing teamTrialsPage.panels
const (
	vetPanel = iota
	histogramPanel
	skillPanel
	teamTrialsPanelCount
)

var (
	vetColumns = [...]tableColumn{
		races.VetIdColumn:       {alignment: text.End},
		races.VetFieldedColumn:  {icon: fieldedMarkSize, iconOnly: true, alignment: text.Middle},
		races.VetNameColumn:     {fill: true},
		races.VetDistanceColumn: {icon: distanceIconSize},
		races.VetStyleColumn:    {icon: runningStyleIconSize, iconOnly: true, alignment: text.Middle},
		races.VetRacesColumn:    {alignment: text.End},
		races.VetMaxColumn:      {alignment: text.End},
		races.VetAvgColumn:      {alignment: text.End},
	}
	// Procs always shows four digits and Rate two decimals; Name takes the rest of the sidebar
	skillColumns = [...]tableColumn{
		races.SkillNameColumn:  {fill: true},
		races.SkillProcsColumn: {sample: "0000", alignment: text.End},
		races.SkillRateColumn:  {sample: "0.00", alignment: text.End},
	}
)

// loadTeamTrialResults loads the saved team trial races from the game's default save directory
func loadTeamTrialResults() (races.TeamTrialResultSet, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return races.TeamTrialResultSet{}, fmt.Errorf("get home dir: %w", err)
	}
	return races.LoadRacesFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
}

// teamTrialsPage shows the TT veteran table beside a sidebar of the selected veteran's score
// histogram and skill activations
type teamTrialsPage struct {
	summary   races.TTUmaSummary
	dataStore *data.DataStore
	// All veterans, sorted by the clicked vet table header
	vetTable *races.Table[races.VetTableData]
	// The vetTable rows that pass distanceFilter, in vetTable's order; vetTable itself when no
	// filter is active
	vetShown *races.Table[races.VetTableData]
	// Scratch buffer of the vetTable rows in vetShown
	vetRows        []int
	skillTable     *races.Table[races.SkillTableData]
	histogram      scoreHistogram
	distanceFilter distanceFilter
	vetView        tableView
	skillView      tableView
	sidebarSplit   component.Resize
	panels         [teamTrialsPanelCount]panel
	focus          int
	// Panel captions, formatted when the rows or the selection change
	vetCount, selectedName, selectedRaces string
	// Summary index of the selected veteran, -1 when none is selected
	selectedIndex int
	selectedId    int
}

// newTeamTrialsPage aggregates the loaded races and selects the first vet table row. It does
// not touch the window, so it runs on the loader goroutine.
func newTeamTrialsPage(
	dataStore *data.DataStore,
	resultSet races.TeamTrialResultSet,
	distanceIcons distanceTypeIcons,
) *teamTrialsPage {
	summary := resultSet.Summarize()
	p := &teamTrialsPage{
		summary:        summary,
		dataStore:      dataStore,
		vetTable:       races.NewVetTable(dataStore, summary),
		skillTable:     races.NewSkillTable(),
		histogram:      newScoreHistogram(summary),
		distanceFilter: distanceFilter{icons: distanceIcons},
		sidebarSplit:   component.Resize{Axis: layout.Vertical, Ratio: histogramRatio},
		panels: [teamTrialsPanelCount]panel{
			vetPanel:       {title: "Veterans"},
			histogramPanel: {title: "Score Distribution"},
			skillPanel:     {title: "Skill Activations"},
		},
		selectedIndex: -1,
	}
	p.vetView = newTableView(p.vetTable.Headers(), vetColumns[:], vetRowHeight, true)
	p.vetView.layoutIcon = p.layoutVetIcon
	p.showVetRows()
	p.skillView = newTableView(p.skillTable.Headers(), skillColumns[:], skillRowHeight, false)
	p.skillView.setColumns(p.skillTable.Columns())
	if p.vetShown.Len() > 0 {
		p.selectRow(0)
	}
	return p
}

// showVetRows shows the vetTable rows that pass the distance filter, keeping the selected
// veteran highlighted when it is among them
func (p *teamTrialsPage) showVetRows() {
	p.vetShown = p.vetTable
	p.vetCount = fmt.Sprintf("%d veterans", p.vetTable.Len())
	if p.distanceFilter.selected != races.NoneDistance {
		p.vetRows = p.vetRows[:0]
		for row, distance := range p.vetTable.Data().DistanceTypes {
			if distance == p.distanceFilter.selected {
				p.vetRows = append(p.vetRows, row)
			}
		}
		p.vetShown = p.vetTable.Filter(p.vetRows)
		p.vetCount = fmt.Sprintf("%d of %d veterans", p.vetShown.Len(), p.vetTable.Len())
	}
	p.vetView.setColumns(p.vetShown.Columns())
	p.vetView.selectedRow = slices.Index(p.vetShown.Data().TrainedCharaIds, p.selectedId)
}

// selectRow shows the histogram and skill activations of the veteran in the shown vet table row
func (p *teamTrialsPage) selectRow(row int) {
	shown := p.vetShown.Data()
	p.selectedId = shown.GetTrainedCharaId(row)
	p.selectedIndex = slices.Index(p.summary.TrainedCharaIds, p.selectedId)
	p.selectedName = shown.Names[row]
	raceCount := p.summary.Scores[p.selectedIndex].Len()
	p.selectedRaces = formatRaceCount(raceCount)
	p.vetView.selectedRow = row
	races.RefillSkillTable(
		p.skillTable,
		p.dataStore,
		p.summary.SkillActivations[p.selectedIndex],
		raceCount,
	)
	p.skillView.setColumns(p.skillTable.Columns())
	p.skillView.sortColumn, p.skillView.sortState = p.skillTable.SortedColumn()
	p.skillView.scrollToTop()
}

// layoutVetIcon draws the fielded mark, distance type icon or running style icon of a vet
// table cell
func (p *teamTrialsPage) layoutVetIcon(
	gtx layout.Context,
	theme *material.Theme,
	row, col int,
) {
	shown := p.vetShown.Data()
	switch col {
	case races.VetFieldedColumn:
		if shown.Fielded[row] {
			drawCheckMark(gtx, colorAccent)
		}
	case races.VetDistanceColumn:
		if icon := p.distanceFilter.icons[shown.DistanceTypes[row]]; iconLoaded(icon) {
			widget.Image{Src: icon, Fit: widget.Contain}.Layout(gtx)
		}
	case races.VetStyleColumn:
		layoutRunningStyleIcon(gtx, theme, shown.StyleTypes[row])
	}
}

// drawCheckMark strokes a check mark across the square given by the exact constraints
func drawCheckMark(gtx layout.Context, c color.NRGBA) {
	side := float32(gtx.Constraints.Max.X)
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(side*0.14, side*0.54))
	path.LineTo(f32.Pt(side*0.40, side*0.78))
	path.LineTo(f32.Pt(side*0.88, side*0.24))
	paint.FillShape(gtx.Ops, c, clip.Stroke{Path: path.End(), Width: side * 0.16}.Op())
}

func (p *teamTrialsPage) update(gtx layout.Context) {
	p.focus = focusPanel(gtx, p.panels[:], p.focus)
	sortColumn, row := p.vetView.update(gtx)
	if row >= 0 {
		p.selectRow(row)
	}
	if sortColumn >= 0 {
		p.vetTable.Sort(sortColumn)
		p.vetView.sortColumn, p.vetView.sortState = p.vetTable.SortedColumn()
		p.showVetRows()
	}
	if p.distanceFilter.update(gtx) {
		p.showVetRows()
		p.vetView.scrollToTop()
		// A veteran the filter hides gives way to the first shown one
		if p.vetView.selectedRow < 0 && p.vetShown.Len() > 0 {
			p.selectRow(0)
		}
	}
	if sortColumn, _ := p.skillView.update(gtx); sortColumn >= 0 {
		p.skillTable.Sort(sortColumn)
		p.skillView.sortColumn, p.skillView.sortState = p.skillTable.SortedColumn()
		p.skillView.columns = p.skillTable.Columns()
	}
}

func (p *teamTrialsPage) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	p.update(gtx)
	sidebar := min(gtx.Dp(sidebarWidth), gtx.Constraints.Max.X/2)
	return layout.Flex{}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return p.layoutPanel(gtx, theme, vetPanel, "", func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.layoutVetToolbar(gtx, theme)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return p.vetView.layout(gtx, theme)
					}),
				)
			})
		}),
		layout.Rigid(layout.Spacer{Width: gutterWidth}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(sidebar, gtx.Constraints.Max.Y))
			return p.layoutSidebar(gtx, theme)
		}),
	)
}

// layoutSidebar stacks the histogram over the skill table, split by a draggable gutter
func (p *teamTrialsPage) layoutSidebar(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	p.sidebarSplit.Ratio = min(max(p.sidebarSplit.Ratio, minHistogramRatio), maxHistogramRatio)
	return p.sidebarSplit.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			return p.layoutPanel(gtx, theme, histogramPanel, p.selectedName,
				func(gtx layout.Context) layout.Dimensions {
					return p.histogram.layout(gtx, theme, p.selectedIndex)
				},
			)
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.layoutPanel(gtx, theme, skillPanel, p.selectedRaces,
				func(gtx layout.Context) layout.Dimensions {
					return p.skillView.layout(gtx, theme)
				},
			)
		},
		layoutSplitHandle,
	)
}

func (p *teamTrialsPage) layoutPanel(
	gtx layout.Context,
	theme *material.Theme,
	index int,
	detail string,
	body layout.Widget,
) layout.Dimensions {
	return p.panels[index].layout(gtx, theme, p.focus == index, detail, body)
}

// layoutVetToolbar draws the distance filter and the shown veteran count above the vet table
func (p *teamTrialsPage) layoutVetToolbar(
	gtx layout.Context,
	theme *material.Theme,
) layout.Dimensions {
	inset := layout.Inset{
		Top:    toolbarPadding,
		Bottom: toolbarPadding,
		Left:   tableCellPadding,
		Right:  tableCellPadding,
	}
	dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(textLabel(theme, smallTextSize, distanceFilterLabel, colorTextMuted).Layout),
			layout.Rigid(layout.Spacer{Width: toolbarLabelGap}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.distanceFilter.layout(gtx, theme)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				label := textLabel(theme, smallTextSize, p.vetCount, colorTextMuted)
				label.Alignment = text.End
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return label.Layout(gtx)
			}),
		)
	})
	line := gtx.Dp(borderWidth)
	fillRect(gtx, colorBorder, image.Rect(0, dims.Size.Y-line, dims.Size.X, dims.Size.Y))
	return dims
}

// layoutSplitHandle draws the gutter between the stacked sidebar panels, with a grip at its
// center
func layoutSplitHandle(gtx layout.Context) layout.Dimensions {
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(gutterWidth))
	grip := image.Rectangle{Max: image.Pt(gtx.Dp(splitGripLength), gtx.Dp(splitGripThickness))}
	grip = grip.Add(size.Sub(grip.Max).Div(2))
	paint.FillShape(gtx.Ops, colorBorderStrong, clip.UniformRRect(grip, grip.Dy()/2).Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}
