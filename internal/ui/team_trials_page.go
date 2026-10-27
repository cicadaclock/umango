package ui

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/races"
)

const (
	vetTableRatio    = 0.7
	histogramRatio   = 0.45
	splitHandleWidth = unit.Dp(6)
)

// loadTeamTrialResults loads the saved team trial races from the game's default save directory
func loadTeamTrialResults() (races.TeamTrialResultSet, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return races.TeamTrialResultSet{}, fmt.Errorf("get home dir: %w", err)
	}
	return races.LoadRacesFolder(filepath.Join(home, "Documents", "Saved races", "Team trials"))
}

// teamTrialsPage shows the TT veteran table beside the selected veteran's score histogram and
// skill activations
type teamTrialsPage struct {
	summary    races.TTUmaSummary
	dataStore  *data.DataStore
	vetTable   *races.Table[races.VetTableData]
	skillTable *races.Table[races.SkillTableData]
	histogram  scoreHistogram
	vetView    tableView
	skillView  tableView
	split      component.Resize
	rightSplit component.Resize
	// Summary index of the selected veteran, -1 when none is selected
	selectedIndex int
	selectedId    int
}

// newTeamTrialsPage aggregates the loaded races and selects the first vet table row. It does
// not touch the window, so it runs on the loader goroutine.
func newTeamTrialsPage(dataStore *data.DataStore, resultSet races.TeamTrialResultSet) *teamTrialsPage {
	summary := resultSet.Summarize()
	p := &teamTrialsPage{
		summary:       summary,
		dataStore:     dataStore,
		vetTable:      races.NewVetTable(dataStore, summary),
		skillTable:    races.NewSkillTable(),
		histogram:     newScoreHistogram(summary),
		split:         component.Resize{Axis: layout.Horizontal, Ratio: vetTableRatio},
		rightSplit:    component.Resize{Axis: layout.Vertical, Ratio: histogramRatio},
		selectedIndex: -1,
	}
	p.vetView = newTableView(p.vetTable.Headers(), true)
	p.vetView.setColumns(p.vetTable.Columns(), p.vetTable.ColumnWidths())
	p.vetView.layoutCell = p.layoutVetCell
	p.skillView = newTableView(p.skillTable.Headers(), false)
	p.skillView.setColumns(p.skillTable.Columns(), p.skillTable.ColumnWidths())
	if p.vetTable.Len() > 0 {
		p.selectRow(0)
	}
	return p
}

// selectRow shows the histogram and skill activations of the veteran in the vet table row
func (p *teamTrialsPage) selectRow(row int) {
	p.selectedId = p.vetTable.Data().GetTrainedCharaId(row)
	p.selectedIndex = slices.Index(p.summary.TrainedCharaIds, p.selectedId)
	p.vetView.selectedRow = row
	races.RefillSkillTable(
		p.skillTable,
		p.dataStore,
		p.summary.SkillActivations[p.selectedIndex],
		p.summary.Scores[p.selectedIndex].Len(),
	)
	p.skillView.setColumns(p.skillTable.Columns(), p.skillTable.ColumnWidths())
	p.skillView.scrollToTop()
}

// layoutVetCell draws the running style icon in place of the vet table's style text
func (p *teamTrialsPage) layoutVetCell(
	gtx layout.Context,
	theme *material.Theme,
	row, col int,
) bool {
	if col != races.VetStyleColumn {
		return false
	}
	gtx, offset := iconCell(gtx)
	layoutRunningStyleIcon(gtx, theme, p.vetTable.Data().StyleTypes[row])
	offset.Pop()
	return true
}

func (p *teamTrialsPage) update(gtx layout.Context) {
	sortColumn, row := p.vetView.update(gtx)
	if row >= 0 {
		p.selectRow(row)
	}
	if sortColumn >= 0 {
		p.vetTable.Sort(sortColumn)
		p.vetView.columns = p.vetTable.Columns()
		p.vetView.selectedRow = slices.Index(p.vetTable.Data().TrainedCharaIds, p.selectedId)
	}
	if sortColumn, _ := p.skillView.update(gtx); sortColumn >= 0 {
		p.skillTable.Sort(sortColumn)
		p.skillView.columns = p.skillTable.Columns()
	}
}

func (p *teamTrialsPage) layout(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	p.update(gtx)
	return p.split.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			return p.vetView.layout(gtx, theme)
		},
		func(gtx layout.Context) layout.Dimensions {
			return p.rightSplit.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					return p.histogram.layout(gtx, theme, p.selectedIndex)
				},
				func(gtx layout.Context) layout.Dimensions {
					return p.skillView.layout(gtx, theme)
				},
				verticalSplitHandle,
			)
		},
		horizontalSplitHandle,
	)
}

func horizontalSplitHandle(gtx layout.Context) layout.Dimensions {
	return layoutSplitHandle(gtx, layout.Horizontal)
}

func verticalSplitHandle(gtx layout.Context) layout.Dimensions {
	return layoutSplitHandle(gtx, layout.Vertical)
}

// layoutSplitHandle draws the draggable bar between the panes of a component.Resize
func layoutSplitHandle(gtx layout.Context, axis layout.Axis) layout.Dimensions {
	size := axis.Convert(image.Pt(gtx.Dp(splitHandleWidth), axis.Convert(gtx.Constraints.Max).Y))
	paint.FillShape(gtx.Ops, colorDivider, clip.Rect{Max: size}.Op())
	return layout.Dimensions{Size: size}
}
