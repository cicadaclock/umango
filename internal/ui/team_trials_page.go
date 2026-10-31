package ui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"slices"

	"gioui.org/f32"
	"gioui.org/io/key"
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

	vetRowHeight        = unit.Dp(32)
	skillRowHeight      = unit.Dp(28)
	fieldedMarkSize     = unit.Dp(14)
	toolbarPadding      = unit.Dp(6)
	toolbarLabelGap     = unit.Dp(8)
	splitGripLength     = unit.Dp(32)
	splitGripThickness  = unit.Dp(2)
	distanceFilterLabel = "Distance"
)

// Panels of the TT page, indexing teamTrialsPage.panels
const (
	vetPanel = iota
	histogramPanel
	skillPanel
	teamTrialsPanelCount
)

// Roles of the veterans chosen in the vet table, indexing teamTrialsPage.chosen
const (
	selectedVet = iota
	comparedVet
	chosenVetCount
)

// Views of the histogram panel, indexing teamTrialsPage.scoreViewTabs
const (
	distributionView = iota
	breakdownView
	scoreViewCount
)

var (
	vetColumns = [...]tableColumn{
		races.VetIdColumn:      {alignment: text.End},
		races.VetFieldedColumn: {icon: fieldedMarkSize, alignment: text.Middle},
		races.VetNameColumn:    {fill: true},
		races.VetRacesColumn:   {alignment: text.End},
		races.VetMaxColumn:     {alignment: text.End},
		races.VetAvgColumn:     {alignment: text.End},
		races.VetSpurtColumn:   {alignment: text.End},
	}
	// Procs always shows four digits and Rate two decimals; Name takes the rest of the sidebar
	skillColumns = [...]tableColumn{
		races.SkillNameColumn:  {fill: true},
		races.SkillProcsColumn: {sample: "0000", alignment: text.End},
		races.SkillRateColumn:  {sample: "0.00", alignment: text.End},
	}
	// Highlights of the skill tabs of the chosen veterans, by role
	skillTabHighlights = [chosenVetCount]highlightColors{
		selectedVet: selectionHighlight,
		comparedVet: comparisonHighlight,
	}
	// A role no veteran is chosen for
	noVet = chosenVet{index: -1}
	// Tab label and panel title of each histogram panel view
	scoreViewNames = [scoreViewCount]string{
		distributionView: "Distribution",
		breakdownView:    "Breakdown",
	}
	scoreViewTitles = [scoreViewCount]string{
		distributionView: "Score Distribution",
		breakdownView:    "Score Breakdown",
	}
)

// chosenVet is a veteran chosen in the vet table
type chosenVet struct {
	// Summary index, -1 when no veteran is chosen
	index int
	id    int
	name  string
}

// loadTeamTrialResults loads the saved team trial races from the game's default save directory
func loadTeamTrialResults() (races.TeamTrialResultSet, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return races.TeamTrialResultSet{}, fmt.Errorf("get home dir: %w", err)
	}
	return races.LoadRacesFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
}

// teamTrialsPage shows the TT veteran table beside a sidebar of the selected veteran's score
// histogram or score breakdown, and skill activations
//
// Clicking a row selects its veteran. Ctrl+clicking another row compares its veteran to the
// selected one: the histogram outlines it against the selected veteran, the breakdown lists it
// beneath, and tabs over the skill table switch between the two. Ctrl+clicking the compared row
// again or pressing Escape ends the comparison.
type teamTrialsPage struct {
	summary   races.TTUmaSummary
	dataStore *data.DataStore
	// All veterans, sorted by the clicked vet table header
	vetTable *races.Table[races.VetTableData]
	// The vetTable rows that pass distanceFilter, in vetTable's order; vetTable itself when no
	// filter is active
	vetShown *races.Table[races.VetTableData]
	// Scratch buffer of the vetTable rows in vetShown
	vetRows    []int
	skillTable *races.Table[races.SkillTableData]
	histogram  scoreHistogram
	breakdown  scoreBreakdown
	// Tabs over the histogram panel, and the view they show
	scoreViewTabs  [scoreViewCount]widget.Clickable
	scoreView      int
	distanceFilter distanceFilter
	vetView        tableView
	skillView      tableView
	sidebarSplit   component.Resize
	panels         [teamTrialsPanelCount]panel
	focus          int
	chosen         [chosenVetCount]chosenVet
	// Tabs over the skill table while comparing, by role
	skillTabs [chosenVetCount]widget.Clickable
	// Role of the chosen veteran whose skills the skill table shows
	skillVet int
	// Panel captions, formatted when the rows or the chosen veterans change
	vetCount, histogramCaption, skillCaption string
}

