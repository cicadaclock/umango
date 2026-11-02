//go:build !windows

// Package racesim runs the race engine library and reads saved races for it.
package racesim

import "errors"

// errUnsupported reports that this platform has no way to call the engine library
var errUnsupported = errors.New("the race engine library requires Windows")

// Engine is the race engine library, which this platform cannot load
type Engine struct{}

// PreparedRace identifies a field and course that the library holds for repeated runs
type PreparedRace uintptr

// Open reports that this platform cannot load the engine library
func Open(Files) (*Engine, error) {
	return nil, errUnsupported
}

// Prepare reports that this platform cannot load the engine library
func (*Engine) Prepare([]byte) (PreparedRace, error) {
	return 0, errUnsupported
}

// Release has no race to free on this platform
func (*Engine) Release(PreparedRace) {}

// Run reports that this platform cannot load the engine library
func (*Engine) Run(PreparedRace, int32, []int32) error {
	return errUnsupported
}
