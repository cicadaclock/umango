package ui

import (
	"cmp"
	"fmt"
	"log"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/db"
	"github.com/cicadaclock/umango/internal/races"
	"github.com/cicadaclock/umango/internal/racesim"
)

const (
	skillAnalysisPanelTitle = "Skill Analysis"
	startAnalysisText       = "Start"
	stopAnalysisText        = "Stop"
	selectCategoryText      = "Select a category, then an uma on that team"
	selectUmaText           = "Select an uma on the team"
	readyToStartText        = "Start the analysis to simulate each skill on every course of the pool"
	firstResultText         = "The first results appear when each course has run %d races"
	missingEngineText       = "The race engine files are missing, as the Simulator tab lists"
	findingSkillsText       = "Finding skills that can activate…"
	skillAnalysisCaption    = "Proc%: share of races in which the skill activated. Win: change of the win rate. " +
		"Score per SP: change of the team trial score before bonuses, for each skill point."
	skillAnalysisRowHeight = unit.Dp(26)
	// Space between the groups of controls in the toolbar
	skillAnalysisGroupGap = unit.Dp(14)
	skillAnalysisCaptionY = unit.Dp(6)
	// Runners on one team
	teamRunnerCount = 3
	// Column of the score per skill point, by which the table first sorts
	scorePerPointColumn = 4
)

// skillAnalysisRow holds the values of one skill in the skill analysis table
type skillAnalysisRow struct {
	name string
	// Skill points that the uma would spend to learn the skill, before any hint discount
	skillPoints int
	// Change of the team trial score before bonuses for each skill point
	scorePerPoint float64
	summary       racesim.SkillSummary
}

// skillAnalysisColumns defines the skill analysis table. A column without a key sorts by skill name.
var skillAnalysisColumns = [...]struct {
	header string
	style  tableColumn
	key    func(row *skillAnalysisRow) float64
	// The first click on the header sorts in ascending order, which suits text and costs
	ascending bool
	cell      func(row *skillAnalysisRow) string
}{
	{"Skill", tableColumn{fill: true, fillFitsCells: true}, nil, true,
		func(row *skillAnalysisRow) string { return row.name }},
	{"SP", tableColumn{alignment: text.End, sample: "9999"},
		func(row *skillAnalysisRow) float64 { return float64(row.skillPoints) }, true,
		func(row *skillAnalysisRow) string { return strconv.Itoa(row.skillPoints) }},
	{"Proc%", tableColumn{alignment: text.End, sample: "100.0%"},
		func(row *skillAnalysisRow) float64 { return row.summary.ProcRate }, false,
		func(row *skillAnalysisRow) string { return fmt.Sprintf("%.1f%%", row.summary.ProcRate*100) }},
	{"Win", tableColumn{alignment: text.End, sample: "+00.0%"},
		func(row *skillAnalysisRow) float64 { return row.summary.WinDelta }, false,
		func(row *skillAnalysisRow) string { return fmt.Sprintf("%+.1f%%", row.summary.WinDelta*100) }},
	{"Score per SP", tableColumn{alignment: text.End, sample: "+00.00"},
		func(row *skillAnalysisRow) float64 { return row.scorePerPoint }, false,
		func(row *skillAnalysisRow) string { return fmt.Sprintf("%+.2f", row.scorePerPoint) }},
}

// teamRoundLoad carries the latest round of a category to the window goroutine
type teamRoundLoad struct {
	category races.DistanceType
	round    *racesim.TeamRound
	err      error
}

// skillAnalysisUpdate carries the state of a running analysis to the window goroutine
type skillAnalysisUpdate struct {
	report racesim.SkillSurveyReport
	// Rows of the last finished batch
	rows []skillAnalysisRow
	// The analysis has ended, and err holds its failure if any
	finished bool
	err      error
}

// skillAnalysisPage simulates each skill that one uma of a team could learn on every course of the
// team trial pool of a category, and tabulates how often each skill activates and what it gains.
// The analysis runs in batches, and the table shows the result of each finished batch.
type skillAnalysisPage struct {
	dataStore  *data.DataStore
	engines    *engineHolder
	resultsDir string
	panel      panel
	categories distanceFilter
	// Round of the selected category in the latest team trial result, nil until it is read
	round         *racesim.TeamRound
	runnerButtons [teamRunnerCount]widget.Clickable
	// Team index of the selected uma, -1 when none is selected
	selectedRunner int
	startButton    widget.Clickable
	table          tableView
	rows           []skillAnalysisRow
	baseline       racesim.SkillSurveyBaseline
	report         racesim.SkillSurveyReport
	errorText      string
	// Category and uma of the analysis that the table shows
	analysisTitle string
	// Races per course behind the rows
	racesShown int
	running    bool
	startTime  time.Time
	stop       *atomic.Bool
	roundLoads chan teamRoundLoad
	updates    chan skillAnalysisUpdate
}

