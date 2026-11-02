package ui

import (
	"errors"
	"fmt"
	"image"
	"io"
	"io/fs"
	"log"
	"math/rand/v2"
	"strconv"

	"gioui.org/layout"
	"gioui.org/text"
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
	placeHeader          = "Place"
	placementUmaHeader   = "Uma"
	placementRowHeight   = unit.Dp(32)
	missingFileGap       = unit.Dp(6)
	missingFilePadding   = unit.Dp(16)
	simulatorButtonGap   = unit.Dp(8)
)

// simulation carries the result of a background simulation to the window goroutine
type simulation struct {
	// Cell text of the placement table, one slice per column
	columns [][]string
	// View of the course of a newly imported race, nil after a trial of the same race
	course *courseView
	err    error
}

// simulatorPage simulates the field of a saved Champions Meeting race from the race's seed, then
// from a random seed at each new trial
type simulatorPage struct {
	dataStore                              *data.DataStore
	panel                                  panel
	importButton, trialButton, checkButton widget.Clickable
	table                                  tableView
	files                                  racesim.Files
	missing                                []string
	errorText                              string
	course                                 *courseView
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
	names         []string
	finishOrder   []int32
}

// newSimulatorPage returns a simulator page that has checked for the engine files
func newSimulatorPage(dataStore *data.DataStore) *simulatorPage {
	p := &simulatorPage{
		dataStore: dataStore,
		panel:     panel{title: simulatorPanelTitle},
		results:   make(chan simulation, 1),
		table: newTableView(
			[]string{placeHeader, placementUmaHeader},
			[]tableColumn{{sample: "18", alignment: text.End}, {fill: true}},
			placementRowHeight,
			false,
		),
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
			p.table.setColumns(result.columns)
			p.table.scrollToTop()
			if result.course != nil {
				p.course = result.course
			}
		}
	default:
	}
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

// importRace prepares the field and course of a race file and simulates it from the race's seed.
// The engine starts at the first import, so that it costs nothing at application startup.
func (p *simulatorPage) importRace(source io.Reader) simulation {
	savedRace, err := io.ReadAll(source)
	if err != nil {
		return simulation{err: fmt.Errorf("read race file: %w", err)}
	}
	race, err := racesim.DecodeChampionsRace(savedRace)
	if err != nil {
		return simulation{err: err}
	}
	if p.engine == nil {
		p.engine, err = racesim.Open(p.files)
		if err != nil {
			return simulation{err: err}
		}
	}
	prepared, err := p.engine.Prepare(savedRace)
	if err != nil {
		return simulation{err: err}
	}
	if p.race != 0 {
		p.engine.Release(p.race)
	}
	p.race, p.names = prepared, race.Names
	p.finishOrder = make([]int32, len(race.Names))
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
			return newCourseNote(title, fmt.Sprintf(courseLayoutMissingFormat, path))
		}
		if err != nil {
			log.Printf("load course layouts: %v", err)
			return newCourseNote(title, err.Error())
		}
		p.courseLayouts = layouts
	}
	layout, ok := p.courseLayouts[race.Course.Id]
	if !ok {
		return newCourseNote(title, fmt.Sprintf(courseLayoutAbsentFormat, race.Course.Id))
	}
	return newCourseView(title, race.Course.Distance, &layout)
}

// simulate runs the prepared race from a seed and returns the placement table
func (p *simulatorPage) simulate(seed int32) simulation {
	if err := p.engine.Run(p.race, seed, p.finishOrder); err != nil {
		return simulation{err: err}
	}
	places := make([]string, len(p.finishOrder))
	names := make([]string, len(p.finishOrder))
	for runner, order := range p.finishOrder {
		places[order] = strconv.Itoa(int(order) + 1)
		names[order] = p.names[runner]
	}
	return simulation{columns: [][]string{places, names}}
}

// layout draws the toolbar above the missing files, the course and placements, or an instruction
func (p *simulatorPage) layout(
	gtx layout.Context,
	theme *material.Theme,
	picker *explorer.Explorer,
	invalidate func(),
) layout.Dimensions {
	p.update(gtx, picker, invalidate)
	return p.panel.layout(gtx, theme, true, "", func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutToolbar(gtx, theme)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				switch {
				case len(p.missing) > 0:
					return p.layoutMissingFiles(gtx, theme)
				case p.hasResult:
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return p.course.layout(gtx, theme)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return p.table.layout(gtx, theme)
						}),
					)
				}
				return layoutCenteredText(gtx, theme, emptySimulatorText, colorTextMuted)
			}),
		)
	})
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
