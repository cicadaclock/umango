package races

import (
	"encoding/json"
	"fmt"
	"os"
)

// Result of a single team trial round
type RaceResult struct {
	DistanceType int `json:"distance_type"`
	// Base64-encoded compressed race scenario blob
	RaceScenario               string        `json:"race_scenario"`
	Round                      int           `json:"round"`
	TeamTotalScore             int           `json:"team_total_score"`
	WinType                    int           `json:"win_type"`
	CurrentConsecutiveWinCount int           `json:"current_consecutive_win_count"`
	BonusRateByNextWin         int           `json:"bonus_rate_by_next_win"`
	CharaResultArray           []CharaResult `json:"chara_result_array"`
}

// Result of a single uma in a team trial round
type CharaResult struct {
	FrameOrder     int `json:"frame_order"`
	TrainedCharaId int `json:"trained_chara_id"`
	TeamId         int `json:"team_id"`
	FinishOrder    int `json:"finish_order"`
	FinishTime     int `json:"finish_time"`
	// Scores earned during the race
	ScoreArray []Score `json:"score_array"`
}

// Score earned from a single scoring event
type Score struct {
	RawScoreId int `json:"raw_score_id"`
	// Number of times the scoring event occurred
	Num   int `json:"num"`
	Score int `json:"score"`
	// Bonuses applied on top of the raw score
	BonusArray []ScoreBonus `json:"bonus_array"`
}

// Bonus applied to a score
type ScoreBonus struct {
	ScoreBonusId    int `json:"score_bonus_id"`
	BonusScore      int `json:"bonus_score"`
	ConditionType   int `json:"condition_type"`
	ConditionValue1 int `json:"condition_value_1"`
	ConditionValue2 int `json:"condition_value_2"`
	ScoreRate       int `json:"score_rate"`
}

func LoadRaceResults(path string) ([]RaceResult, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	var teamTrial struct {
		RaceResultArray []RaceResult `json:"race_result_array"`
	}
	if err := json.Unmarshal(file, &teamTrial); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return teamTrial.RaceResultArray, nil
}

// Charas that earned scores, which is only the player's own 3 umas per round
func (raceResult RaceResult) ScoredCharas() []CharaResult {
	scored := make([]CharaResult, 0, len(raceResult.CharaResultArray))
	for _, charaResult := range raceResult.CharaResultArray {
		if len(charaResult.ScoreArray) > 0 {
			scored = append(scored, charaResult)
		}
	}
	return scored
}
