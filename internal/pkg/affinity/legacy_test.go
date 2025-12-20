package affinity

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