// newSkillAnalysisPage returns a skill analysis page that reads team trial results from a folder
func newSkillAnalysisPage(dataStore *data.DataStore, engines *engineHolder, resultsDir string) *skillAnalysisPage {
	headers := make([]string, len(skillAnalysisColumns))
	styles := make([]tableColumn, len(skillAnalysisColumns))
	for col := range skillAnalysisColumns {
		headers[col] = skillAnalysisColumns[col].header
		styles[col] = skillAnalysisColumns[col].style
	}
	p := &skillAnalysisPage{
		dataStore:      dataStore,
		engines:        engines,
		resultsDir:     resultsDir,
		panel:          panel{title: skillAnalysisPanelTitle},
		selectedRunner: -1,
		table:          newTableView(headers, styles, skillAnalysisRowHeight, false),
		roundLoads:     make(chan teamRoundLoad, 1),
		updates:        make(chan skillAnalysisUpdate, 1),
	}
	p.table.sortColumn, p.table.sortState = scorePerPointColumn, races.Descending
	return p
}

// update applies the clicks since the last frame and the latest results of the background jobs
func (p *skillAnalysisPage) update(gtx layout.Context, invalidate func()) {
	if !p.running {
		if p.categories.update(gtx) {
			p.selectCategory(invalidate)
		}
		for i := range p.runnerButtons {
			if p.runnerButtons[i].Clicked(gtx) {
				p.selectedRunner = i
			}
		}
	}
	if p.startButton.Clicked(gtx) {
		if p.running {
			p.stop.Store(true)
		} else if p.round != nil && p.selectedRunner >= 0 {
			p.start(invalidate)
		}
	}
	if col, _, _ := p.table.update(gtx); col >= 0 {
		p.sortBy(col)
	}
	select {
	case load := <-p.roundLoads:
		p.applyRoundLoad(load)
	default:
	}
	select {
	case update := <-p.updates:
		p.applyUpdate(update)
	default:
	}
}

// selectCategory forgets the team of the last category and reads the team of the selected category
// outside the window goroutine
func (p *skillAnalysisPage) selectCategory(invalidate func()) {
	p.round, p.selectedRunner, p.errorText = nil, -1, ""
	category := p.categories.selected
	if category == races.NoneDistance {
		return
	}
	go func() {
		defer invalidate()
		p.roundLoads <- p.loadRound(category)
	}()
}

// loadRound returns the round of a category in the latest team trial result
func (p *skillAnalysisPage) loadRound(category races.DistanceType) teamRoundLoad {
	rounds, err := racesim.ReadLatestTeamRounds(p.resultsDir)
	if err != nil {
		return teamRoundLoad{category: category, err: err}
	}
	for i := range rounds {
		if races.DistanceType(rounds[i].Category) == category {
			return teamRoundLoad{category: category, round: &rounds[i]}
		}
	}
	return teamRoundLoad{category: category, err: fmt.Errorf("the latest team trial has no %s round", category)}
}

// applyRoundLoad shows the team of a read round. A read for a category no longer selected is dropped.
func (p *skillAnalysisPage) applyRoundLoad(load teamRoundLoad) {
	if load.category != p.categories.selected {
		return
	}
	if load.err != nil {
		log.Printf("read team of %s: %v", load.category, load.err)
		p.errorText = load.err.Error()
		return
	}
	p.errorText = ""
	p.round = load.round
}

// start begins an analysis of the selected uma on the pool of the selected category
func (p *skillAnalysisPage) start(invalidate func()) {
	files := racesim.Locate()
	if len(files.Missing()) > 0 {
		p.errorText = missingEngineText
		return
	}
	runner := &p.round.Team[p.selectedRunner]
	config := racesim.SkillSurveyConfig{
		Round:      p.round,
		Runner:     runner,
		Courses:    p.poolCourses(p.categories.selected),
		Conditions: p.dataStore.RaceConditions(),
		Skills:     p.dataStore.GeneralSkills(),
		ScoreRules: p.dataStore.ScoreRules(),
	}
	p.analysisTitle = fmt.Sprintf("%s, %s", p.categories.selected, p.dataStore.CharaName(runner.CharaId))
	p.table.widths = nil
	p.table.setColumns(nil)
	p.table.scrollToTop()
	p.rows, p.racesShown, p.errorText = nil, 0, ""
	p.report = racesim.SkillSurveyReport{}
	p.running, p.startTime = true, time.Now()
	p.stop = &atomic.Bool{}
	go p.runAnalysis(&config, files, p.stop, invalidate)
}

