package races

import (
	"path/filepath"
	"testing"
)

func TestRaceResultsSoA(t *testing.T) {
	path := filepath.Join("..", "testdata", "team_trial.json")
	results, err := LoadRaces(path)
	if err != nil {
		t.Fatalf("load race results: %v", err)
	}

	soa := NewRaceResultsSoA(results.RaceResultArray)
	if soa.Len() != 5 {
		t.Fatalf("Len() == %d, want 5", soa.Len())
	}
	if soa.TeamTotalScores.Len() != 5 || len(soa.CharaResultArrays) != 5 {
		t.Fatal("SoA slices not parallel to race count")
	}

	// Mile race (distance type 1) is round 2
	mile := soa.FilterByDistanceType(1)
	if mile.Len() != 1 {
		t.Fatalf("mile.Len() == %d, want 1", mile.Len())
	}
	if mile.TeamTotalScores.Get(0) != 142642 {
		t.Errorf("TeamTotalScores[0] == %d, want 142642", mile.TeamTotalScores.Get(0))
	}

	// Chara 2984 only raced in round 1
	chara := soa.FilterByTrainedCharaId(2984)
	if chara.Len() != 1 {
		t.Fatalf("chara.Len() == %d, want 1", chara.Len())
	}
	if len(chara.CharaResultArrays[0]) != 1 {
		t.Fatalf("len(CharaResultArrays[0]) == %d, want 1", len(chara.CharaResultArrays[0]))
	}
	totalScores := chara.CharaTotalScores()
	if len(totalScores) != 1 || totalScores[0] != 56084 {
		t.Errorf("CharaTotalScores() == %v, want [56084]", totalScores)
	}

	// Unknown ids filter everything out
	if soa.FilterByDistanceType(99).Len() != 0 {
		t.Error("want empty SoA for unknown distance type")
	}
	if soa.FilterByTrainedCharaId(-1).Len() != 0 {
		t.Error("want empty SoA for unknown trained chara id")
	}
}
