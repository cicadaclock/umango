package races

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestRaceResultsSoA(t *testing.T) {
	path := filepath.Join("..", "testdata", "team_trial.json")
	results, err := LoadRaceResults(path)
	if err != nil {
		t.Fatalf("load race results: %v", err)
	}

	soa := NewRaceResultsSoA(results)
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

func TestCharaResultSoAAppend(t *testing.T) {
	c := CharaResultSoA{}

	c.TotalScore.append(10)
	if !reflect.DeepEqual(c.TotalScore.Score, []int{10}) {
		t.Errorf("TotalScore.Score == %v, want [10]", c.TotalScore.Score)
	}
}

func TestCharaResultSoAAverage(t *testing.T) {
	c := CharaResultSoA{}
	c.TotalScore.append(10)
	c.TotalScore.append(20)
	if c.TotalScore.Sum != 30 {
		t.Errorf("TotalScore.Sum == %d, want 30", c.TotalScore.Sum)
	}
	if c.TotalScore.Average() != 15 {
		t.Errorf("TotalScore.Average() == %d, want 15", c.TotalScore.Average())
	}
}

func TestCharaResultSoALen(t *testing.T) {
	c := CharaResultSoA{}
	c.TotalScore.append(10)
	c.TotalScore.append(20)
	c.TotalScore.append(30)
	if c.TotalScore.Len() != 3 {
		t.Errorf("TotalScore.Len() == %d, want 3", c.TotalScore.Len())
	}
}

func TestCharaResultSoAFilter(t *testing.T) {
	c := CharaResultSoA{}
	c.TotalScore.append(10)
	c.TotalScore.append(20)
	c.TotalScore.append(30)
	c.TotalScore.append(40)
	filteredScore := c.TotalScore.Filter([]int{1, 2})
	if !reflect.DeepEqual(filteredScore.Score, []int{20, 30}) {
		t.Errorf("filteredScore == %v, want [20 30]", filteredScore)
	}
}
