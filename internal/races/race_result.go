package races

// RaceResult is the result of one team trial round
type RaceResult struct {
	DistanceType DistanceType `json:"distance_type"`
	// Encoded race scenario: base64 of gzip of the binary scenario format
	RaceScenario     string        `json:"race_scenario"`
	Round            int           `json:"round"`
	CharaResultArray []CharaResult `json:"chara_result_array"`
}

// CharaResult is the result of one uma in a team trial round
type CharaResult struct {
	TrainedCharaId  int          `json:"trained_chara_id"`
	TeamId          int          `json:"team_id"`
	ScoreEventArray []ScoreEvent `json:"score_array"`
}

// ScoreEvent is one score event earned by an uma, such as a placement or a skill activation
type ScoreEvent struct {
	// team_stadium_raw_score ID in master.mdb
	RawScoreId int `json:"raw_score_id"`
	// Base score times the number of occurrences, plus every bonus in BonusArray
	Score      int          `json:"score"`
	BonusArray []ScoreBonus `json:"bonus_array"`
}

// ScoreBonus is one bonus included in the score of a score event
type ScoreBonus struct {
	// team_stadium_score_bonus ID in master.mdb
	ScoreBonusId int `json:"score_bonus_id"`
	Score        int `json:"bonus_score"`
}

// FindCharaResults returns the result of the uma with the trained chara ID, or a zero
// CharaResult when the uma did not race in the round
func (raceResult RaceResult) FindCharaResults(trainedCharaId int) CharaResult {
	for _, charaResult := range raceResult.CharaResultArray {
		if charaResult.TrainedCharaId == trainedCharaId {
			return charaResult
		}
	}
	return CharaResult{}
}

// BonusScore returns the total bonus score of every uma in the round
func (raceResult RaceResult) BonusScore() int {
	sum := 0
	for _, charaResult := range raceResult.CharaResultArray {
		sum += charaResult.BonusScore()
	}
	return sum
}

// TotalScore returns the total score of the uma in the round
func (charaResult CharaResult) TotalScore() int {
	sum := 0
	for _, scoreEvent := range charaResult.ScoreEventArray {
		sum += scoreEvent.Score
	}
	return sum
}

// BonusScore returns the bonus score of the uma in the round
func (charaResult CharaResult) BonusScore() int {
	sum := 0
	for _, scoreEvent := range charaResult.ScoreEventArray {
		for _, bonus := range scoreEvent.BonusArray {
			sum += bonus.Score
		}
	}
	return sum
}

// DecodeScenario decodes the race scenario of the round
func (raceResult RaceResult) DecodeScenario() (RaceScenario, error) {
	return DecodeRaceScenario(raceResult.RaceScenario)
}
