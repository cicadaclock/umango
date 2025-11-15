package affinity

import (
	"testing"

	"github.com/cicadaclock/umango/internal/pkg/structs"
)

func TestLegacyAffinityForFullLegacy(t *testing.T) {
	uma_1 := structs.Uma{
		Id:   1001,
		Name: "Special Week",
	}
	uma_2 := structs.Uma{
		Id:   1002,
		Name: "Silence Suzuka",
	}
	uma_3 := structs.Uma{
		Id:   1003,
		Name: "Tokai Teio",
	}
	uma_4 := structs.Uma{
		Id:   1003,
		Name: "Tokai Teio",
	}
	uma_5 := structs.Uma{
		Id:   1005,
		Name: "Fuji Kiseki",
	}
	uma_6 := structs.Uma{
		Id:   1008,
		Name: "Vodka",
	}
	uma_7 := structs.Uma{
		Id:   1015,
		Name: "TM Opera O",
	}

	legacy := Legacy{
		Trainee:    uma_1,
		Parent_1:   uma_2,
		Parent_2:   uma_3,
		Parent_1_1: uma_4,
		Parent_1_2: uma_5,
		Parent_2_1: uma_6,
		Parent_2_2: uma_7,
	}

	affinity := legacy.Affinity()
	if affinity != 125 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestLegacyAffinityForPartialLegacy(t *testing.T) {
	uma_1 := structs.Uma{
		Id:   1001,
		Name: "Special Week",
	}
	uma_2 := structs.Uma{
		Id:   1002,
		Name: "Silence Suzuka",
	}
	uma_3 := structs.Uma{
		Id:   1003,
		Name: "Tokai Teio",
	}
	uma_4 := structs.Uma{
		Id:   1003,
		Name: "Tokai Teio",
	}
	uma_5 := structs.Uma{
		Id:   1005,
		Name: "Fuji Kiseki",
	}

	legacy := Legacy{
		Trainee:    uma_1,
		Parent_1:   uma_2,
		Parent_2:   uma_3,
		Parent_1_1: uma_4,
		Parent_1_2: uma_5,
	}

	affinity := legacy.Affinity()
	if affinity != 89 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestLegacyAffinityForEmptyLegacy(t *testing.T) {
	legacy := Legacy{}

	affinity := legacy.Affinity()
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestLegacyAffinityForSameUmaInParent(t *testing.T) {
	uma_1 := structs.Uma{
		Id:   1001,
		Name: "Special Week",
	}
	uma_2 := structs.Uma{
		Id:   1001,
		Name: "Special Week",
	}

	legacy := Legacy{
		Trainee:  uma_1,
		Parent_1: uma_2,
	}

	affinity := legacy.Affinity()
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestLegacyAffinityForEmptyParent(t *testing.T) {
	uma_1 := structs.Uma{
		Id:   1001,
		Name: "Special Week",
	}

	legacy := Legacy{
		Trainee: uma_1,
	}

	affinity := legacy.Affinity()
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}
