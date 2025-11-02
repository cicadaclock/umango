package calculator

import (
	"github.com/cicadaclock/umango/internal/pkg/affinity"
	"github.com/cicadaclock/umango/internal/pkg/structs"
)

func CalculateDuoAffinity(a, b structs.Uma) int {
	if a.Id == b.Id {
		return 0
	}

	relationIds_a := affinity.RelationIds(a)
	relationIds_b := affinity.RelationIds(b)
	matchedRelationIds := affinity.MatchRelationIds(relationIds_a, relationIds_b)
	return affinity.SumAffinity(matchedRelationIds)
}
