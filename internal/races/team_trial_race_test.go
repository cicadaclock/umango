package races

import (
	"reflect"
	"testing"
)

// ttResult mirrors a real team trial, reduced to the fields Summarize
// parses. Each round has 3 of our umas alongside 2 opponents
var ttResult = TeamTrialResult{
	RaceStartParamsArray: []RaceStartParams{
		{
			Round: 1,
			RaceHorseDataArray: []RaceHorseData{
				{TeamId: 1, TrainedCharaId: 2984, CardId: 100402, RunningStyle: Front},
				{TeamId: 1, TrainedCharaId: 2928, CardId: 105001, RunningStyle: End},
				{TeamId: 1, TrainedCharaId: 2846, CardId: 100602, RunningStyle: Pace},
				{TeamId: 0, TrainedCharaId: 56},
				{TeamId: 2, TrainedCharaId: 1625},
			},
		},
		{
			Round: 2,
			RaceHorseDataArray: []RaceHorseData{
				{TeamId: 1, TrainedCharaId: 2977, CardId: 101801, RunningStyle: Late},
				{TeamId: 1, TrainedCharaId: 2994, CardId: 104101, RunningStyle: Pace},
				{TeamId: 1, TrainedCharaId: 3106, CardId: 105101, RunningStyle: Pace},
				{TeamId: 2, TrainedCharaId: 1642},
				{TeamId: 0, TrainedCharaId: 28},
			},
		},
		{
			Round: 3,
			RaceHorseDataArray: []RaceHorseData{
				{TeamId: 1, TrainedCharaId: 2924, CardId: 103201, RunningStyle: Pace},
				{TeamId: 1, TrainedCharaId: 3049, CardId: 101101, RunningStyle: Late},
				{TeamId: 1, TrainedCharaId: 3041, CardId: 101701, RunningStyle: Late},
				{TeamId: 2, TrainedCharaId: 1455},
				{TeamId: 2, TrainedCharaId: 1613},
			},
		},
		{
			Round: 4,
			RaceHorseDataArray: []RaceHorseData{
				{TeamId: 1, TrainedCharaId: 3002, CardId: 106801, RunningStyle: Front},
				{TeamId: 1, TrainedCharaId: 2939, CardId: 102501, RunningStyle: End},
				{TeamId: 1, TrainedCharaId: 3060, CardId: 100701, RunningStyle: End},
				{TeamId: 0, TrainedCharaId: 120},
				{TeamId: 2, TrainedCharaId: 1504},
			},
		},
		{
			Round: 5,
			RaceHorseDataArray: []RaceHorseData{
				{TeamId: 1, TrainedCharaId: 2935, CardId: 104601, RunningStyle: Front},
				{TeamId: 1, TrainedCharaId: 3056, CardId: 101401, RunningStyle: Pace},
				{TeamId: 1, TrainedCharaId: 3102, CardId: 101901, RunningStyle: Late},
				{TeamId: 0, TrainedCharaId: 146},
				{TeamId: 2, TrainedCharaId: 1208},
			}},
	},
	RaceResultArray: []RaceResult{
		{
			Round:        1,
			DistanceType: Mile,
			CharaResultArray: []CharaResult{
				{TeamId: 1, TrainedCharaId: 2984, ScoreEventArray: []ScoreEvent{{Score: 56084}}},
				{TeamId: 1, TrainedCharaId: 2928, ScoreEventArray: []ScoreEvent{{Score: 43787}}},
				{TeamId: 1, TrainedCharaId: 2846, ScoreEventArray: []ScoreEvent{{Score: 34559}}},
				{TeamId: 0, TrainedCharaId: 56, ScoreEventArray: []ScoreEvent{}},
				{TeamId: 2, TrainedCharaId: 1625, ScoreEventArray: []ScoreEvent{}},
			},
		},
		{
			Round:        2,
			DistanceType: Sprint,
			CharaResultArray: []CharaResult{
				{TeamId: 1, TrainedCharaId: 2977, ScoreEventArray: []ScoreEvent{{Score: 47721}}},
				{TeamId: 1, TrainedCharaId: 2994, ScoreEventArray: []ScoreEvent{{Score: 45260}}},
				{TeamId: 1, TrainedCharaId: 3106, ScoreEventArray: []ScoreEvent{{Score: 49661}}},
				{TeamId: 2, TrainedCharaId: 1642, ScoreEventArray: []ScoreEvent{}},
				{TeamId: 0, TrainedCharaId: 28, ScoreEventArray: []ScoreEvent{}},
			},
		},
		{
			Round:        3,
			DistanceType: Medium,
			CharaResultArray: []CharaResult{
				{TeamId: 1, TrainedCharaId: 2924, ScoreEventArray: []ScoreEvent{{Score: 27324}}},
				{TeamId: 1, TrainedCharaId: 3049, ScoreEventArray: []ScoreEvent{{Score: 42826}}},
				{TeamId: 1, TrainedCharaId: 3041, ScoreEventArray: []ScoreEvent{{Score: 49349}}},
				{TeamId: 2, TrainedCharaId: 1455, ScoreEventArray: []ScoreEvent{}},
				{TeamId: 2, TrainedCharaId: 1613, ScoreEventArray: []ScoreEvent{}},
			},
		},
		{
			Round:        4,
			DistanceType: Long,
			CharaResultArray: []CharaResult{
				{TeamId: 1, TrainedCharaId: 3002, ScoreEventArray: []ScoreEvent{{Score: 61380}}},
				{TeamId: 1, TrainedCharaId: 2939, ScoreEventArray: []ScoreEvent{{Score: 34860}}},
				{TeamId: 1, TrainedCharaId: 3060, ScoreEventArray: []ScoreEvent{{Score: 45066}}},
				{TeamId: 0, TrainedCharaId: 120, ScoreEventArray: []ScoreEvent{}},
				{TeamId: 2, TrainedCharaId: 1504, ScoreEventArray: []ScoreEvent{}},
			},
		},
		{
			Round:        5,
			DistanceType: Dirt,
			CharaResultArray: []CharaResult{
				{TeamId: 1, TrainedCharaId: 2935, ScoreEventArray: []ScoreEvent{{Score: 32500}}},
				{TeamId: 1, TrainedCharaId: 3056, ScoreEventArray: []ScoreEvent{{Score: 50178}}},
				{TeamId: 1, TrainedCharaId: 3102, ScoreEventArray: []ScoreEvent{{Score: 66996}}},
				{TeamId: 0, TrainedCharaId: 146, ScoreEventArray: []ScoreEvent{}},
				{TeamId: 2, TrainedCharaId: 1208, ScoreEventArray: []ScoreEvent{}},
			}},
	},
}

