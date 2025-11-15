package affinity

import (
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/structs"
)

// Return a list of relation IDs for the selected Uma
func RelationIds(uma structs.Uma) []int {
	relationIds := make([]int, 0, 1000)
	for _, relationMember := range data.RelationMembers {
		if relationMember.UmaId == uma.Id {
			relationIds = append(relationIds, relationMember.RelationId)
		}
	}
	return relationIds
}

func MatchRelationIds(relationIdsArgs ...[]int) []int {
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
