package races

import (
	"path/filepath"
	"testing"
)

func TestLoadRaceResults(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "team_trial.json")
	results, err := LoadRaceResults(path)
	if err != nil {
		t.Fatalf("load race results: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("len(results) == %d, want 5", len(results))
	}

	round1 := results[0]
	if round1.Round != 1 {
		t.Errorf("Round == %d, want 1", round1.Round)
	}
	if round1.TeamTotalScore != 134430 {
		t.Errorf("TeamTotalScore == %d, want 134430", round1.TeamTotalScore)
	}
	if len(round1.CharaResultArray) != 12 {
		t.Fatalf("len(CharaResultArray) == %d, want 12", len(round1.CharaResultArray))
	}

	chara := round1.CharaResultArray[2]
	if chara.TrainedCharaId != 2984 {
		t.Errorf("TrainedCharaId == %d, want 2984", chara.TrainedCharaId)
	}
	if len(chara.ScoreArray) != 8 {
		t.Fatalf("len(ScoreArray) == %d, want 8", len(chara.ScoreArray))
	}
	if chara.ScoreArray[0].BonusArray[0].BonusScore != 5940 {
		t.Errorf("BonusScore == %d, want 5940", chara.ScoreArray[0].BonusArray[0].BonusScore)
	}
}

func TestLoadRaceResultsForEmptyPath(t *testing.T) {
	_, err := LoadRaceResults("")
	if err == nil {
		t.Error("want error for empty path")
	}
}
