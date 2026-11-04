package ui

import (
	"errors"
	"fmt"
	"image"
	"io"
	"io/fs"
	"log"
	"math/rand/v2"
	"slices"
	"sync"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	"github.com/cicadaclock/umango/internal/data"
	"github.com/cicadaclock/umango/internal/racesim"
)

const (
	simulatorPanelTitle  = "Simulator"
	importRaceButtonText = "Import race file…"
	newTrialButtonText   = "New Trial"
	checkFilesButtonText = "Check Again"
	emptySimulatorText   = "Import a Champions Meeting race file to simulate the race"
	missingFilesText     = "The simulator requires these missing files:"
	missingFileGap       = unit.Dp(6)
	missingFilePadding   = unit.Dp(16)
	simulatorButtonGap   = unit.Dp(8)
	// Races that one trial simulates, whose averages the placement table shows
	racesPerTrial = 100
)

// simulatorDetail states what the values of the placement table are
var simulatorDetail = fmt.Sprintf("Average of %d simulated races", racesPerTrial)

// engineHolder opens the race engine at its first use, so that the engine costs nothing at
// application startup, and shares it between the pages that simulate races
type engineHolder struct {
	mutex  sync.Mutex
	engine *racesim.Engine
}

// open returns the race engine, opening it when no earlier call succeeded
func (h *engineHolder) open(files racesim.Files) (*racesim.Engine, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if h.engine != nil {
		return h.engine, nil
	}
	engine, err := racesim.Open(files)
	if err != nil {
		return nil, err
	}
	h.engine = engine
	return engine, nil
}

// simulation carries the result of a background simulation to the window goroutine
type simulation struct {
	// Cell text of the placement table, one slice per column
	columns [][]string
	// Gate index of the runner on each row of the placement table
	rowRunners []int
	// Skill activity of each runner in gate order
	skillViews []skillActivityView
	// View of the course of a newly imported race, nil after a trial of the same race
	course *courseView
	err    error
}

// simulatorPage simulates the field of a saved Champions Meeting race in trials of racesPerTrial
// races: the first trial from the race's seed, then from a random seed at each new trial
type simulatorPage struct {
	dataStore                              *data.DataStore
	engines                                *engineHolder
	panel                                  panel
	importButton, trialButton, checkButton widget.Clickable
	table                                  tableView
	files                                  racesim.Files
	missing                                []string
	errorText                              string
	course                                 *courseView
	rowRunners                             []int
	skillViews                             []skillActivityView
	skillList                              widget.List
	skillKey                               paint.ImageOp
	// Whether every required engine file exists
	filesReady bool
	// Whether a race is prepared and whether the table holds a result
	hasRace, hasResult bool
	// Whether a background job is running. The job owns the fields below it until it delivers
	// its result.
	busy    bool
	results chan simulation

	engine *racesim.Engine
	// Layout of every course, read at the first import that finds the file
	courseLayouts map[int]racesim.CourseLayout
	race          racesim.PreparedRace
	runners       []racesim.Runner
	// Course distance of the prepared race in meters
	courseDistance float64
	// Skill points of each runner in gate order
	skillPoints []int
	// Outcomes of the last trial
	outcomes []racesim.RunnerOutcome
}

// newSimulatorPage returns a simulator page that has checked for the engine files
func newSimulatorPage(dataStore *data.DataStore, engines *engineHolder) *simulatorPage {
	p := &simulatorPage{
		dataStore: dataStore,
		engines:   engines,
		panel:     panel{title: simulatorPanelTitle},
		results:   make(chan simulation, 1),
		table:     newPlacementTable(),
		skillList: widget.List{List: layout.List{Axis: layout.Vertical}},
		skillKey:  newSkillKeyImage(),
	}
	p.checkFiles()
	return p
}

// checkFiles finds the engine files and lists those that are missing
func (p *simulatorPage) checkFiles() {
	p.files = racesim.Locate()
	p.missing = p.files.Missing()
	p.filesReady = len(p.missing) == 0
}

// update starts the job of a clicked button and applies a finished job
func (p *simulatorPage) update(gtx layout.Context, picker *explorer.Explorer, invalidate func()) {
	if p.checkButton.Clicked(gtx) && !p.busy {
		p.checkFiles()
	}
	if p.importButton.Clicked(gtx) && !p.busy {
		p.busy = true
		go p.chooseRace(picker, invalidate)
	}
	if p.trialButton.Clicked(gtx) && !p.busy && p.hasRace {
		p.busy = true
		go p.runTrial(invalidate)
	}
	if _, row, _ := p.table.update(gtx); row >= 0 {
		p.selectRow(row)
	}
	select {
	case result := <-p.results:
		p.busy = false
		switch {
		case errors.Is(result.err, explorer.ErrUserDecline):
		case result.err != nil:
			log.Printf("simulate race: %v", result.err)
			p.errorText = result.err.Error()
		default:
			p.errorText = ""
			p.hasRace, p.hasResult = true, true
			p.applyResult(&result)
		}
	default:
	}
}

