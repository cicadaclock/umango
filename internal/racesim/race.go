package racesim

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

const (
	championsRaceType = "Champions"
	// runningStyleEx of a Front Runner that runs away
	runawayStyleEx = "Oonige"
)

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
	// Runners in gate order
	Runners []Runner
}

// Runner holds what the application shows of one runner, which is the same in every simulated race
type Runner struct {
	Name, Trainer string
	// Running style, numbered as in races.RunStyle
	RunningStyle int
	Runaway      bool
	// Mood from 1 (Awful) to 5 (Great)
	Mood int
	// Raw stats as trained
	Speed, Stamina, Power, Guts, Wit int
	// Skills that the runner learned
	SkillIds []int
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
	HorseIndex     int            `json:"horseIndex"`
	CharaName      string         `json:"charaName"`
	TrainerName    string         `json:"trainerName"`
	RunningStyleEx string         `json:"runningStyleEx"`
	Entry          championsEntry `json:"responseHorseData"`
}

// championsEntry is the entry data of one runner of a saved Champions Meeting race
type championsEntry struct {
	Speed        int `json:"speed"`
	Stamina      int `json:"stamina"`
	Power        int `json:"pow"`
	Guts         int `json:"guts"`
	Wit          int `json:"wiz"`
	RunningStyle int `json:"running_style"`
	Mood         int `json:"motivation"`
	Skills       []struct {
		Id int `json:"skill_id"`
	} `json:"skill_array"`
}

// DecodeChampionsRace reads the course, seed and runners of a saved Champions Meeting race
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
		Course:  file.CourseSet,
		Seed:    int32(file.RandomSeed),
		Runners: decodeRunners(file.Runners),
	}, nil
}

// decodeRunners returns what the application shows of each runner
func decodeRunners(runners []championsRunner) []Runner {
	decoded := make([]Runner, len(runners))
	for i := range runners {
		entry := &runners[i].Entry
		skillIds := make([]int, len(entry.Skills))
		for k := range entry.Skills {
			skillIds[k] = entry.Skills[k].Id
		}
		decoded[i] = Runner{
			Name:         runners[i].CharaName,
			Trainer:      runners[i].TrainerName,
			RunningStyle: entry.RunningStyle,
			Runaway:      runners[i].RunningStyleEx == runawayStyleEx,
			Mood:         entry.Mood,
			Speed:        entry.Speed,
			Stamina:      entry.Stamina,
			Power:        entry.Power,
			Guts:         entry.Guts,
			Wit:          entry.Wit,
			SkillIds:     skillIds,
		}
	}
	return decoded
}
