package races

import (
	"slices"
)

// ScoreArray holds the total score of each race of one uma, with their sum
type ScoreArray struct {
	Sum   int
	Score []int
}

// append adds the score of one race
func (s *ScoreArray) append(i int) {
	s.Score = append(s.Score, i)
	s.Sum += i
}

// Average returns the mean score, or 0 when there are no scores
func (s ScoreArray) Average() int {
	if s.Len() == 0 {
		return 0
	}
	return s.Sum / s.Len()
}

// Len returns the number of scores
func (s ScoreArray) Len() int {
	return len(s.Score)
}

// Get returns score i
func (s ScoreArray) Get(i int) int {
	return s.Score[i]
}

// Max returns the highest score; there must be at least one score
func (s ScoreArray) Max() int {
	return slices.Max(s.Score)
}

// Min returns the lowest score; there must be at least one score
func (s ScoreArray) Min() int {
	return slices.Min(s.Score)
}

// Filter returns the scores at the given indices, skipping indices out of range
func (s ScoreArray) Filter(indices []int) ScoreArray {
	filtered := ScoreArray{Score: make([]int, 0, len(indices))}
	for _, i := range indices {
		if i >= 0 && i < len(s.Score) {
			filtered.append(s.Score[i])
		}
	}
	return filtered
}

// HistogramCoords counts the scores in buckets stepSize wide and returns the lower bound and
// count of each nonempty bucket, in ascending order
func (s ScoreArray) HistogramCoords(stepSize int) ([]int, []int) {
	if s.Len() == 0 {
		return nil, nil
	}
	first := getScoreBucketIndex(s.Min(), stepSize)
	counts := make([]int, getScoreBucketIndex(s.Max(), stepSize)-first+1)
	for _, score := range s.Score {
		counts[getScoreBucketIndex(score, stepSize)-first]++
	}

	xPts := make([]int, 0, len(counts))
	yPts := make([]int, 0, len(counts))
	for i, count := range counts {
		if count > 0 {
			xPts = append(xPts, (first+i)*stepSize)
			yPts = append(yPts, count)
		}
	}
	return xPts, yPts
}

// getScoreBucketIndex returns the index of the bucket stepSize wide that holds score
func getScoreBucketIndex(score, stepSize int) int {
	return score / stepSize
}
