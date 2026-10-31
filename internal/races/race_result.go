package races

// Result of a single team trial round
type RaceResult struct {
	DistanceType DistanceType `json:"distance_type"`
	// Base64-encoded compressed race scenario blob
	//
	// Format is base64 -> gzip -> little-endian binary
	RaceScenario     string        `json:"race_scenario"`
	Round            int           `json:"round"`
	CharaResultArray []CharaResult `json:"chara_result_array"`
}

// Result of a single uma in a team trial round
type CharaResult struct {
	TrainedCharaId int `json:"trained_chara_id"`
	TeamId         int `json:"team_id"`
	// Scoring events earned during the race
	ScoreEventArray []ScoreEvent `json:"score_array"`
}

type ScoreEvent struct {
	// Type of score (lengths, placement, skills, rushed, etc.)
	RawScoreId int `json:"raw_score_id"`
	// Final score value for a given scoring event: its base score times the number of times it
	// occurred, plus its bonuses
	Score int `json:"score"`
	// Bonuses comprising the raw score (minus base score)
	BonusArray []ScoreBonus `json:"bonus_array"`
}

// Bonus applied to a score
type ScoreBonus struct {
	// Type of bonus score (Opponent rating, support bonus, ace bonus, streak)
	ScoreBonusId int `json:"score_bonus_id"`
	Score        int `json:"bonus_score"`
}

// Find chara results from a single race
func (raceResult RaceResult) FindCharaResults(trainedCharaId int) CharaResult {
	for _, charaResult := range raceResult.CharaResultArray {
		if charaResult.TrainedCharaId == trainedCharaId {
			return charaResult
		}
	}
	return CharaResult{}
}

// Total bonus score of all charas in a single race
func (raceResult RaceResult) BonusScore() int {
	sum := 0
	for _, charaResult := range raceResult.CharaResultArray {
		sum += charaResult.BonusScore()
	}
	return sum
}

// Total score of a single chara
func (charaResult CharaResult) TotalScore() int {
	sum := 0
	for _, scoreEvent := range charaResult.ScoreEventArray {
		sum += scoreEvent.Score
	}
	return sum
}

// Bonus score of a single chara
func (charaResult CharaResult) BonusScore() int {
	sum := 0
	for _, scoreEvent := range charaResult.ScoreEventArray {
		for _, bonus := range scoreEvent.BonusArray {
			sum += bonus.Score
		}
	}
	return sum
}

// Decodes the RaceScenario blob of this race
func (raceResult RaceResult) DecodeScenario() (RaceScenario, error) {
	return DecodeRaceScenario(raceResult.RaceScenario)
}
