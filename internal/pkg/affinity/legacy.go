package affinity

import (
	"fmt"

	"github.com/cicadaclock/umango/internal/pkg/data"
)

// Uma representation of a full affinity tree
type Legacy struct {
	CharaId00 int
	CharaId10 int
	CharaId11 int
	CharaId12 int
	CharaId20 int
	CharaId21 int
	CharaId22 int
}

func (legacy Legacy) Print(dataStore *data.DataStore) {
	fmt.Println((*dataStore).CharaNames[legacy.CharaId00])
	fmt.Println("├──", (*dataStore).CharaNames[legacy.CharaId10])
	fmt.Println("│   ├──", (*dataStore).CharaNames[legacy.CharaId11])
	fmt.Println("│   └──", (*dataStore).CharaNames[legacy.CharaId12])
	fmt.Println("└──", (*dataStore).CharaNames[legacy.CharaId20])
	fmt.Println("    ├──", (*dataStore).CharaNames[legacy.CharaId21])
	fmt.Println("    └──", (*dataStore).CharaNames[legacy.CharaId22])
}

func (legacy Legacy) Affinity(dataStore *data.DataStore) int {
	sum := 0
	sum += calculateDuoAffinity(dataStore, legacy.CharaId00, legacy.CharaId10)
	sum += calculateDuoAffinity(dataStore, legacy.CharaId00, legacy.CharaId20)
	sum += calculateDuoAffinity(dataStore, legacy.CharaId10, legacy.CharaId20)
	sum += calculateTrioAffinity(dataStore, legacy.CharaId00, legacy.CharaId10, legacy.CharaId11)
	sum += calculateTrioAffinity(dataStore, legacy.CharaId00, legacy.CharaId10, legacy.CharaId12)
	sum += calculateTrioAffinity(dataStore, legacy.CharaId00, legacy.CharaId20, legacy.CharaId21)
	sum += calculateTrioAffinity(dataStore, legacy.CharaId00, legacy.CharaId20, legacy.CharaId22)
	return sum
}

func sumAffinity(dataStore *data.DataStore, relationIds []int) int {
	var sum int = 0
	for _, relationId := range relationIds {
		sum += (*dataStore).SuccessionRelations[relationId]
	}
	return sum
}

func calculateDuoAffinity(dataStore *data.DataStore, charaId1, charaId2 int) int {
	if charaId1 == charaId2 {
		return 0
	}

	relationIds_a := (*dataStore).SuccessionRelationMembers[charaId1]
	relationIds_b := (*dataStore).SuccessionRelationMembers[charaId2]
	matchedRelationIds := matchRelationIds(relationIds_a, relationIds_b)
	return sumAffinity(dataStore, matchedRelationIds)
}

func calculateTrioAffinity(dataStore *data.DataStore, charaId1, charaId2, charaId3 int) int {
	if charaId1 == charaId2 || charaId1 == charaId3 || charaId2 == charaId3 {
		return 0
	}
	relationIds_a := (*dataStore).SuccessionRelationMembers[charaId1]
	relationIds_b := (*dataStore).SuccessionRelationMembers[charaId2]
	relationIds_c := (*dataStore).SuccessionRelationMembers[charaId3]
	matchedRelationIds := matchRelationIds(relationIds_a, relationIds_b, relationIds_c)
	return sumAffinity(dataStore, matchedRelationIds)
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
