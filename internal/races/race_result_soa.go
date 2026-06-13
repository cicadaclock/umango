package races

import (
	"maps"
	"slices"
)

type RaceResultsSoA struct {
	TeamTotalScoreSum      int
	TeamTotalBonusScoreSum int
	DistanceTypes          []int
	TeamTotalScores        []int
	TeamTotalBonusScores   []int
	WinTypes               []int
	CharaResultArrays      [][]CharaResult
}

type CharaResultSoA struct {
	TotalScoreSum int
	BonusScoreSum int
	// TotalScore    []int
	// BonusScore    []int
	TotalScore ScoreArray
	BonusScore ScoreArray
}

type ScoreArray struct {
	Sum   int
	Score []int
}

func NewRaceResultsSoA(raceResults []RaceResult) RaceResultsSoA {
	soa := makeRaceResultsSoA(len(raceResults))
	for _, raceResult := range raceResults {
		soa.appendRace(raceResult)
	}
	return soa
}

func makeRaceResultsSoA(capacity int) RaceResultsSoA {
	return RaceResultsSoA{
		DistanceTypes:     make([]int, 0, capacity),
		TeamTotalScores:   make([]int, 0, capacity),
		WinTypes:          make([]int, 0, capacity),
		CharaResultArrays: make([][]CharaResult, 0, capacity),
	}
}

func (soa *RaceResultsSoA) appendRace(raceResult RaceResult) {
	soa.DistanceTypes = append(soa.DistanceTypes, raceResult.DistanceType)
	soa.TeamTotalScores = append(soa.TeamTotalScores, raceResult.TeamTotalScore)
	soa.TeamTotalScoreSum += raceResult.TeamTotalScore
	bonusScore := raceResult.BonusScore()
	soa.TeamTotalBonusScores = append(soa.TeamTotalBonusScores, bonusScore)
	soa.TeamTotalBonusScoreSum += bonusScore
	soa.WinTypes = append(soa.WinTypes, raceResult.WinType)
	soa.CharaResultArrays = append(soa.CharaResultArrays, raceResult.CharaResultArray)
}

// Returns the number of unique player trainedCharaIds in the race result set
func (soa RaceResultsSoA) UniqueCharas() []int {
	trainedCharaIDs := make(map[int]bool, 15)
	for _, charaResults := range soa.CharaResultArrays {
		for _, charaResult := range charaResults {
			if len(charaResult.ScoreEventArray) == 0 {
				continue
			}
			trainedCharaIDs[charaResult.TrainedCharaId] = true
		}
	}
	return slices.Collect(maps.Keys(trainedCharaIDs))
}

// Maps trainedCharaIds to their race result set
func (soa RaceResultsSoA) CharaResultSoA() map[int]*CharaResultSoA {
	mapCharaResultSoA := make(map[int]*CharaResultSoA, 15)
	for _, charaResults := range soa.CharaResultArrays {
		for _, charaResult := range charaResults {
			if len(charaResult.ScoreEventArray) == 0 {
				continue
			}
			soa := mapCharaResultSoA[charaResult.TrainedCharaId]
			if soa == nil {
				soa = &CharaResultSoA{}
				mapCharaResultSoA[charaResult.TrainedCharaId] = soa
			}
			soa.append(charaResult)
		}
	}
	return mapCharaResultSoA
}

func (soa RaceResultsSoA) Len() int {
	return len(soa.TeamTotalScores)
}

func (soa RaceResultsSoA) get(index int) RaceResult {
	return RaceResult{
		DistanceType:     soa.DistanceTypes[index],
		TeamTotalScore:   soa.TeamTotalScores[index],
		WinType:          soa.WinTypes[index],
		CharaResultArray: soa.CharaResultArrays[index],
	}
}

func (soa RaceResultsSoA) FilterByDistanceType(distanceType int) RaceResultsSoA {
	filtered := makeRaceResultsSoA(soa.Len())
	for i := range soa.DistanceTypes {
		if soa.DistanceTypes[i] == distanceType {
			filtered.appendRace(soa.get(i))
		}
	}
	return filtered
}

func (soa RaceResultsSoA) FilterByTrainedCharaId(trainedCharaId int) RaceResultsSoA {
	filtered := makeRaceResultsSoA(soa.Len())
	for i, charaResults := range soa.CharaResultArrays {
		for _, charaResult := range charaResults {
			if charaResult.TrainedCharaId == trainedCharaId {
				race := soa.get(i)
				race.CharaResultArray = []CharaResult{charaResult}
				filtered.appendRace(race)
				break
			}
		}
	}
	return filtered
}

func (soa RaceResultsSoA) CharaTotalScores() []int {
	totalScores := make([]int, soa.Len())
	for i, charaResults := range soa.CharaResultArrays {
		for _, charaResult := range charaResults {
			totalScores[i] += charaResult.TotalScore()
		}
	}
	return totalScores
}

func (soa RaceResultsSoA) TotalScoreAverage() int {
	return soa.TeamTotalScoreSum / soa.Len()
}

func (soa RaceResultsSoA) BonusScoreAverage() int {
	return soa.TeamTotalBonusScoreSum / soa.Len()
}

func (soa *CharaResultSoA) append(charaResult CharaResult) {
	totalScore := charaResult.TotalScore()
	bonusScore := charaResult.BonusScore()
	soa.TotalScore.append(totalScore)
	soa.BonusScore.append(bonusScore)
}

func (soa CharaResultSoA) TotalScoreAverage() int {
	return soa.TotalScore.Average()
}

func (s *ScoreArray) append(i int) {
	s.Score = append(s.Score, i)
	s.Sum += i
}

func (s ScoreArray) Average() int {
	return s.Sum / s.Len()
}

func (s ScoreArray) Len() int {
	return len(s.Score)
}

// Filter selects only the elements that match the provided indices
func (s ScoreArray) Filter(indices []int) ScoreArray {
	filtered := ScoreArray{}
	for _, i := range indices {
		if i >= 0 && i < len(s.Score) {
			filtered.append(s.Score[i])
		}
	}
	return filtered
}
