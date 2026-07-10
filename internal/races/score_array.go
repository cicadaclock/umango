package races

import (
	"math"
	"slices"
)

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

func (s ScoreArray) Min() int {
	return slices.Min(s.Score)
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

// HistogramCoords returns a pair of ([]x, []y) allocated into buckets
func (s ScoreArray) HistogramCoords(stepSize int) ([]int, []int) {
	counts := make(map[int]int)
	for _, score := range s.Score {
		counts[getScoreBucketIndex(score, stepSize)]++
	}

	xPts := make([]int, 0, len(counts))
	yPts := make([]int, 0, len(counts))
	for i, b := range counts {
		xPts = append(xPts, i*stepSize)
		yPts = append(yPts, b)
	}
	return xPts, yPts
}

func getScoreBucketIndex(score, stepSize int) int {
	return int(math.Ceil(float64(score) / float64(stepSize)))
}
