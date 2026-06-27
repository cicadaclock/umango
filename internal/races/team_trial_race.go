package races

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

// Checks if RaceStartParamsArray and RaceResultArray rounds are matched
//
// Guarantees data processing can occur on both arrays using the same index
func (ttr TeamTrialResult) IsInAscendingOrder() bool {
	for i := range 5 {
		if ttr.RaceStartParamsArray[i].Round != ttr.RaceResultArray[i].Round {
			return false
		}
	}
	return true
}

// Checks if 5 races occur in a team trial result like expected
func (ttr TeamTrialResult) HasCorrectRaceCount() bool {
	if len(ttr.RaceStartParamsArray) != 5 {
		return false
	}
	if len(ttr.RaceResultArray) != 5 {
		return false
	}
	return true
}

// Returns CharaResults for up to 15 umas in the race
func (ttr TeamTrialResult) GetMyCharaResults() []CharaResult {
	charaResults := make([]CharaResult, 0, 15)
	for i := range 5 {
		umas := ttr.RaceStartParamsArray[i].GetMyUmas()
		for _, uma := range umas {
			charaResults = append(charaResults, ttr.RaceResultArray[i].FindCharaResults(uma.TrainedCharaId))
		}
	}
	return charaResults
}
