package affinity

import (
	"github.com/cicadaclock/umango/internal/pkg/structs"
)

// Uma representation of a full affinity tree
type Legacy struct {
	Trainee    structs.Uma
	Parent_1   structs.Uma
	Parent_1_1 structs.Uma
	Parent_1_2 structs.Uma
	Parent_2   structs.Uma
	Parent_2_1 structs.Uma
	Parent_2_2 structs.Uma
}

func (legacy Legacy) Affinity() int {
	a := CalculateDuoAffinity(legacy.Trainee, legacy.Parent_1)
	b := CalculateDuoAffinity(legacy.Trainee, legacy.Parent_2)
	c := CalculateDuoAffinity(legacy.Parent_1, legacy.Parent_2)

	d := CalculateTrioAffinity(legacy.Trainee, legacy.Parent_1, legacy.Parent_1_1)
	e := CalculateTrioAffinity(legacy.Trainee, legacy.Parent_1, legacy.Parent_1_2)

	f := CalculateTrioAffinity(legacy.Trainee, legacy.Parent_2, legacy.Parent_2_1)
	g := CalculateTrioAffinity(legacy.Trainee, legacy.Parent_2, legacy.Parent_2_2)

	return a + b + c + d + e + f + g
}