// poolCourses returns the courses of the team trial pool of a category
func (p *skillAnalysisPage) poolCourses(category races.DistanceType) []db.TeamTrialRace {
	var courses []db.TeamTrialRace
	for _, course := range p.dataStore.TeamTrialPool() {
		if races.TeamTrialCategory(course.Distance, course.Ground) == category {
			courses = append(courses, course)
		}
	}
	return courses
}

// runAnalysis runs the skill survey outside the window goroutine and delivers its reports
func (p *skillAnalysisPage) runAnalysis(
	config *racesim.SkillSurveyConfig,
	files racesim.Files,
	stop *atomic.Bool,
	invalidate func(),
) {
	var (
		rows   []skillAnalysisRow
		latest racesim.SkillSurveyReport
	)
	engine, err := p.engines.open(files)
	if err == nil {
		config.Engine = engine
		err = racesim.RunSkillSurvey(config, stop, func(report racesim.SkillSurveyReport) {
			if report.Summaries != nil {
				rows = p.newRows(report.Summaries)
			}
			latest = report
			p.deliver(skillAnalysisUpdate{report: report, rows: rows}, invalidate)
		})
	}
	p.deliver(skillAnalysisUpdate{report: latest, rows: rows, finished: true, err: err}, invalidate)
}

// deliver passes an update to the window goroutine, replacing a pending update that was not
// applied, because the page does not read updates while another page shows. Each update holds the
// whole state, so nothing is lost.
func (p *skillAnalysisPage) deliver(update skillAnalysisUpdate, invalidate func()) {
	select {
	case <-p.updates:
	default:
	}
	p.updates <- update
	invalidate()
}

// newRows returns the table rows of the skill summaries. Upgrading a skill costs what the upgrade
// costs beyond the tier that the uma already owns.
func (p *skillAnalysisPage) newRows(summaries []racesim.SkillSummary) []skillAnalysisRow {
	rows := make([]skillAnalysisRow, len(summaries))
	for i := range summaries {
		summary := &summaries[i]
		ids := []int{summary.SkillId}
		points := p.dataStore.TotalSkillPoints(ids)
		if summary.ReplacedSkillId != 0 {
			points -= p.dataStore.TotalSkillPoints([]int{summary.ReplacedSkillId})
		}
		rows[i] = skillAnalysisRow{
			name:          p.dataStore.SkillNames(ids)[0],
			skillPoints:   points,
			scorePerPoint: summary.ScoreDelta / float64(points),
			summary:       *summary,
		}
	}
	return rows
}

// applyUpdate shows the state of the running analysis
func (p *skillAnalysisPage) applyUpdate(update skillAnalysisUpdate) {
	p.report = update.report
	if update.report.RacesPerCourse != p.racesShown {
		p.racesShown = update.report.RacesPerCourse
		p.baseline = update.report.Baseline
		p.rows = update.rows
		p.refreshCells()
	}
	if update.finished {
		p.running = false
		if update.err != nil {
			log.Printf("analyze skills: %v", update.err)
			p.errorText = update.err.Error()
		}
	}
}

// sortBy sorts the rows by a column. A second click on the sorted column reverses the order.
func (p *skillAnalysisPage) sortBy(col int) {
	ascending := skillAnalysisColumns[col].ascending
	if p.table.sortColumn == col {
		ascending = p.table.sortState == races.Descending
	}
	p.table.sortColumn, p.table.sortState = col, races.Descending
	if ascending {
		p.table.sortState = races.Ascending
	}
	p.refreshCells()
}

