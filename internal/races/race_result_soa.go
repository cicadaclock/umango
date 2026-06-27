package races

import (
	"maps"
	"slices"
)

type RaceResultsSoA struct {
	DistanceTypes        []DistanceType
	TeamTotalScores      ScoreArray
	TeamTotalBonusScores ScoreArray
	WinTypes             []int
	CharaResultArrays    [][]CharaResult
}

type CharaResultSoA struct {
	TotalScoreSum int
	BonusScoreSum int
	TotalScore    ScoreArray
	BonusScore    ScoreArray
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
		DistanceTypes:     make([]DistanceType, 0, capacity),
		WinTypes:          make([]int, 0, capacity),
		CharaResultArrays: make([][]CharaResult, 0, capacity),
	}
}

func (soa *RaceResultsSoA) appendRace(raceResult RaceResult) {
	soa.DistanceTypes = append(soa.DistanceTypes, DistanceType(raceResult.DistanceType))
	soa.WinTypes = append(soa.WinTypes, raceResult.WinType)
	soa.CharaResultArrays = append(soa.CharaResultArrays, raceResult.CharaResultArray)
	soa.TeamTotalScores.append(raceResult.TeamTotalScore)
	soa.TeamTotalBonusScores.append(raceResult.BonusScore())
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
	return len(soa.DistanceTypes)
}

func (soa RaceResultsSoA) get(i int) RaceResult {
	return RaceResult{
		DistanceType:     int(soa.DistanceTypes[i]),
		TeamTotalScore:   soa.TeamTotalScores.Get(i),
		WinType:          soa.WinTypes[i],
		CharaResultArray: soa.CharaResultArrays[i],
	}
}

func (soa RaceResultsSoA) FilterByDistanceType(distanceType int) RaceResultsSoA {
	filtered := makeRaceResultsSoA(soa.Len())
	for i := range soa.DistanceTypes {
		if soa.DistanceTypes[i] == DistanceType(distanceType) {
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
	return soa.TeamTotalScores.Average()
}

func (soa RaceResultsSoA) BonusScoreAverage() int {
	return soa.TeamTotalBonusScores.Average()
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
