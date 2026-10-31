package races

import (
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

// HistogramCoords counts the scores in buckets stepSize wide
//
// Returns:
// - xPts: Lower bound of each nonempty bucket, ascending; nil when there are no scores
// - yPts: Number of scores in the bucket at the same position of xPts
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

func getScoreBucketIndex(score, stepSize int) int {
	return score / stepSize
}