// applyResult shows the result of a trial. The selected runner stays selected through a new trial
// of the same race, and importing a race clears the selection.
func (p *simulatorPage) applyResult(result *simulation) {
	selected := -1
	if result.course != nil {
		p.course = result.course
	} else if p.table.selectedRow >= 0 {
		selected = p.rowRunners[p.table.selectedRow]
	}
	p.rowRunners, p.skillViews = result.rowRunners, result.skillViews
	p.table.selectedRow = slices.Index(p.rowRunners, selected)
	p.table.setColumns(result.columns)
	p.table.widths = nil
	p.table.scrollToTop()
}

// selectRow selects the runner of a row, or clears the selection when the row is already selected
func (p *simulatorPage) selectRow(row int) {
	if row == p.table.selectedRow {
		p.table.selectedRow = -1
		return
	}
	p.table.selectedRow = row
	p.skillList.Position = layout.Position{}
}

// chooseRace waits for the file picker, then imports and simulates the chosen race file outside
// the window goroutine
func (p *simulatorPage) chooseRace(picker *explorer.Explorer, invalidate func()) {
	defer invalidate()
	file, err := picker.ChooseFile(".json")
	if err != nil {
		p.results <- simulation{err: err}
		return
	}
	defer file.Close()
	p.results <- p.importRace(file)
}

// runTrial simulates the prepared race from a random seed outside the window goroutine
func (p *simulatorPage) runTrial(invalidate func()) {
	defer invalidate()
	p.results <- p.simulate(rand.Int32())
}

// importRace prepares the field and course of a race file and simulates a trial from the race's seed
func (p *simulatorPage) importRace(source io.Reader) simulation {
	savedRace, err := io.ReadAll(source)
	if err != nil {
		return simulation{err: fmt.Errorf("read race file: %w", err)}
	}
	race, err := racesim.DecodeChampionsRace(savedRace)
	if err != nil {
		return simulation{err: err}
	}
	p.engine, err = p.engines.open(p.files)
	if err != nil {
		return simulation{err: err}
	}
	prepared, err := p.engine.Prepare(savedRace)
	if err != nil {
		return simulation{err: err}
	}
	if p.race != 0 {
		p.engine.Release(p.race)
	}
	p.race, p.runners = prepared, race.Runners
	p.courseDistance = float64(race.Course.Distance)
	p.skillPoints = make([]int, len(race.Runners))
	for i := range race.Runners {
		p.skillPoints[i] = p.dataStore.TotalSkillPoints(race.Runners[i].SkillIds)
	}
	p.outcomes = make([]racesim.RunnerOutcome, racesPerTrial*len(race.Runners))
	result := p.simulate(race.Seed)
	result.course = p.viewOfCourse(&race)
	return result
}

// viewOfCourse returns the view of the course of a race. A course layout that cannot be read does
// not stop the simulation, so the view then carries a note in place of the course.
func (p *simulatorPage) viewOfCourse(race *racesim.Race) *courseView {
	title := courseTitle(p.dataStore.RaceTrackName(race.Course.RaceTrackId), &race.Course)
	if p.courseLayouts == nil {
		path := p.files.CourseLayoutPath()
		layouts, err := racesim.LoadCourseLayouts(path)
		if errors.Is(err, fs.ErrNotExist) {
			return newCourseNote(title, fmt.Sprintf(courseLayoutMissingFormat, path), race.Course.Distance)
		}
		if err != nil {
			log.Printf("load course layouts: %v", err)
			return newCourseNote(title, err.Error(), race.Course.Distance)
		}
		p.courseLayouts = layouts
	}
	layout, ok := p.courseLayouts[race.Course.Id]
	if !ok {
		return newCourseNote(title, fmt.Sprintf(courseLayoutAbsentFormat, race.Course.Id), race.Course.Distance)
	}
	return newCourseView(title, race.Course.Distance, &layout)
}

