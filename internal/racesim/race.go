package racesim

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

const championsRaceType = "Champions"

// CourseSet holds the race_course_set values of a course, numbered as in master.mdb
type CourseSet struct {
	Id          int `json:"id"`
	RaceTrackId int `json:"raceTrackId"`
	Distance    int `json:"distance"`
	Ground      int `json:"ground"`
	Inout       int `json:"inout"`
	Turn        int `json:"turn"`
}

// Race holds what the application shows of a saved race. The library reads the field and conditions
// from the file itself, and the recorded result is not used.
type Race struct {
	Course CourseSet
	// Seed of the recorded race
	Seed int32
	// Display name of each runner in gate order, with the trainer added where the field repeats
	// a chara
	Names []string
}

// championsFile holds the fields of a saved Champions Meeting race that this package reads
type championsFile struct {
	RaceType   string            `json:"raceType"`
	RandomSeed int64             `json:"randomSeed"`
	CourseSet  CourseSet         `json:"raceCourseSet"`
	Runners    []championsRunner `json:"raceHorse"`
}

// championsRunner is one runner of a saved Champions Meeting race
type championsRunner struct {
	HorseIndex  int    `json:"horseIndex"`
	CharaName   string `json:"charaName"`
	TrainerName string `json:"trainerName"`
}

// DecodeChampionsRace reads the course, seed and runner names of a saved Champions Meeting race
func DecodeChampionsRace(savedRace []byte) (Race, error) {
	var file championsFile
	if err := json.Unmarshal(savedRace, &file); err != nil {
		return Race{}, fmt.Errorf("decode race file: %w", err)
	}
	if file.RaceType != championsRaceType {
		return Race{}, errors.New("the file is not a saved Champions Meeting race")
	}
	if len(file.Runners) == 0 {
		return Race{}, errors.New("the race file has no runners")
	}
	slices.SortStableFunc(file.Runners, func(a, b championsRunner) int {
		return a.HorseIndex - b.HorseIndex
	})
	return Race{
		Course: file.CourseSet,
		Seed:   int32(file.RandomSeed),
		Names:  runnerNames(file.Runners),
	}, nil
}

// runnerNames returns the display name of each runner
func runnerNames(runners []championsRunner) []string {
	counts := make(map[string]int, len(runners))
	for i := range runners {
		counts[runners[i].CharaName]++
	}
	names := make([]string, len(runners))
	for i := range runners {
		names[i] = runners[i].CharaName
		if counts[names[i]] > 1 && runners[i].TrainerName != "" {
			names[i] += " (" + runners[i].TrainerName + ")"
		}
	}
	return names
}
