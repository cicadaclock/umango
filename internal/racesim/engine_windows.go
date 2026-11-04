// Package racesim runs the race engine library and reads saved races for it.
package racesim

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"syscall"
	"unsafe"
)

const (
	// Interface version of the engine library that this package calls
	interfaceVersion = 5
	// Longest failure text that this package reads from the library
	errorTextCapacity = 1024
)

// Engine is the race engine library with its game data loaded. Different prepared races may run
// concurrently, but one prepared race is not safe for concurrent use. A failure reported during
// concurrent runs may carry the text of the other run's failure.
type Engine struct {
	// Game data that the library holds
	data uintptr

	prepare, release, run, releaseSkillSpans, possibleSkills, runSkillTrial, lastError *syscall.LazyProc
}

// skillTrialRequest is a SkillTrial in the layout of the library's SkillTrial
type skillTrialRequest struct {
	runner                      uintptr
	courseSetId, raceInstanceId int32
	skillIds, replacedSkillIds  *int32
	skillCount                  uintptr
	variations                  *RaceVariation
	variationCount              uintptr
	scoreRules                  *ScoreRules
}

// request returns the trial in the layout of the library
func (t *SkillTrial) request() skillTrialRequest {
	return skillTrialRequest{
		runner:           uintptr(t.Runner),
		courseSetId:      t.CourseSetId,
		raceInstanceId:   t.RaceInstanceId,
		skillIds:         unsafe.SliceData(t.SkillIds),
		replacedSkillIds: unsafe.SliceData(t.ReplacedSkillIds),
		skillCount:       uintptr(len(t.SkillIds)),
		variations:       unsafe.SliceData(t.Variations),
		variationCount:   uintptr(len(t.Variations)),
		scoreRules:       t.Rules,
	}
}

// trialOutput is what the library stores for one run, in the layout of its TrialOutput
type trialOutput struct {
	outcomes       *RunnerOutcome
	outcomeCount   uintptr
	skillSpans     *SkillSpan
	skillSpanCount uintptr
}

// PreparedRace identifies a field and course that the library holds for repeated runs
type PreparedRace uintptr

// Open loads the engine library and its game data. Both last for the life of the process.
func Open(files Files) (*Engine, error) {
	library := syscall.NewLazyDLL(files.LibraryPath())
	if err := library.Load(); err != nil {
		return nil, fmt.Errorf("load race engine library: %w", err)
	}
	version := library.NewProc("racesim_interface_version")
	open := library.NewProc("racesim_open")
	e := &Engine{
		prepare:           library.NewProc("racesim_prepare_saved_race"),
		release:           library.NewProc("racesim_release"),
		run:               library.NewProc("racesim_run"),
		releaseSkillSpans: library.NewProc("racesim_release_skill_spans"),
		possibleSkills:    library.NewProc("racesim_possible_skills"),
		runSkillTrial:     library.NewProc("racesim_run_skill_trial"),
		lastError:         library.NewProc("racesim_last_error"),
	}
	procs := [...]*syscall.LazyProc{
		version, open, e.prepare, e.release, e.run, e.releaseSkillSpans, e.possibleSkills, e.runSkillTrial, e.lastError,
	}
	for _, proc := range procs {
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("race engine library has no function %s: %w", proc.Name, err)
		}
	}
	if found, _, _ := version.Call(); found != interfaceVersion {
		return nil, fmt.Errorf("race engine library has interface version %d, expected %d", found, interfaceVersion)
	}
	directory := files.DataPath()
	e.data, _, _ = open.Call(uintptr(unsafe.Pointer(unsafe.StringData(directory))), uintptr(len(directory)))
	if e.data == 0 {
		return nil, fmt.Errorf("load race engine data: %w", e.lastFailure())
	}
	return e, nil
}

// lastFailure returns the failure that the library holds
func (e *Engine) lastFailure() error {
	var text [errorTextCapacity]byte
	length, _, _ := e.lastError.Call(uintptr(unsafe.Pointer(&text[0])), errorTextCapacity)
	return errors.New(strings.TrimSpace(string(text[:length])))
}

// Prepare has the library hold the field, course and conditions of a saved race file for repeated
// runs. Release frees the race.
func (e *Engine) Prepare(savedRace []byte) (PreparedRace, error) {
	if len(savedRace) == 0 {
		return 0, errors.New("prepare race: the race file is empty")
	}
	race, _, _ := e.prepare.Call(uintptr(unsafe.Pointer(&savedRace[0])), uintptr(len(savedRace)))
	if race == 0 {
		return 0, fmt.Errorf("prepare race: %w", e.lastFailure())
	}
	return PreparedRace(race), nil
}

// Release frees a prepared race
func (e *Engine) Release(race PreparedRace) {
	e.release.Call(uintptr(race))
}

// Run simulates one race per seed on all processor threads, stores the outcome of each runner and
// returns every skill span in seed order. The outcomes of seed k occupy the elements from
// k * runners in gate order, so outcomes holds len(seeds) * runners elements. The library
// overwrites the elements on each run.
func (e *Engine) Run(race PreparedRace, seeds []int32, outcomes []RunnerOutcome) ([]SkillSpan, error) {
	if len(seeds) == 0 || len(outcomes) == 0 {
		return nil, errors.New("run race: the run has no seeds or the race has no runners")
	}
	output := trialOutput{outcomes: &outcomes[0], outcomeCount: uintptr(len(outcomes))}
	status, _, _ := e.run.Call(
		e.data,
		uintptr(race),
		uintptr(unsafe.Pointer(&seeds[0])),
		uintptr(len(seeds)),
		uintptr(unsafe.Pointer(&output)),
	)
	if status != 0 {
		return nil, fmt.Errorf("run race: %w", e.lastFailure())
	}
	defer e.releaseSkillSpans.Call(uintptr(unsafe.Pointer(output.skillSpans)), output.skillSpanCount)
	return slices.Clone(unsafe.Slice(output.skillSpans, output.skillSpanCount)), nil
}

// PossibleSkills sets each element of possible, which holds one element per skill of the trial, to
// 1 when that skill can activate for the runner on the trial's course under any of the trial's
// variations, otherwise 0. Skills count as unable to activate only when a fixed fact of the race
// rules them out, such as the course, the ground or the running style.
func (e *Engine) PossibleSkills(race PreparedRace, trial *SkillTrial, possible []uint8) error {
	request := trial.request()
	status, _, _ := e.possibleSkills.Call(
		e.data,
		uintptr(race),
		uintptr(unsafe.Pointer(&request)),
		uintptr(unsafe.Pointer(unsafe.SliceData(possible))),
	)
	if status != 0 {
		return fmt.Errorf("find possible skills: %w", e.lastFailure())
	}
	return nil
}

// RunSkillTrial simulates the trial on all processor threads but one. It stores the runner's
// baseline for each variation, and for each skill and variation what the skill changed against the
// same race with the skill muted. The outcomes of skill k occupy the elements from
// k * len(trial.Variations) in variation order.
func (e *Engine) RunSkillTrial(
	race PreparedRace,
	trial *SkillTrial,
	baselines []BaselineOutcome,
	outcomes []SkillPairOutcome,
) error {
	request := trial.request()
	status, _, _ := e.runSkillTrial.Call(
		e.data,
		uintptr(race),
		uintptr(unsafe.Pointer(&request)),
		uintptr(unsafe.Pointer(unsafe.SliceData(baselines))),
		uintptr(unsafe.Pointer(unsafe.SliceData(outcomes))),
	)
	if status != 0 {
		return fmt.Errorf("run skill trial: %w", e.lastFailure())
	}
	return nil
}