// newTeamTrialsPage aggregates the loaded races and selects the first vet table row. It does
// not touch the window, so it runs on the loader goroutine.
func newTeamTrialsPage(dataStore *data.DataStore, resultSet races.TeamTrialResultSet) *teamTrialsPage {
	summary := resultSet.Summarize(dataStore)
	p := &teamTrialsPage{
		summary:      summary,
		dataStore:    dataStore,
		vetTable:     races.NewVetTable(dataStore, summary),
		skillTable:   races.NewSkillTable(),
		histogram:    newScoreHistogram(summary),
		breakdown:    newScoreBreakdown(summary),
		sidebarSplit: component.Resize{Axis: layout.Vertical, Ratio: histogramRatio},
		panels: [teamTrialsPanelCount]panel{
			vetPanel:       {title: "Veterans"},
			histogramPanel: {title: scoreViewTitles[distributionView]},
			skillPanel:     {title: "Skill Activations"},
		},
		chosen: [chosenVetCount]chosenVet{noVet, noVet},
	}
	p.vetView = newTableView(p.vetTable.Headers(), vetColumns[:], vetRowHeight, true)
	p.vetView.layoutIcon = p.layoutFieldedMark
	p.showVetRows()
	p.skillView = newTableView(p.skillTable.Headers(), skillColumns[:], skillRowHeight, false)
	p.syncSkillView()
	if p.vetShown.Len() > 0 {
		p.selectRow(0)
	}
	return p
}

// showVetRows shows the vetTable rows that pass the distance filter, keeping the chosen
// veterans highlighted when they are among them
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
	p.markChosenRows()
}

// markChosenRows highlights the shown vet table rows of the chosen veterans
func (p *teamTrialsPage) markChosenRows() {
	p.vetView.selectedRow = p.shownRow(selectedVet)
	p.vetView.comparedRow = p.shownRow(comparedVet)
}

// shownRow returns the shown vet table row of the chosen veteran of role, -1 when none is
// chosen or the distance filter hides it
func (p *teamTrialsPage) shownRow(role int) int {
	if p.chosen[role].index < 0 {
		return -1
	}
	return slices.Index(p.vetShown.Data().TrainedCharaIds, p.chosen[role].id)
}

// rowVet returns the veteran in the shown vet table row
func (p *teamTrialsPage) rowVet(row int) chosenVet {
	shown := p.vetShown.Data()
	id := shown.GetTrainedCharaId(row)
	return chosenVet{
		index: slices.Index(p.summary.TrainedCharaIds, id),
		id:    id,
		name:  shown.Names[row],
	}
}

// selectRow selects the veteran in the shown vet table row, ending a comparison with itself
func (p *teamTrialsPage) selectRow(row int) {
	p.chosen[selectedVet] = p.rowVet(row)
	if p.chosen[comparedVet].index == p.chosen[selectedVet].index {
		p.chosen[comparedVet] = noVet
	}
	p.showChosen(selectedVet)
}

