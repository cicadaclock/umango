package racesim

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	directoryName        = "racesim"
	libraryFileName      = "racesim.dll"
	dataDirectoryName    = "data"
	courseLayoutFileName = "courses.json"
)

// Files and folders of game data that the engine library reads
var requiredDataEntries = [...]string{"master_sim.db", "race_paramdefine.json", "courseeventparam", "coursepath"}

// Files holds the location of the race engine files that are not part of the executable
type Files struct {
	// Folder of the engine files, which need not exist
	Directory string
}

// Locate finds the racesim folder beside the executable or, failing that, in the working directory
func Locate() Files {
	return Files{Directory: findDirectory()}
}

// findDirectory returns the first racesim folder that exists, or the folder beside the executable
// when none exists
func findDirectory() string {
	var candidates []string
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), directoryName))
	}
	if working, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(working, directoryName))
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return directoryName
}

// LibraryPath returns the path of the engine library
func (f Files) LibraryPath() string {
	return filepath.Join(f.Directory, libraryFileName)
}

// DataPath returns the path of the folder of engine game data
func (f Files) DataPath() string {
	return filepath.Join(f.Directory, dataDirectoryName)
}

// CourseLayoutPath returns the path of the course layout file, which only the course view needs
func (f Files) CourseLayoutPath() string {
	return filepath.Join(f.Directory, courseLayoutFileName)
}

// Missing describes each required file that the folder lacks
func (f Files) Missing() []string {
	var missing []string
	if _, err := os.Stat(f.LibraryPath()); err != nil {
		missing = append(missing, fmt.Sprintf("Race engine library: %s", f.LibraryPath()))
	}
	for _, entry := range requiredDataEntries {
		path := filepath.Join(f.DataPath(), entry)
		if _, err := os.Stat(path); err != nil {
			missing = append(missing, fmt.Sprintf("Race engine data: %s", path))
		}
	}
	return missing
}
