package affinity

import (
	"github.com/cicadaclock/umango/internal/pkg/data"
)

func SumAffinity(relationIds []int) int {
	var sum int = 0
	for _, relationId := range relationIds {
		sum += data.Relations[relationId]
	}
	return sum
}
