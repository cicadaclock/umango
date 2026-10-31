package veteran

import (
	"testing"

	"github.com/cicadaclock/umango/internal/data"
)

// setup loads master.mdb from the local game install, which these tests require
func setup(t *testing.T) *data.DataStore {
	dataStore, err := data.Init()
	if err != nil {
		t.Errorf("init dataStore: %v", err)
	}
	return dataStore
}

// TestLegacyAffinityForFullLegacy checks the affinity of a legacy tree with every member
func TestLegacyAffinityForFullLegacy(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{
		TraineeCharaId: 1001,
		Parent1: Parent{
			Member:       Member{CharaId: 1002},
			GrandParent1: Member{CharaId: 1003},
			GrandParent2: Member{CharaId: 1005},
		},
		Parent2: Parent{
			Member:       Member{CharaId: 1003},
			GrandParent1: Member{CharaId: 1008},
			GrandParent2: Member{CharaId: 1015},
		},
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 144 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

// TestLegacyAffinityForPartialLegacy checks the affinity of a legacy tree with one parent
// missing its parents
func TestLegacyAffinityForPartialLegacy(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{
		TraineeCharaId: 1001,
		Parent1: Parent{
			Member:       Member{CharaId: 1002},
			GrandParent1: Member{CharaId: 1003},
			GrandParent2: Member{CharaId: 1005},
		},
		Parent2: Parent{
			Member: Member{CharaId: 1003},
		},
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 102 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

// TestLegacyAffinityForEmptyLegacy checks that an empty legacy tree has no affinity
func TestLegacyAffinityForEmptyLegacy(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{}
	affinity := legacy.Affinity(dataStore)
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

// TestLegacyAffinityForSameUmaInParent checks that a parent of the same chara as the trainee
// has no affinity
func TestLegacyAffinityForSameUmaInParent(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{
		TraineeCharaId: 1001,
		Parent1: Parent{
			Member: Member{CharaId: 1001},
		},
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

// TestLegacyAffinityForEmptyParent checks that a trainee without parents has no affinity
func TestLegacyAffinityForEmptyParent(t *testing.T) {
	dataStore := setup(t)

	legacy := Legacy{
		TraineeCharaId: 1001,
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

// TestRaceAffinity checks the count of win saddles that a parent shares with each of its parents
func TestRaceAffinity(t *testing.T) {
	parent := []int{1, 2, 5, 10, 11, 12, 13, 15, 16, 17, 18, 23, 25, 26, 27, 34, 63, 145, 146, 147}
	grandparent1 := []int{4, 5, 6, 10, 13, 14, 15, 17, 23, 26, 27, 61, 122, 130}
	grandparent2 := []int{2, 6, 7, 10, 11, 14, 15, 17, 18, 21, 23, 25, 26, 29, 32, 34, 35, 39, 65, 85}
	affinity := calculateRaceAffinity(parent, grandparent1, grandparent2)
	if affinity != 18 {
		t.Errorf("Affinity == %d, want 18", affinity)
	}
}

// TestLegacyAffinityIncludesRaceAffinity checks that the total affinity adds the race
// affinity to the base affinity
func TestLegacyAffinityIncludesRaceAffinity(t *testing.T) {
	dataStore := setup(t)
	legacy := Legacy{
		TraineeCharaId: 1001,
		Parent1: Parent{
			Member: Member{
				CharaId:      1002,
				WinSaddleIds: []int{1, 2, 5, 10, 11, 12, 13, 15, 16, 17, 18, 23, 25, 26, 27, 34, 63, 145, 146, 147},
			},
			GrandParent1: Member{
				CharaId:      1003,
				WinSaddleIds: []int{4, 5, 6, 10, 13, 14, 15, 17, 23, 26, 27, 61, 122, 130},
			},
			GrandParent2: Member{
				CharaId:      1005,
				WinSaddleIds: []int{2, 6, 7, 10, 11, 14, 15, 17, 18, 21, 23, 25, 26, 29, 32, 34, 35, 39, 65, 85},
			},
		},
		Parent2: Parent{
			Member:       Member{CharaId: 1003},
			GrandParent1: Member{CharaId: 1008},
			GrandParent2: Member{CharaId: 1015},
		},
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 144+18 {
		t.Errorf("Legacy %v == %d, want %d", legacy, affinity, 125+18)
	}
}