// Current umas in ttResult, in round order
var currentUmas = []int{
	2984, 2928, 2846,
	2977, 2994, 3106,
	2924, 3049, 3041,
	3002, 2939, 3060,
	2935, 3056, 3102,
}

// Distance of each round in order
var distanceOrder = []DistanceType{Mile, Sprint, Medium, Long, Dirt}

func TestAppend(t *testing.T) {
	s := TeamTrialResultSet{}
	s.append(ttResult)
	s.append(ttResult)
	got := len(s.Set)
	want := 2
	if got != want {
		t.Errorf("len(s.Set) == %d, want %d", got, want)
	}
}

// Tests if a single TT result summary outputs the expected values
func TestSummarizeSingleResult(t *testing.T) {
	s := TeamTrialResultSet{Set: []TeamTrialResult{ttResult}}.Summarize()

	// A result fields at most 15 umas, and opponents are not ours
	if !reflect.DeepEqual(s.TrainedCharaIds, currentUmas) {
		t.Errorf("s.TrainedCharaIds == %v, want %v", s.TrainedCharaIds, currentUmas)
	}
	if s.FieldedCount != len(currentUmas) {
		t.Errorf("s.FieldedCount == %d, want 15", s.FieldedCount)
	}
	for i := range s.Len() {
		// 3 results per distance, so divide by 3
		want := distanceOrder[i/3]
		if s.DistanceTypes[i] != want {
			t.Errorf("s.DistanceTypes[%d] == %v, want %v", i, s.DistanceTypes[i], want)
		}
		// IDs should be the same since it's just a diff format
		if s.CharaData[i].TrainedCharaId != s.TrainedCharaIds[i] {
			t.Errorf("s.CharaData[%d] ID == %d, want %d", i, s.CharaData[i].TrainedCharaId, s.TrainedCharaIds[i])
		}
		// Test data has only 1 score
		if s.Scores[i].Len() != 1 {
			t.Errorf("s.Scores[%d].Len() == %d, want 1", i, s.Scores[i].Len())
		}
	}
}

