package races

// Result of a single team trial round
type RaceResult struct {
	DistanceType int `json:"distance_type"`
	// Base64-encoded compressed race scenario blob
	RaceScenario   string `json:"race_scenario"`
	Round          int    `json:"round"`
	TeamTotalScore int    `json:"team_total_score"`
	// Win (1) or loss (2)
	WinType                    int           `json:"win_type"`
	CurrentConsecutiveWinCount int           `json:"current_consecutive_win_count"`
	BonusRateByNextWin         int           `json:"bonus_rate_by_next_win"`
	CharaResultArray           []CharaResult `json:"chara_result_array"`
}

// Result of a single uma in a team trial round
type CharaResult struct {
	// Starting gate
	FrameOrder     int `json:"frame_order"`
	TrainedCharaId int `json:"trained_chara_id"`
	TeamId         int `json:"team_id"`
	// Placing in race
	FinishOrder int `json:"finish_order"`
	FinishTime  int `json:"finish_time"`
	// Scoring events earned during the race
	ScoreEventArray []ScoreEvent `json:"score_array"`
}

type ScoreEvent struct {
	// Type of score (lengths, placement, skills, rushed, etc.)
	RawScoreId int `json:"raw_score_id"`
	// Number of times the scoring event occurred
	Num int `json:"num"`
	// Final score value for a given scoring event. Score = BonusScores + (BaseScore * Num)
	Score int `json:"score"`
	// Bonuses comprising the raw score (minus base score)
	BonusArray []ScoreBonus `json:"bonus_array"`
}

// Bonus applied to a score
type ScoreBonus struct {
	// Type of bonus score (Opponent rating, support bonus, ace bonus, streak)
	ScoreBonusId int `json:"score_bonus_id"`
	Score        int `json:"bonus_score"`
	// No idea what these do
	ConditionType   int `json:"condition_type"`
	ConditionValue1 int `json:"condition_value_1"`
	ConditionValue2 int `json:"condition_value_2"`
	ScoreRate       int `json:"score_rate"`
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