// refreshCells orders the rows by the sorted column and replaces the text of the table
func (p *skillAnalysisPage) refreshCells() {
	key := skillAnalysisColumns[p.table.sortColumn].key
	sign := 1
	if p.table.sortState == races.Descending {
		sign = -1
	}
	slices.SortStableFunc(p.rows, func(a, b skillAnalysisRow) int {
		order := 0
		if key != nil {
			order = sign * cmp.Compare(key(&a), key(&b))
		} else {
			order = sign * cmp.Compare(a.name, b.name)
		}
		return cmp.Or(order, cmp.Compare(a.name, b.name))
	})
	columns := make([][]string, len(skillAnalysisColumns))
	for col := range skillAnalysisColumns {
		cells := make([]string, len(p.rows))
		for i := range p.rows {
			cells[i] = skillAnalysisColumns[col].cell(&p.rows[i])
		}
		columns[col] = cells
	}
	p.table.setColumns(columns)
}

// detailText states what the table shows and the baseline of the uma
func (p *skillAnalysisPage) detailText() string {
	if p.racesShown == 0 {
		return ""
	}
	return fmt.Sprintf("%s: %d of %d races per course, baseline score %.0f, win rate %.1f%%",
		p.analysisTitle, p.racesShown, racesim.SkillSurveyBatches*racesim.SkillSurveyBatchRaces,
		p.baseline.Score, p.baseline.WinRate*100)
}

// statusText states the progress of the running analysis, or the last failure
func (p *skillAnalysisPage) statusText() string {
	if p.errorText != "" || !p.running {
		return p.errorText
	}
	r := &p.report
	if r.TotalCourseRuns == 0 {
		return findingSkillsText
	}
	status := fmt.Sprintf("%d of %d skills can activate", r.PossibleSkills, r.CandidateSkills)
	if r.CourseRuns > 0 {
		left := time.Duration(float64(time.Since(p.startTime)) * float64(r.TotalCourseRuns-r.CourseRuns) / float64(r.CourseRuns))
		status += fmt.Sprintf(", %.0f min left", max(left.Minutes(), 1))
	}
	return status
}

// layout draws the toolbar above the table, or an instruction
func (p *skillAnalysisPage) layout(gtx layout.Context, theme *material.Theme, invalidate func()) layout.Dimensions {
	p.update(gtx, invalidate)
	return p.panel.layout(gtx, theme, true, p.detailText(), func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutToolbar(gtx, theme)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if len(p.rows) == 0 {
					return layoutCenteredText(gtx, theme, p.instruction(), colorTextMuted)
				}
				return p.layoutResult(gtx, theme)
			}),
		)
	})
}

// instruction returns the text that takes the place of the table
func (p *skillAnalysisPage) instruction() string {
	switch {
	case p.running:
		return fmt.Sprintf(firstResultText, racesim.SkillSurveyBatchRaces)
	case p.round == nil:
		return selectCategoryText
	case p.selectedRunner < 0:
		return selectUmaText
	}
	return readyToStartText
}

// layoutToolbar draws the category pills, the team pills, the start button and the status
func (p *skillAnalysisPage) layoutToolbar(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layoutPageToolbar(gtx, func(gtx layout.Context) layout.Dimensions {
		gap := layout.Rigid(layout.Spacer{Width: skillAnalysisGroupGap}.Layout)
		children := make([]layout.FlexChild, 0, teamRunnerCount+5)
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if p.running {
				gtx = gtx.Disabled()
			}
			return p.categories.layout(gtx, theme)
		}), gap)
		if p.round != nil {
			for i := range p.round.Team {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.running {
						gtx = gtx.Disabled()
					}
					var highlight highlightColors
					if i == p.selectedRunner {
						highlight = selectionHighlight
					}
					name := p.dataStore.CharaName(p.round.Team[i].CharaId)
					dims := layoutPill(gtx, theme, &p.runnerButtons[i], name, highlight)
					dims.Size.X += gtx.Dp(filterGap)
					return dims
				}))
			}
			children = append(children, gap)
		}
		label := startAnalysisText
		if p.running {
			label = stopAnalysisText
		}
		ready := p.running || (p.round != nil && p.selectedRunner >= 0)
		statusColor := colorTextMuted
		if p.errorText != "" {
			statusColor = colorError
		}
		children = append(children,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutActionButton(gtx, theme, &p.startButton, label, !ready)
			}),
			gap,
			layout.Flexed(1, textLabel(theme, smallTextSize, p.statusText(), statusColor).Layout),
		)
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
	})
}

// layoutResult draws the caption above the table
func (p *skillAnalysisPage) layoutResult(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Top: skillAnalysisCaptionY, Bottom: skillAnalysisCaptionY, Left: tableCellPadding}
			return inset.Layout(gtx, textLabel(theme, smallTextSize, skillAnalysisCaption, colorTextMuted).Layout)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return p.table.layout(gtx, theme)
		}),
	)
}
