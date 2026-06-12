package races

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

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

func LoadRaceResultsFolder(directoryPath string) ([]RaceResult, error) {
	// 20 TT samples, 5 race results per TT sample = 100 results
	allRaceResults := make([]RaceResult, 0, 100)
	filepath.WalkDir(directoryPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		} else if filepath.Ext(d.Name()) != ".json" {
			return nil
		}

		raceResults, err := LoadRaceResults(path)
		// If parsing race results fails, just skip it
		if err != nil {
			return nil
		}
		allRaceResults = append(allRaceResults, raceResults...)
		return nil
	})
	return allRaceResults, nil
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
