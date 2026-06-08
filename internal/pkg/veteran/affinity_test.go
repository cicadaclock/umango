package veteran

import (
	"testing"

	"github.com/cicadaclock/umango/internal/pkg/data"
)

func setup(t *testing.T) *data.DataStore {
	dataStore, err := data.Init()
	if err != nil {
		t.Errorf("init dataStore: %v", err)
	}
	return dataStore
}

func TestLegacyAffinityForFullLegacy(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{
		CharaId00: 1001,
		CharaId10: 1002,
		CharaId20: 1003,
		CharaId11: 1003,
		CharaId12: 1005,
		CharaId21: 1008,
		CharaId22: 1015,
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 125 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestLegacyAffinityForPartialLegacy(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{
		CharaId00: 1001,
		CharaId10: 1002,
		CharaId20: 1003,
		CharaId11: 1003,
		CharaId12: 1005,
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 89 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestLegacyAffinityForEmptyLegacy(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{}
	affinity := legacy.Affinity(dataStore)
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestLegacyAffinityForSameUmaInParent(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{
		CharaId00: 1001,
		CharaId10: 1001,
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestLegacyAffinityForEmptyParent(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{
		CharaId00: 1001,
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestRaceAffinity(t *testing.T) {
	parent := []int{1, 2, 5, 10, 11, 12, 13, 15, 16, 17, 18, 23, 25, 26, 27, 34, 63, 145, 146, 147}
	grandparent1 := []int{4, 5, 6, 10, 13, 14, 15, 17, 23, 26, 27, 61, 122, 130}
	grandparent2 := []int{2, 6, 7, 10, 11, 14, 15, 17, 18, 21, 23, 25, 26, 29, 32, 34, 35, 39, 65, 85}
	affinity := calculateRaceAffinity(parent, grandparent1, grandparent2)
	if affinity != 18 {
		t.Errorf("Affinity == %d, want 18", affinity)
	}
}
