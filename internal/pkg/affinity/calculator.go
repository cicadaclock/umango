package affinity

import (
	"github.com/cicadaclock/umango/internal/pkg/data"
	"github.com/cicadaclock/umango/internal/pkg/structs"
)

func SumAffinity(relationIds []int) int {
	var sum int = 0
	for _, relationId := range relationIds {
		sum += data.Relations[relationId]
	}
	return sum
}

func CalculateDuoAffinity(a, b structs.Uma) int {
	if a.Id == b.Id {
		return 0
	}

	relationIds_a := RelationIds(a)
	relationIds_b := RelationIds(b)
	matchedRelationIds := MatchRelationIds(relationIds_a, relationIds_b)
	return SumAffinity(matchedRelationIds)
}

func CalculateTrioAffinity(a, b, c structs.Uma) int {
	if a.Id == b.Id || a.Id == c.Id || b.Id == c.Id {
		return 0
	}
	relationIds_a := RelationIds(a)
	relationIds_b := RelationIds(b)
	relationIds_c := RelationIds(c)
	matchedRelationIds := MatchRelationIds(relationIds_a, relationIds_b, relationIds_c)
	return SumAffinity(matchedRelationIds)
}
