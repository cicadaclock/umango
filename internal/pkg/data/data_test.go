package data

import (
	"testing"
)

func TestSkillName(t *testing.T) {
	var db DB
	db.Open()
	defer db.SqlDB.Close()
	dataStore, err := New(db)
	if err != nil {
		t.Errorf("error loading data store: %v", err)
	}
	skillName := (*dataStore).SkillSparkNames[10680103]
	if skillName != "Victory Cheer!" {
		t.Errorf("SkillSparkNames[10680103] = \"%s\", want \"Victory Cheer!\"", skillName)
	}
}