// simulate runs a trial of the prepared race and returns the placement table
func (p *simulatorPage) simulate(firstSeed int32) simulation {
	seeds := trialSeeds(firstSeed)
	spans, err := p.engine.Run(p.race, seeds[:], p.outcomes)
	if err != nil {
		return simulation{err: err}
	}
	summaries := racesim.Summarize(p.outcomes, len(p.runners))
	activities := racesim.SummarizeSkills(spans, len(p.runners), p.courseDistance)
	rows := make([]placementRow, len(p.runners))
	skillViews := make([]skillActivityView, len(p.runners))
	for i := range rows {
		rows[i] = placementRow{
			gate:        i + 1,
			runner:      &p.runners[i],
			skillPoints: p.skillPoints[i],
			summary:     &summaries[i],
		}
		skillViews[i] = newSkillActivityView(p.dataStore, &p.runners[i], activities[i])
	}
	columns := placementCells(rows)
	rowRunners := make([]int, len(rows))
	for i := range rows {
		rowRunners[i] = rows[i].gate - 1
	}
	return simulation{columns: columns, rowRunners: rowRunners, skillViews: skillViews}
}

// trialSeeds returns the seeds of a trial. The first seed determines the others, so a race file
// gives the same averages each time that it is imported. The first values of the engine's generator
// follow a regular pattern across consecutive seeds, so the others do not count up from the first.
func trialSeeds(first int32) [racesPerTrial]int32 {
	generator := rand.New(rand.NewPCG(uint64(first), 0))
	seeds := [racesPerTrial]int32{first}
	for i := 1; i < racesPerTrial; i++ {
		seeds[i] = generator.Int32()
	}
	return seeds
}

// layout draws the toolbar above the missing files, the simulation result, or an instruction
func (p *simulatorPage) layout(
	gtx layout.Context,
	theme *material.Theme,
	picker *explorer.Explorer,
	invalidate func(),
) layout.Dimensions {
	p.update(gtx, picker, invalidate)
	detail := ""
	if p.hasResult {
		detail = simulatorDetail
	}
	return p.panel.layout(gtx, theme, true, detail, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutToolbar(gtx, theme)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				switch {
				case len(p.missing) > 0:
					return p.layoutMissingFiles(gtx, theme)
				case p.hasResult:
					return p.layoutResult(gtx, theme)
				}
				return layoutCenteredText(gtx, theme, emptySimulatorText, colorTextMuted)
			}),
		)
	})
}

// layoutResult draws the placements above the course and the skill activity of the selected
// runner. The table takes the height of its rows and gives up height only when the window is too
// short for the course and the skill rows.
func (p *simulatorPage) layoutResult(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	courseDims := p.course.layout(gtx, theme)
	course := macro.Stop()
	selected := p.table.selectedRow >= 0
	reserved := courseDims.Size.Y
	if selected {
		reserved += gtx.Dp(skillListMinimumHeight)
	}
	tableHeight := min(p.table.contentHeight(gtx, theme), max(gtx.Constraints.Max.Y-reserved, 0))
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.Y = tableHeight
			return p.table.layout(gtx, theme)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			course.Add(gtx.Ops)
			return courseDims
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if !selected {
				return layoutCenteredText(gtx, theme, selectRunnerText, colorTextMuted)
			}
			view := &p.skillViews[p.rowRunners[p.table.selectedRow]]
			return view.layout(gtx, theme, &p.skillList, p.course, p.skillKey)
		}),
	)
}

// layoutToolbar draws the buttons that apply to the state of the engine files and the last error
func (p *simulatorPage) layoutToolbar(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	return layoutPageToolbar(gtx, func(gtx layout.Context) layout.Dimensions {
		button := func(clickable *widget.Clickable, label string) layout.FlexChild {
			return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				dims := layoutActionButton(gtx, theme, clickable, label, p.busy)
				dims.Size.X += gtx.Dp(simulatorButtonGap)
				return dims
			})
		}
		errorLabel := layout.Flexed(1, textLabel(theme, smallTextSize, p.errorText, colorError).Layout)
		flex := layout.Flex{Alignment: layout.Middle}
		switch {
		case !p.filesReady:
			return flex.Layout(gtx, button(&p.checkButton, checkFilesButtonText), errorLabel)
		case p.hasRace:
			return flex.Layout(gtx,
				button(&p.importButton, importRaceButtonText),
				button(&p.trialButton, newTrialButtonText),
				errorLabel,
			)
		}
		return flex.Layout(gtx, button(&p.importButton, importRaceButtonText), errorLabel)
	})
}

// layoutMissingFiles lists the required files that are missing
func (p *simulatorPage) layoutMissingFiles(gtx layout.Context, theme *material.Theme) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	children := make([]layout.FlexChild, 0, 2*len(p.missing)+1)
	heading := material.Label(theme, textSize, missingFilesText)
	heading.Color = colorError
	children = append(children, layout.Rigid(heading.Layout))
	for _, file := range p.missing {
		label := material.Label(theme, textSize, file)
		label.Color = colorText
		children = append(children,
			layout.Rigid(layout.Spacer{Height: missingFileGap}.Layout),
			layout.Rigid(label.Layout),
		)
	}
	inset := layout.UniformInset(missingFilePadding)
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
}
