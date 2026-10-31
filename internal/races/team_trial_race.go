// Package races reads team trial result files and summarizes the results of each veteran.
package races

import (
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
)

// TeamTrialResultSet holds the loaded team trial results in file order, oldest first
type TeamTrialResultSet struct {
	Set []TeamTrialResult
}

// TeamTrialResult is one saved team trial of 5 rounds. Only the fields that the application
// reads are declared, so decoding skips the rest of each file.
type TeamTrialResult struct {
	RaceStartParamsArray []RaceStartParams `json:"race_start_params_array"`
	RaceResultArray      []RaceResult      `json:"race_result_array"`
}

// append adds one team trial result
func (ttrs *TeamTrialResultSet) append(ttr TeamTrialResult) {
	ttrs.Set = append(ttrs.Set, ttr)
}

// TTUmaSummary holds the aggregated results of each veteran, one slice per field, indexed by
// summary index. The fielded team comes first.
type TTUmaSummary struct {
	TrainedCharaIds []int
	// Entry data from the newest round of each veteran
	CharaData []RaceHorseData
	// Total score of each race
	Scores []ScoreArray
	// Points by score category, summed over every race
	ScoreBreakdowns []ScoreBreakdown
	// Distance category of the newest round
	DistanceTypes []DistanceType
	// Skill ID to activation count, summed over every race
	SkillActivations []map[int]int
	// Races whose course and race scenario are known, and those of them with a full spurt
	SpurtRaces []int
	FullSpurts []int
	// Number of veterans on the fielded team
	FieldedCount int
}

// Len returns the number of veterans
func (s TTUmaSummary) Len() int {
	return len(s.TrainedCharaIds)
}

// CourseMapper provides the course distances that Summarize needs
type CourseMapper interface {
	// Course distance in meters of a race instance ID, or 0 when unknown
	CourseDistance(raceInstanceId int) int
}

// decodedScenario holds the parts of a race scenario that Summarize reads
type decodedScenario struct {
	horseResults []HorseResult
	activations  []SkillActivation
	err          error
}

// decodeScenarios decodes the race scenario of every round on all cores. Round r of result i
// is at index i*5+r; results without 5 rounds are left zero.
func (ttrs TeamTrialResultSet) decodeScenarios() []decodedScenario {
	decoded := make([]decodedScenario, 5*len(ttrs.Set))
	var next atomic.Int64
	var workers sync.WaitGroup
	for range runtime.GOMAXPROCS(0) {
		workers.Go(func() {
			decoder := ScenarioDecoder{SkipFrames: true}
			for {
				i := int(next.Add(1) - 1)
				if i >= len(decoded) {
					return
				}
				ttr := &ttrs.Set[i/5]
				if !ttr.hasCorrectRaceCount() {
					continue
				}
				scenario, err := decoder.Decode(ttr.RaceResultArray[i%5].RaceScenario)
				decoded[i] = decodedScenario{scenario.HorseResults, scenario.SkillActivations(), err}
			}
		})
	}
	workers.Wait()
	return decoded
}

// Summarize aggregates the team trial results into the results of each veteran. courses
// provides the course distances that judge full spurts.
func (ttrs TeamTrialResultSet) Summarize(courses CourseMapper) TTUmaSummary {
	s := TTUmaSummary{}
	rows := make(map[int]int, 50)
	decoded := ttrs.decodeScenarios()
	// Newest first, so the first result read holds the fielded team
	newest := true
	for set, ttr := range slices.Backward(ttrs.Set) {
		if !ttr.hasCorrectRaceCount() {
			continue
		}
		for round := range 5 {
			raceResult := ttr.RaceResultArray[round]
			umas := ttr.RaceStartParamsArray[round].GetMyUmas()
			if newest {
				s.FieldedCount += len(umas)
			}

			scenario := &decoded[set*5+round]
			course := courses.CourseDistance(ttr.RaceStartParamsArray[round].RaceInstanceId)
			spurtKnown := scenario.err == nil && course > 0
			// Skill IDs activated by each horse index in the round
			activationMap := make(map[int][]int, 12)
			for _, activation := range scenario.activations {
				activationMap[activation.HorseIndex] = append(activationMap[activation.HorseIndex], activation.SkillId)
			}

			for _, uma := range umas {
				// The newest round of a veteran adds its row
				index, exists := rows[uma.TrainedCharaId]
				if !exists {
					index = s.Len()
					rows[uma.TrainedCharaId] = index
					s.TrainedCharaIds = append(s.TrainedCharaIds, uma.TrainedCharaId)
					s.CharaData = append(s.CharaData, uma)
					s.Scores = append(s.Scores, ScoreArray{})
					s.ScoreBreakdowns = append(s.ScoreBreakdowns, ScoreBreakdown{})
					s.DistanceTypes = append(s.DistanceTypes, raceResult.DistanceType)
					activations := make(map[int]int)
					skillIds := activationMap[uma.FrameOrder-1]
					for _, skillId := range skillIds {
						activations[skillId]++
					}
					s.SkillActivations = append(s.SkillActivations, activations)
					s.SpurtRaces = append(s.SpurtRaces, 0)
					s.FullSpurts = append(s.FullSpurts, 0)
				}
				if result := raceResult.FindCharaResults(uma.TrainedCharaId); len(result.ScoreEventArray) > 0 {
					s.Scores[index].append(result.TotalScore())
					s.ScoreBreakdowns[index].add(result.ScoreEventArray)
					skillIds := activationMap[uma.FrameOrder-1]
					for _, skillId := range skillIds {
						s.SkillActivations[index][skillId]++
					}
					// Horse index of the uma in the race scenario
					horse := uma.FrameOrder - 1
					if spurtKnown && horse >= 0 && horse < len(scenario.horseResults) {
						s.SpurtRaces[index]++
						if scenario.horseResults[horse].fullSpurt(course) {
							s.FullSpurts[index]++
						}
					}
				}
			}
		}
		if newest {
			newest = false
		}
	}
	return s
}

// IsValidData reports whether the result has 5 rounds whose start parameters and results agree
func (ttr TeamTrialResult) IsValidData() bool {
	return ttr.hasMatchingRounds() && ttr.hasCorrectRaceCount()
}

// hasMatchingRounds reports whether the start parameters and results at each index belong to
// the same round
func (ttr TeamTrialResult) hasMatchingRounds() bool {
	for i := range 5 {
		if ttr.RaceStartParamsArray[i].Round != ttr.RaceResultArray[i].Round {
			return false
		}
	}
	return true
}

// hasCorrectRaceCount reports whether the result has 5 rounds of start parameters and results
func (ttr TeamTrialResult) hasCorrectRaceCount() bool {
	return len(ttr.RaceStartParamsArray) == 5 && len(ttr.RaceResultArray) == 5
}
