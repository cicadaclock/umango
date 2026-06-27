package races

type TableData struct {
	TrainedCharaIds []int
	MaxScores       []int
	AvgScores       []int
}

func NewTableData(ttrs TeamTrialResultSet) TableData {
	result := TableData{}
	scores := make(map[int]*ScoreArray)
	for _, ttr := range ttrs.Set {
		charaResults := ttr.GetMyCharaResults()
		for _, result := range charaResults {
			scoreArray := scores[result.TrainedCharaId]
			scoreArray.append(result.TotalScore())
		}
	}

	trainedCharaIds := make([]int, 0, len(scores))
	maxScores := make([]int, 0, len(scores))
	avgScores := make([]int, 0, len(scores))
	for trainedCharaId, scoreArray := range scores {
		trainedCharaIds = append(trainedCharaIds, trainedCharaId)
		avgScores = append(avgScores, scoreArray.Average())
		maxScores = append(maxScores, scoreArray.Max())
	}
	return result
}
