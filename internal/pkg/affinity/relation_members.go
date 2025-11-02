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

// Returns overlapping relation IDs for two ID lists
func MatchRelationIds(relationId_a, relationId_b []int) []int {
	match := make(map[int]int)
	for _, relationId := range relationId_a {
		match[relationId] += 1
	}
	for _, relationId := range relationId_b {
		match[relationId] += 2
	}
	relationIds := make([]int, 0, len(match))
	for relationId, val := range match {
		if val == 3 {
			relationIds = append(relationIds, relationId)
		}
	}
	return relationIds
}
