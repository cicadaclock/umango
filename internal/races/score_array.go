package races

import "slices"

type ScoreArray struct {
	Sum   int
	Score []int
}

func (s *ScoreArray) append(i int) {
	s.Score = append(s.Score, i)
	s.Sum += i
}

func (s ScoreArray) Average() int {
	if s.Len() == 0 {
		return 0
	}
	return s.Sum / s.Len()
}

func (s ScoreArray) Len() int {
	return len(s.Score)
}

func (s ScoreArray) Get(i int) int {
	return s.Score[i]
}

func (s ScoreArray) Max() int {
	return slices.Max(s.Score)
}

// Filter selects only the elements that match the provided indices
func (s ScoreArray) Filter(indices []int) ScoreArray {
	filtered := ScoreArray{Score: make([]int, 0, len(indices))}
	for _, i := range indices {
		if i >= 0 && i < len(s.Score) {
			filtered.append(s.Score[i])
		}
	}
	return filtered
}
