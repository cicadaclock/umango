package racesim

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

const (
	// Saved team trial result files, whose names sort in order of time
	teamTrialFilePattern = "TT-*.json"
	// team_id of the player's team in a team trial result
	playerTeamId = 1
	// raceType of a team trial race in a saved race
	teamStadiumRaceType = "TeamStadium"
	// Time of day of every race in the team trial pool: Daytime
	teamTrialRaceTime = 2
)

// TeamRound is one round of a saved team trial, in the form that the library reads
type TeamRound struct {
	// Team trial category, numbered as races.DistanceType
	Category int
	// Seed of the recorded race
	Seed int64
	// Runners in the race
	RunnerCount int
	// The round as a saved race, for Engine.Prepare
	SavedRace []byte
	// The player's runners in team order
	Team []TeamRunner
}

// TeamRunner is one runner of the player's team in a team trial round
type TeamRunner struct {
	// Index of the runner in the saved race
	Index   int
	CharaId int
	// Skills that the runner learned
	SkillIds []int
}

// teamTrialFile holds the fields of a team trial result file that this package reads
type teamTrialFile struct {
	Starts []struct {
		Round           int               `json:"round"`
		RaceInstanceId  int               `json:"race_instance_id"`
		Season          int               `json:"season"`
		Weather         int               `json:"weather"`
		GroundCondition int               `json:"ground_condition"`
		RandomSeed      int64             `json:"random_seed"`
		Horses          []json.RawMessage `json:"race_horse_data_array"`
	} `json:"race_start_params_array"`
	Results []teamTrialResult `json:"race_result_array"`
}

// teamTrialResult holds the fields of the result of one round that this package reads
type teamTrialResult struct {
	Round        int    `json:"round"`
	DistanceType int    `json:"distance_type"`
	Scenario     string `json:"race_scenario"`
}

// teamTrialHorse holds the fields of one runner of a team trial round that this package reads
type teamTrialHorse struct {
	TeamId       int `json:"team_id"`
	TeamMemberId int `json:"team_member_id"`
	FrameOrder   int `json:"frame_order"`
	CharaId      int `json:"chara_id"`
	Skills       []struct {
		Id int `json:"skill_id"`
	} `json:"skill_array"`
}

// savedTeamRace is a team trial round in the layout of a saved race. Skill trials replace its
// course, so the course is left unset.
type savedTeamRace struct {
	RaceType   string `json:"raceType"`
	RandomSeed int64  `json:"randomSeed"`
	CourseSet  struct {
		Id int `json:"id"`
	} `json:"raceCourseSet"`
	GroundCondition    int              `json:"groundCondition"`
	Weather            int              `json:"weather"`
	Season             int              `json:"season"`
	Time               int              `json:"time"`
	Runners            []savedTeamHorse `json:"raceHorse"`
	RaceInstanceMaster struct {
		Id   int `json:"id"`
		Time int `json:"time"`
	} `json:"raceInstanceMaster"`
	Scenario string `json:"simDataBase64"`
}

// savedTeamHorse is one runner of a savedTeamRace
type savedTeamHorse struct {
	HorseIndex int             `json:"horseIndex"`
	Entry      json.RawMessage `json:"responseHorseData"`
}

// ReadLatestTeamRounds returns the rounds of the latest team trial result file in a folder
func ReadLatestTeamRounds(directory string) ([]TeamRound, error) {
	paths, err := filepath.Glob(filepath.Join(directory, teamTrialFilePattern))
	if err != nil {
		return nil, fmt.Errorf("list team trial results: %w", err)
	}
	if len(paths) == 0 {
		return nil, errors.New("no saved team trial results")
	}
	latest := slices.Max(paths)
	file, err := os.ReadFile(latest)
	if err != nil {
		return nil, fmt.Errorf("read team trial result: %w", err)
	}
	rounds, err := decodeTeamRounds(file)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", filepath.Base(latest), err)
	}
	return rounds, nil
}

// decodeTeamRounds returns the rounds of a team trial result file
func decodeTeamRounds(file []byte) ([]TeamRound, error) {
	var trial teamTrialFile
	if err := json.Unmarshal(file, &trial); err != nil {
		return nil, err
	}
	rounds := make([]TeamRound, 0, len(trial.Starts))
	for i := range trial.Starts {
		start := &trial.Starts[i]
		k := slices.IndexFunc(trial.Results, func(r teamTrialResult) bool { return r.Round == start.Round })
		if k < 0 {
			return nil, fmt.Errorf("round %d has no result", start.Round)
		}
		horses := make([]teamTrialHorse, len(start.Horses))
		order := make([]int, len(start.Horses))
		for h := range start.Horses {
			if err := json.Unmarshal(start.Horses[h], &horses[h]); err != nil {
				return nil, fmt.Errorf("round %d: %w", start.Round, err)
			}
			order[h] = h
		}
		slices.SortStableFunc(order, func(a, b int) int {
			return cmp.Compare(horses[a].FrameOrder, horses[b].FrameOrder)
		})
		saved := savedTeamRace{
			RaceType:        teamStadiumRaceType,
			RandomSeed:      start.RandomSeed,
			GroundCondition: start.GroundCondition,
			Weather:         start.Weather,
			Season:          start.Season,
			Time:            teamTrialRaceTime,
			Runners:         make([]savedTeamHorse, len(order)),
			Scenario:        trial.Results[k].Scenario,
		}
		saved.RaceInstanceMaster.Id = start.RaceInstanceId
		saved.RaceInstanceMaster.Time = teamTrialRaceTime
		for index, h := range order {
			saved.Runners[index] = savedTeamHorse{HorseIndex: index, Entry: start.Horses[h]}
		}
		savedRace, err := json.Marshal(saved)
		if err != nil {
			return nil, fmt.Errorf("round %d: %w", start.Round, err)
		}
		rounds = append(rounds, TeamRound{
			Category:    trial.Results[k].DistanceType,
			Seed:        start.RandomSeed,
			RunnerCount: len(order),
			SavedRace:   savedRace,
			Team:        playerTeam(horses, order),
		})
	}
	return rounds, nil
}

// playerTeam returns the player's runners in team order. The indexes follow order, which lists
// the horses in gate order.
func playerTeam(horses []teamTrialHorse, order []int) []TeamRunner {
	team := make([]TeamRunner, 0, 3)
	members := make([]int, 0, 3)
	for index, h := range order {
		if horses[h].TeamId != playerTeamId {
			continue
		}
		skillIds := make([]int, len(horses[h].Skills))
		for k := range horses[h].Skills {
			skillIds[k] = horses[h].Skills[k].Id
		}
		position, _ := slices.BinarySearch(members, horses[h].TeamMemberId)
		members = slices.Insert(members, position, horses[h].TeamMemberId)
		team = slices.Insert(team, position, TeamRunner{Index: index, CharaId: horses[h].CharaId, SkillIds: skillIds})
	}
	return team
}
