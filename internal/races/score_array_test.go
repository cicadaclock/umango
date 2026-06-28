package races

import (
	"reflect"
	"testing"
)

func TestScoreArrayAppend(t *testing.T) {
	s := ScoreArray{}

	s.append(10)
	if !reflect.DeepEqual(s.Score, []int{10}) {
		t.Errorf("s.Score == %v, want [10]", s.Score)
	}
}

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

func TestScoreArrayLen(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	s.append(30)
	if s.Len() != 3 {
		t.Errorf("s.Len() == %d, want 3", s.Len())
	}
}

func TestScoreArrayGet(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	s.append(30)
	if s.Get(1) != 20 {
		t.Errorf("s.Get(1) == %d, want 20", s.Get(1))
	}
}

func TestScoreArrayMax(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	s.append(30)
	if s.Max() != 30 {
		t.Errorf("s.Max() == %d, want 30", s.Max())
	}
}

func TestScoreArrayMin(t *testing.T) {
	s := ScoreArray{}
	s.append(10)
	s.append(20)
	s.append(30)
	if s.Min() != 10 {
		t.Errorf("s.Min() == %d, want 10", s.Min())
	}
}

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
