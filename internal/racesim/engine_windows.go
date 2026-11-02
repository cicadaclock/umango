// Package racesim runs the race engine library and reads saved races for it.
package racesim

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	// Interface version of the engine library that this package calls
	interfaceVersion = 1
	// Longest failure text that this package reads from the library
	errorTextCapacity = 1024
)

// Engine is the race engine library with its game data loaded. It is not safe for concurrent use.
type Engine struct {
	// Game data that the library holds
	data uintptr

	prepare, release, run, lastError *syscall.LazyProc
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
		prepare:   library.NewProc("racesim_prepare_saved_race"),
		release:   library.NewProc("racesim_release"),
		run:       library.NewProc("racesim_run"),
		lastError: library.NewProc("racesim_last_error"),
	}
	for _, proc := range [...]*syscall.LazyProc{version, open, e.prepare, e.release, e.run, e.lastError} {
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

// Run simulates one race from a seed and stores the finish order of each runner, counted from 0,
// in finishOrder, which holds one element per runner in gate order
func (e *Engine) Run(race PreparedRace, seed int32, finishOrder []int32) error {
	if len(finishOrder) == 0 {
		return errors.New("run race: the race has no runners")
	}
	status, _, _ := e.run.Call(
		e.data,
		uintptr(race),
		uintptr(uint32(seed)),
		uintptr(unsafe.Pointer(&finishOrder[0])),
		uintptr(len(finishOrder)),
	)
	if status != 0 {
		return fmt.Errorf("run race: %w", e.lastFailure())
	}
	return nil
}