// compareRow compares the veteran in the shown vet table row to the selected one, or ends the
// comparison when it is already compared
func (p *teamTrialsPage) compareRow(row int) {
	vet := p.rowVet(row)
	switch vet.index {
	case p.chosen[selectedVet].index:
		// A veteran is not compared to itself
	case p.chosen[comparedVet].index:
		p.endComparison()
	default:
		p.chosen[comparedVet] = vet
		p.showChosen(comparedVet)
	}
}

// endComparison returns to showing the selected veteran alone
func (p *teamTrialsPage) endComparison() {
	p.chosen[comparedVet] = noVet
	p.showChosen(selectedVet)
}

// releaseHiddenVets ends a comparison with a veteran the distance filter hides, and gives a
// hidden selected veteran's place to the first shown one
func (p *teamTrialsPage) releaseHiddenVets() {
	if p.chosen[comparedVet].index >= 0 && p.vetView.comparedRow < 0 {
		p.endComparison()
	}
	if p.vetView.selectedRow < 0 && p.vetShown.Len() > 0 {
		p.selectRow(0)
	}
}

// showChosen shows the chosen veterans in the vet table highlights, the histogram and the
// breakdown, and the skill activations of the one of role skillRole
func (p *teamTrialsPage) showChosen(skillRole int) {
	selected, compared := p.chosen[selectedVet], p.chosen[comparedVet]
	p.markChosenRows()
	p.histogram.show(selected.index, compared.index)
	p.breakdown.show(selected.index, compared.index)
	p.histogramCaption = selected.name
	if compared.index >= 0 {
		p.histogramCaption = selected.name + " vs " + compared.name
	}
	p.showSkills(skillRole)
	p.syncSkillView()
	p.skillView.scrollToTop()
}

// showSkills fills the skill table, in its original order, with the skill activations of the
// chosen veteran of role
func (p *teamTrialsPage) showSkills(role int) {
	index := p.chosen[role].index
	raceCount := p.summary.Scores[index].Len()
	p.skillVet = role
	p.skillCaption = formatRaceCount(raceCount)
	races.RefillSkillTable(p.skillTable, p.dataStore, p.summary.SkillActivations[index], raceCount)
}

// switchSkills shows the skill activations of the chosen veteran of role in the skill table's
// current sort
func (p *teamTrialsPage) switchSkills(role int) {
	col, state := p.skillTable.SortedColumn()
	p.showSkills(role)
	p.skillTable.SortBy(col, state)
	p.syncSkillView()
	p.skillView.scrollToTop()
}

// syncSkillView renders the skill table's rows and sort state into the skill view
func (p *teamTrialsPage) syncSkillView() {
	p.skillView.setColumns(p.skillTable.Columns())
	p.skillView.sortColumn, p.skillView.sortState = p.skillTable.SortedColumn()
}

