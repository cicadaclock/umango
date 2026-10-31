package races

import (
	"reflect"
	"testing"
)

// TestScoreArrayAppend checks that append adds a score
func TestScoreArrayAppend(t *testing.T) {
	s := ScoreArray{}

	s.append(10)
	if !reflect.DeepEqual(s.Score, []int{10}) {
		t.Errorf("s.Score == %v, want [10]", s.Score)
	}
}

// TestScoreArrayAverage checks the sum and mean of the scores
func TestScoreArrayAverage(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	if s.Sum != 30 {
		t.Errorf("s.Sum == %d, want 30", s.Sum)
	}
	if s.Average() != 15 {
		t.Errorf("s.Average() == %d, want 15", s.Average())
	}
}

// TestScoreArrayLen checks the number of scores
func TestScoreArrayLen(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	s.append(30)
	if s.Len() != 3 {
		t.Errorf("s.Len() == %d, want 3", s.Len())
	}
}

// TestScoreArrayGet checks the score at an index
func TestScoreArrayGet(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	s.append(30)
	if s.Get(1) != 20 {
		t.Errorf("s.Get(1) == %d, want 20", s.Get(1))
	}
}

// TestScoreArrayMax checks the highest score
func TestScoreArrayMax(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	s.append(30)
	if s.Max() != 30 {
		t.Errorf("s.Max() == %d, want 30", s.Max())
	}
}

// TestScoreArrayMin checks the lowest score
func TestScoreArrayMin(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	s.append(30)
	if s.Min() != 10 {
		t.Errorf("s.Min() == %d, want 10", s.Min())
	}
}

// TestScoreArrayFilter checks that Filter keeps the scores at the given indices
func TestScoreArrayFilter(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	s.append(30)
	s.append(40)
	filteredScore := s.Filter([]int{1, 2})
	if !reflect.DeepEqual(filteredScore.Score, []int{20, 30}) {
		t.Errorf("filteredScore == %v, want [20 30]", filteredScore)
	}
}

// TestScoreBucket checks the bucket index of scores at and around bucket bounds
func TestScoreBucket(t *testing.T) {
	scores := []int{3000, 5000, 44999, 45000, 45001}
	want := []int{0, 1, 8, 9, 9}
	for i, score := range scores {
		got := getScoreBucketIndex(score, 5000)
		if got != want[i] {
			t.Errorf("getScoreBucketIndex(%[1]d, %[1]d) == %[2]d, want %[3]d", i, got, want)
		}
	}
}
