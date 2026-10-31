package veteran

import (
	"fmt"
)

// AffinityMapper provides the master.mdb data that the affinity calculation needs
type AffinityMapper interface {
	// Chara ID of a card ID
	CardChara(cardId int) int
	// Name of a chara ID
	CharaName(charaId int) string
	// Succession relation types that include a chara ID
	RelationMembers(charaId int) []int
	// Affinity points of a succession relation type
	RelationPoint(relationType int) int
}

// Member is one uma of a legacy tree
type Member struct {
	CharaId int
	// Win saddle IDs in ascending order
	WinSaddleIds []int
}

// Parent is one legacy parent of the trainee, with its own two parents
type Parent struct {
	Member
	GrandParent1 Member
	GrandParent2 Member
}

// Legacy is the legacy tree of a trainee: two parents and their four parents
type Legacy struct {
	TraineeCharaId int
	Parent1        Parent
	Parent2        Parent
}

// NewLegacy builds the legacy tree of a trainee from two veterans
func NewLegacy(traineeCharaId int, parent1, parent2 Veteran, dataStore AffinityMapper) Legacy {
	return Legacy{
		TraineeCharaId: traineeCharaId,
		Parent1:        NewParent(parent1, dataStore),
		Parent2:        NewParent(parent2, dataStore),
	}
}

// NewParent builds a legacy parent from a veteran and its succession charas
func NewParent(v Veteran, dataStore AffinityMapper) Parent {
	parent := Parent{Member: newMember(v.CardId, v.WinSaddleIdArray, dataStore)}
	for _, chara := range v.SuccessionCharaArray {
		// Positions 10 and 20 are the veteran's own two parents
		switch chara.PositionId {
		case 10:
			parent.GrandParent1 = newMember(chara.CardId, chara.WinSaddleIdArray, dataStore)
		case 20:
			parent.GrandParent2 = newMember(chara.CardId, chara.WinSaddleIdArray, dataStore)
		}
	}
	return parent
}

// newMember builds a legacy member from its card ID and win saddle IDs
func newMember(cardId int, winSaddleIds []int, dataStore AffinityMapper) Member {
	return Member{
		CharaId:      dataStore.CardChara(cardId),
		WinSaddleIds: winSaddleIds,
	}
}

// Print writes the legacy tree to standard output for diagnosis
func (legacy Legacy) Print(dataStore AffinityMapper) {
	fmt.Println(dataStore.CharaName(legacy.TraineeCharaId))
	fmt.Println("├──", dataStore.CharaName(legacy.Parent1.CharaId))
	fmt.Println("│   ├──", dataStore.CharaName(legacy.Parent1.GrandParent1.CharaId))
	fmt.Println("│   └──", dataStore.CharaName(legacy.Parent1.GrandParent2.CharaId))
	fmt.Println("└──", dataStore.CharaName(legacy.Parent2.CharaId))
	fmt.Println("    ├──", dataStore.CharaName(legacy.Parent2.GrandParent1.CharaId))
	fmt.Println("    └──", dataStore.CharaName(legacy.Parent2.GrandParent2.CharaId))
}

// Affinity returns the total affinity of the legacy tree: base affinity plus race affinity
func (legacy Legacy) Affinity(dataStore AffinityMapper) int {
	return legacy.BaseAffinity(dataStore) + legacy.RaceAffinity()
}

// BaseAffinity returns the affinity from the succession relations within the legacy tree
func (legacy Legacy) BaseAffinity(dataStore AffinityMapper) int {
	return legacy.Parent1.baseAffinity(dataStore, legacy.TraineeCharaId) +
		legacy.Parent2.baseAffinity(dataStore, legacy.TraineeCharaId) +
		relationAffinity(dataStore, legacy.Parent1.CharaId, legacy.Parent2.CharaId)
}

// baseAffinity returns the relation affinity of the trainee with this parent, and of the
// trainee and this parent with each grandparent
func (parent Parent) baseAffinity(dataStore AffinityMapper, traineeCharaId int) int {
	return relationAffinity(dataStore, traineeCharaId, parent.CharaId) +
		relationAffinity(dataStore, traineeCharaId, parent.CharaId, parent.GrandParent1.CharaId) +
		relationAffinity(dataStore, traineeCharaId, parent.CharaId, parent.GrandParent2.CharaId)
}

// RaceAffinity returns the affinity from the win saddles that each parent shares with its parents
func (legacy Legacy) RaceAffinity() int {
	return legacy.Parent1.raceAffinity() + legacy.Parent2.raceAffinity()
}

// raceAffinity returns the win saddles that this parent shares with each of its parents
func (parent Parent) raceAffinity() int {
	return calculateRaceAffinity(parent.WinSaddleIds, parent.GrandParent1.WinSaddleIds, parent.GrandParent2.WinSaddleIds)
}

// relationAffinity returns the affinity points of every succession relation that includes all
// of charaIds, or 0 when a chara appears twice
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

// sumRelationAffinity returns the total affinity points of the succession relation types
func sumRelationAffinity(dataStore AffinityMapper, relationIds []int) int {
	var sum int = 0
	for _, relationId := range relationIds {
		sum += dataStore.RelationPoint(relationId)
	}
	return sum
}

// matchRelationIds returns the relation types present in every list. Each list sets its own
// bit, so a type present in every list has every bit set.
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

// calculateRaceAffinity counts the win saddles of winSaddle1 that also appear in winSaddle2
// and in winSaddle3. Every list must be in ascending order.
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