// layoutFieldedMark draws the check mark of a fielded veteran's row in the Fielded column, the
// vet table's only icon column
func (p *teamTrialsPage) layoutFieldedMark(gtx layout.Context, row, _ int) {
	if p.vetShown.Data().Fielded[row] {
		drawCheckMark(gtx, colorAccent)
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
	sortColumn, row, modifiers := p.vetView.update(gtx)
	switch {
	case row < 0:
	case modifiers.Contain(key.ModShortcut):
		p.compareRow(row)
	default:
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
		p.releaseHiddenVets()
	}
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if e, ok := e.(key.Event); ok && e.State == key.Press && p.chosen[comparedVet].index >= 0 {
			p.endComparison()
		}
	}
	for view := range p.scoreViewTabs {
		if p.scoreViewTabs[view].Clicked(gtx) {
			p.scoreView = view
			p.panels[histogramPanel].title = scoreViewTitles[view]
		}
	}
	for role := range p.skillTabs {
		if p.skillTabs[role].Clicked(gtx) && role != p.skillVet && p.chosen[role].index >= 0 {
			p.switchSkills(role)
		}
	}
	if sortColumn, _, _ := p.skillView.update(gtx); sortColumn >= 0 {
		p.skillTable.Sort(sortColumn)
		p.syncSkillView()
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
						return layoutTableToolbar(gtx, func(gtx layout.Context) layout.Dimensions {
							return p.layoutVetToolbar(gtx, theme)
						})
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

// layoutSidebar stacks the histogram or the breakdown, under tabs choosing between them, over
// the skill table, split by a draggable gutter, with tabs over the skill table while comparing
func (p *teamTrialsPage) layoutSidebar(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	p.sidebarSplit.Ratio = min(max(p.sidebarSplit.Ratio, minHistogramRatio), maxHistogramRatio)
	return p.sidebarSplit.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			return p.layoutPanel(gtx, theme, histogramPanel, p.histogramCaption,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layoutTableToolbar(gtx, func(gtx layout.Context) layout.Dimensions {
								return p.layoutScoreViewTabs(gtx, theme)
							})
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							if p.scoreView == breakdownView {
								return p.breakdown.layout(gtx, theme)
							}
							return p.histogram.layout(gtx, theme)
						}),
					)
				},
			)
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.layoutPanel(gtx, theme, skillPanel, p.skillCaption,
				func(gtx layout.Context) layout.Dimensions {
					if p.chosen[comparedVet].index < 0 {
						return p.skillView.layout(gtx, theme)
					}
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layoutTableToolbar(gtx, func(gtx layout.Context) layout.Dimensions {
								return p.layoutSkillTabs(gtx, theme)
							})
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return p.skillView.layout(gtx, theme)
						}),
					)
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

// layoutTableToolbar draws content in a padded strip across the constraints, ruled off from the
// table below it
func layoutTableToolbar(gtx layout.Context, content layout.Widget) layout.Dimensions {
	inset := layout.Inset{
		Top:    toolbarPadding,
		Bottom: toolbarPadding,
		Left:   tableCellPadding,
		Right:  tableCellPadding,
	}
	dims := inset.Layout(gtx, content)
	dims.Size.X = gtx.Constraints.Max.X
	line := gtx.Dp(borderWidth)
	fillRect(gtx, colorBorder, image.Rect(0, dims.Size.Y-line, dims.Size.X, dims.Size.Y))
	return dims
}

// layoutVetToolbar draws the distance filter and the shown veteran count
func (p *teamTrialsPage) layoutVetToolbar(
	gtx layout.Context,
	theme *material.Theme,
) layout.Dimensions {
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
}

// layoutSkillTabs draws a pill named for each chosen veteran, at most half the width each, that
// shows its skills in the skill table, highlighted in its veteran's color while active
func (p *teamTrialsPage) layoutSkillTabs(
	gtx layout.Context,
	theme *material.Theme,
) layout.Dimensions {
	tabWidth := max((gtx.Constraints.Max.X-gtx.Dp(filterGap))/2, 0)
	layoutTab := func(role int) layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = tabWidth
			var highlight highlightColors
			if role == p.skillVet {
				highlight = skillTabHighlights[role]
			}
			return layoutPill(gtx, theme, &p.skillTabs[role], p.chosen[role].name, highlight)
		}
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(layoutTab(selectedVet)),
		layout.Rigid(layout.Spacer{Width: filterGap}.Layout),
		layout.Rigid(layoutTab(comparedVet)),
	)
}

// layoutScoreViewTabs draws a pill per histogram panel view, highlighting the shown one
func (p *teamTrialsPage) layoutScoreViewTabs(
	gtx layout.Context,
	theme *material.Theme,
) layout.Dimensions {
	layoutTab := func(view int) layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			var highlight highlightColors
			if view == p.scoreView {
				highlight = selectionHighlight
			}
			return layoutPill(gtx, theme, &p.scoreViewTabs[view], scoreViewNames[view], highlight)
		}
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(layoutTab(distributionView)),
		layout.Rigid(layout.Spacer{Width: filterGap}.Layout),
		layout.Rigid(layoutTab(breakdownView)),
	)
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
