package veteran

import (
	"fmt"
)

// AffinityMapper provides the functions this package needs from DataStore
type AffinityMapper interface {
	// CardChara maps a card ID to its chara ID
	CardChara(cardId int) int
	// CharaName maps a chara ID to its chara name
	CharaName(charaId int) string
	// RelationMembers maps a chara ID to its succession relation types
	RelationMembers(charaId int) []int
	// RelationPoint maps a succession relation type to its relation points
	RelationPoint(relationType int) int
}

// Single uma in the legacy tree
type Member struct {
	// Chara ID of this uma
	CharaId int
	// Race IDs this uma placed 1st in
	WinSaddleIds []int
}

// Single veteran with its own two parents (trainee's grandparents)
type Parent struct {
	Member
	GrandParent1 Member
	GrandParent2 Member
}

// Uma representation of a full affinity tree:
// trainee plus two inherited parents
type Legacy struct {
	TraineeCharaId int
	Parent1        Parent
	Parent2        Parent
}

func NewLegacy(traineeCharaId int, parent1, parent2 Veteran, dataStore AffinityMapper) Legacy {
	return Legacy{
		TraineeCharaId: traineeCharaId,
		Parent1:        NewParent(parent1, dataStore),
		Parent2:        NewParent(parent2, dataStore),
	}
}

func NewParent(v Veteran, dataStore AffinityMapper) Parent {
	parent := Parent{Member: newMember(v.CardId, v.WinSaddleIdArray, dataStore)}
	for _, chara := range v.SuccessionCharaArray {
		switch chara.PositionId {
		case 10:
			parent.GrandParent1 = newMember(chara.CardId, chara.WinSaddleIdArray, dataStore)
		case 20:
			parent.GrandParent2 = newMember(chara.CardId, chara.WinSaddleIdArray, dataStore)
		}
	}
	return parent
}

func newMember(cardId int, winSaddleIds []int, dataStore AffinityMapper) Member {
	return Member{
		CharaId:      dataStore.CardChara(cardId),
		WinSaddleIds: winSaddleIds,
	}
}

func (legacy Legacy) Print(dataStore AffinityMapper) {
	fmt.Println(dataStore.CharaName(legacy.TraineeCharaId))
	fmt.Println("├──", dataStore.CharaName(legacy.Parent1.CharaId))
	fmt.Println("│   ├──", dataStore.CharaName(legacy.Parent1.GrandParent1.CharaId))
	fmt.Println("│   └──", dataStore.CharaName(legacy.Parent1.GrandParent2.CharaId))
	fmt.Println("└──", dataStore.CharaName(legacy.Parent2.CharaId))
	fmt.Println("    ├──", dataStore.CharaName(legacy.Parent2.GrandParent1.CharaId))
	fmt.Println("    └──", dataStore.CharaName(legacy.Parent2.GrandParent2.CharaId))
}

// Total affinity score: base affinity + race affinity
func (legacy Legacy) Affinity(dataStore AffinityMapper) int {
	return legacy.BaseAffinity(dataStore) + legacy.RaceAffinity()
}

// Affinity from base umas
func (legacy Legacy) BaseAffinity(dataStore AffinityMapper) int {
	return legacy.Parent1.baseAffinity(dataStore, legacy.TraineeCharaId) +
		legacy.Parent2.baseAffinity(dataStore, legacy.TraineeCharaId) +
		relationAffinity(dataStore, legacy.Parent1.CharaId, legacy.Parent2.CharaId)
}

func (parent Parent) baseAffinity(dataStore AffinityMapper, traineeCharaId int) int {
	return relationAffinity(dataStore, traineeCharaId, parent.CharaId) +
		relationAffinity(dataStore, traineeCharaId, parent.CharaId, parent.GrandParent1.CharaId) +
		relationAffinity(dataStore, traineeCharaId, parent.CharaId, parent.GrandParent2.CharaId)
}

// Affinity from races won by both parents and their grandparents
func (legacy Legacy) RaceAffinity() int {
	return legacy.Parent1.raceAffinity() + legacy.Parent2.raceAffinity()
}

func (parent Parent) raceAffinity() int {
	return calculateRaceAffinity(parent.WinSaddleIds, parent.GrandParent1.WinSaddleIds, parent.GrandParent2.WinSaddleIds)
}

// Sums the affinity of every succession relation shared by all of the given
// charas
func relationAffinity(dataStore AffinityMapper, charaIds ...int) int {
	relationIdsPerChara := make([][]int, len(charaIds))
	for i, charaId := range charaIds {
		for _, prev := range charaIds[:i] {
			if charaId == prev {
				return 0
			}
		}
		relationIdsPerChara[i] = dataStore.RelationMembers(charaId)
	}
	return sumRelationAffinity(dataStore, matchRelationIds(relationIdsPerChara...))
}

func sumRelationAffinity(dataStore AffinityMapper, relationIds []int) int {
	var sum int = 0
	for _, relationId := range relationIds {
		sum += dataStore.RelationPoint(relationId)
	}
	return sum
}

func matchRelationIds(relationIdsArgs ...[]int) []int {
	match := make(map[int]int)
	matchNum := (1 << len(relationIdsArgs)) - 1

	for i, relationIds := range relationIdsArgs {
		for _, relationId := range relationIds {
			match[relationId] += 1 << i
		}
	}
	relationIds := make([]int, 0, len(match))
	for relationId, val := range match {
		if val == matchNum {
			relationIds = append(relationIds, relationId)
		}
	}
	return relationIds
}

// Counts races shared between winSaddle1 and each other list
// WinSaddle must be sorted in ascending order
func calculateRaceAffinity(winSaddle1, winSaddle2, winSaddle3 []int) int {
	// TODO: Maybe assert winSaddle in ascending order?
	i := 0
	j := 0
	sum := 0
	for _, id := range winSaddle1 {
		for i < len(winSaddle2) {
			if winSaddle2[i] == id {
				sum++
				i++
			} else if winSaddle2[i] < id {
				i++
				continue
			} else if winSaddle2[i] > id {
				break
			}
		}
		for j < len(winSaddle3) {
			if winSaddle3[j] == id {
				sum++
				j++
			} else if winSaddle3[j] < id {
				j++
				continue
			} else if winSaddle3[j] > id {
				break
			}
		}
	}
	return sum
}