// Tests if a single TT result recorded twice outputs the expected values
func TestSummarizeDoubleResult(t *testing.T) {
	s := TeamTrialResultSet{Set: []TeamTrialResult{ttResult, ttResult}}.Summarize()

	// Same result twice means the second race introduced no new umas
	// so only the Scores slice should have 1 extra score
	if !reflect.DeepEqual(s.TrainedCharaIds, currentUmas) {
		t.Errorf("s.TrainedCharaIds == %v, want %v", s.TrainedCharaIds, currentUmas)
	}
	if s.FieldedCount != len(currentUmas) {
		t.Errorf("s.FieldedCount == %d, want 15", s.FieldedCount)
	}
	for i := range s.Len() {
		want := distanceOrder[i/3]
		if s.DistanceTypes[i] != want {
			t.Errorf("s.DistanceTypes[%d] == %v, want %v", i, s.DistanceTypes[i], want)
		}
		if s.CharaData[i].TrainedCharaId != s.TrainedCharaIds[i] {
			t.Errorf("s.CharaData[%d] ID == %d, want %d", i, s.CharaData[i].TrainedCharaId, s.TrainedCharaIds[i])
		}
		// Each uma got 1 score added
		if s.Scores[i].Len() != 2 {
			t.Errorf("s.Scores[%d].Len() == %d, want 2", i, s.Scores[i].Len())
		}
	}
}

