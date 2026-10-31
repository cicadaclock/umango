package races

import "slices"

type TeamTrialResultSet struct {
	Set []TeamTrialResult
}

type TeamTrialResult struct {
	UseItemIdArray       []int             `json:"use_item_id_array"`
	ItemInfoArray        []int             `json:"item_info_array"`
	SupportCardBonus     int               `json:"support_card_bonus"`
	RaceStartParamsArray []RaceStartParams `json:"race_start_params_array"`
	RaceResultArray      []RaceResult      `json:"race_result_array"`
}

func (ttrs *TeamTrialResultSet) append(ttr TeamTrialResult) {
	ttrs.Set = append(ttrs.Set, ttr)
}

// TTUmaSummary aggregates team trial score results per uma
// so that the TTRS only need to be walked once
type TTUmaSummary struct {
	// TrainedCharaIds is also in CharaData, but it's simpler to have a slice
	TrainedCharaIds []int
	CharaData       []RaceHorseData
	Scores          []ScoreArray
	// Points by category summed over the races in Scores
	ScoreBreakdowns  []ScoreBreakdown
	DistanceTypes    []DistanceType
	SkillActivations []map[int]int
	// Races in Scores whose replay and course are known, and those of them in which the uma
	// started its last spurt on entering the final leg
	SpurtRaces []int
	FullSpurts []int
	// FieldedCount counts the number of umas we are fielding on our team
	FieldedCount int
}

func (s TTUmaSummary) Len() int {
	return len(s.TrainedCharaIds)
}

// CourseMapper provides the course data Summarize needs from DataStore
type CourseMapper interface {
	// CourseDistance maps a race instance ID to the length of its course in meters, 0 when
	// unknown
	CourseDistance(raceInstanceId int) int
}

// Summarize walks TTRS once, collecting anything needed into the TTUmaSummary
//
// Inputs:
// - courses: Course lengths, which judge whether each uma's last spurt started on time
func (ttrs TeamTrialResultSet) Summarize(courses CourseMapper) TTUmaSummary {
	s := TTUmaSummary{}
	rows := make(map[int]int, 50)
	// Iterate backwards so that the latest results are first
	// This is needed to figure out the currently used umas
	newest := true
	for _, ttr := range slices.Backward(ttrs.Set) {
		if !ttr.hasCorrectRaceCount() {
			continue
		}
		for round := range 5 {
			raceResult := ttr.RaceResultArray[round]
			umas := ttr.RaceStartParamsArray[round].GetMyUmas()
			// Count number of umas we are fielding
			if newest {
				s.FieldedCount += len(umas)
			}

			// Parse race scenario blob
			scenario, err := raceResult.DecodeScenario()
			course := courses.CourseDistance(ttr.RaceStartParamsArray[round].RaceInstanceId)
			spurtKnown := err == nil && course > 0
			activations := scenario.SkillActivations()
			// Skill IDs activated per uma in the race scenario
			activationMap := make(map[int][]int, 12)
			for _, activation := range activations {
				activationMap[activation.HorseIndex] = append(activationMap[activation.HorseIndex], activation.SkillId)
			}

			// Add uma result to summary
			for _, uma := range umas {
				index, exists := rows[uma.TrainedCharaId]
				if !exists {
					index = s.Len()
					rows[uma.TrainedCharaId] = index
					s.TrainedCharaIds = append(s.TrainedCharaIds, uma.TrainedCharaId)
					s.CharaData = append(s.CharaData, uma)
					s.Scores = append(s.Scores, ScoreArray{})
					s.ScoreBreakdowns = append(s.ScoreBreakdowns, ScoreBreakdown{})
					s.DistanceTypes = append(s.DistanceTypes, raceResult.DistanceType)
					// Count skill activations per skill ID
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
					horse := uma.FrameOrder - 1
					if spurtKnown && horse >= 0 && horse < len(scenario.HorseResults) {
						s.SpurtRaces[index]++
						if scenario.HorseResults[horse].fullSpurt(course) {
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

// IsValidData returns whether this result has 5 races and matches rounds
func (ttr TeamTrialResult) IsValidData() bool {
	return ttr.hasMatchingRounds() && ttr.hasCorrectRaceCount()
}

// hasMatchingRounds checks if RaceStartParamsArray and RaceResultArray rounds
// are matched
//
// Guarantees data processing can occur on both arrays using the same index
func (ttr TeamTrialResult) hasMatchingRounds() bool {
	for i := range 5 {
		if ttr.RaceStartParamsArray[i].Round != ttr.RaceResultArray[i].Round {
			return false
		}
	}
	return true
}

// hasCorrectRaceCount checks if 5 races occur in a team trial result
func (ttr TeamTrialResult) hasCorrectRaceCount() bool {
	return len(ttr.RaceStartParamsArray) == 5 && len(ttr.RaceResultArray) == 5
}
