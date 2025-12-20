package affinity

import (
	"testing"

	"github.com/cicadaclock/umango/internal/pkg/data"
)

func TestLegacyAffinityForFullLegacy(t *testing.T) {
	var db data.DB
	db.Open()
	defer db.SqlDB.Close()
	dataStore, err := data.New(db)
	if err != nil {
		t.Errorf("error loading data store: %v", err)
	}

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
	var db data.DB
	db.Open()
	defer db.SqlDB.Close()
	dataStore, err := data.New(db)
	if err != nil {
		t.Errorf("error loading data store: %v", err)
	}

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
	var db data.DB
	db.Open()
	defer db.SqlDB.Close()
	dataStore, err := data.New(db)
	if err != nil {
		t.Errorf("error loading data store: %v", err)
	}

	legacy := Legacy{}
	affinity := legacy.Affinity(dataStore)
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}

func TestLegacyAffinityForSameUmaInParent(t *testing.T) {
	var db data.DB
	db.Open()
	defer db.SqlDB.Close()
	dataStore, err := data.New(db)
	if err != nil {
		t.Errorf("error loading data store: %v", err)
	}

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
	var db data.DB
	db.Open()
	defer db.SqlDB.Close()
	dataStore, err := data.New(db)
	if err != nil {
		t.Errorf("error loading data store: %v", err)
	}

	legacy := Legacy{
		CharaId00: 1001,
	}
	affinity := legacy.Affinity(dataStore)
	if affinity != 0 {
		t.Errorf("Legacy %v == %d", legacy, affinity)
	}
}