// Tests if two TT results with different umas output the expected list of current umas
func TestSummarizeOrderResult(t *testing.T) {
	// Copy ttResult data so prevResult changes don't affect curResult
	prevResult := TeamTrialResult{
		RaceStartParamsArray: make([]RaceStartParams, 5),
		RaceResultArray:      make([]RaceResult, 5),
	}
	curResult := TeamTrialResult{
		RaceStartParamsArray: make([]RaceStartParams, 5),
		RaceResultArray:      make([]RaceResult, 5),
	}
	for i := range 5 {
		prevResult.RaceStartParamsArray[i] = RaceStartParams{
			Round:              ttResult.RaceStartParamsArray[i].Round,
			RaceHorseDataArray: make([]RaceHorseData, len(ttResult.RaceStartParamsArray[i].RaceHorseDataArray)),
		}
		curResult.RaceStartParamsArray[i] = RaceStartParams{
			Round:              ttResult.RaceStartParamsArray[i].Round,
			RaceHorseDataArray: make([]RaceHorseData, len(ttResult.RaceStartParamsArray[i].RaceHorseDataArray)),
		}
		prevResult.RaceResultArray[i] = RaceResult{
			Round:            ttResult.RaceResultArray[i].Round,
			DistanceType:     ttResult.RaceResultArray[i].DistanceType,
			CharaResultArray: make([]CharaResult, len(ttResult.RaceResultArray[i].CharaResultArray)),
		}
		curResult.RaceResultArray[i] = RaceResult{
			Round:            ttResult.RaceResultArray[i].Round,
			DistanceType:     ttResult.RaceResultArray[i].DistanceType,
			CharaResultArray: make([]CharaResult, len(ttResult.RaceResultArray[i].CharaResultArray)),
		}
		for j, res := range ttResult.RaceResultArray[i].CharaResultArray {
			prevResult.RaceResultArray[i].CharaResultArray[j] = CharaResult{
				TeamId:          res.TeamId,
				TrainedCharaId:  res.TrainedCharaId,
				ScoreEventArray: res.ScoreEventArray,
			}
			curResult.RaceResultArray[i].CharaResultArray[j] = CharaResult{
				TeamId:          res.TeamId,
				TrainedCharaId:  res.TrainedCharaId,
				ScoreEventArray: res.ScoreEventArray,
			}
		}
		for j, rhd := range ttResult.RaceStartParamsArray[i].RaceHorseDataArray {
			prevResult.RaceStartParamsArray[i].RaceHorseDataArray[j] = RaceHorseData{
				TeamId:         rhd.TeamId,
				TrainedCharaId: rhd.TrainedCharaId,
			}
			curResult.RaceStartParamsArray[i].RaceHorseDataArray[j] = RaceHorseData{
				TeamId:         rhd.TeamId,
				TrainedCharaId: rhd.TrainedCharaId,
			}
		}
	}

	// Copied data shouldn't alter curResult slice now
	prevResult.RaceStartParamsArray[0].RaceHorseDataArray[0].TrainedCharaId = 1234
	prevResult.RaceResultArray[0].CharaResultArray[0].TrainedCharaId = 1234
	s := TeamTrialResultSet{Set: []TeamTrialResult{prevResult, curResult}}.Summarize()

	if s.FieldedCount != len(currentUmas) {
		t.Errorf("s.FieldedCount == %d, want 15", s.FieldedCount)
	}
	for i := range s.FieldedCount {
		if s.TrainedCharaIds[i] != currentUmas[i] {
			t.Errorf("s.TrainedCharaIds[%d] == %d, want %d", i, s.TrainedCharaIds[i], currentUmas[i])
		}
	}
}

// Tests if a malformed TeamTrialResultSet summary skips parsing bad results
func TestSummarizeBadResult(t *testing.T) {
	// Malformed race result and race params lengths should skip parsing
	ttr := ttResult
	ttr.RaceResultArray = nil

	s := TeamTrialResultSet{Set: []TeamTrialResult{ttr}}.Summarize()
	if s.Len() != 0 {
		t.Errorf("s.Len() == %d, want 0", s.Len())
	}
}

// Tests if a malformed TeamTrialResultSet summary skips parsing bad results
// but keeps any good results
func TestSummarizeBadResultWithGoodResult(t *testing.T) {
	ttr := ttResult
	ttr.RaceResultArray = nil

	s := TeamTrialResultSet{Set: []TeamTrialResult{ttr, ttResult}}.Summarize()
	if s.Len() != 15 {
		t.Errorf("s.Len() == %d, want 0", s.Len())
	}
}

// Tests if data checker succeeds on good data
func TestIsValidData(t *testing.T) {
	tt := ttResult
	got := tt.IsValidData()
	want := true
	if got != want {
		t.Errorf("ttResult.IsValidData() == %t, want %t", got, want)
	}
}

// Tests if data checker fails on bad data
func TestIsValidDataNotMatchingRounds(t *testing.T) {
	tt := ttResult
	tt.RaceStartParamsArray[0].Round = 2
	tt.RaceStartParamsArray[1].Round = 1
	got := tt.IsValidData()
	want := false
	if got != want {
		t.Errorf("ttResult.IsValidData() == %t, want %t", got, want)
	}
}

// Tests if data checker fails on bad data
func TestIsValidDataNotCorrectLength(t *testing.T) {
	tt := ttResult
	tt.RaceResultArray = append(tt.RaceResultArray, ttResult.RaceResultArray[0])
	got := tt.IsValidData()
	want := false
	if got != want {
		t.Errorf("ttResult.IsValidData() == %t, want %t", got, want)
	}
}
