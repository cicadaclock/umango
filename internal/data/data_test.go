package data

import (
	"reflect"
	"testing"
)

// setup loads master.mdb from the local game install, which these tests require
func setup(t *testing.T) *DataStore {
	dataStore, err := Init()
	if err != nil {
		t.Errorf("init dataStore: %v", err)
	}
	return dataStore
}

// TestFactorNames checks the names of stat, aptitude, race, skill and unique skill factors
func TestFactorNames(t *testing.T) {
	dataStore := setup(t)

	factorIds := []int{
		303,
		3202,
		1000401,
		1001101,
		2003501,
		2004901,
		2010503,
		2011601,
		2015603,
	}
	want := []string{
		"Power",
		"Mile",
		"Oka Sho",
		"Yasuda Kinen",
		"Corner Recovery ○",
		"Nimble Navigator",
		"Shifting Gears",
		"Murmur",
		"Lucky Seven",
	}
	result := dataStore.FactorNames(factorIds)
	if !reflect.DeepEqual(result, want) {
		t.Errorf("factor ids: %v\nGot: %v\nWant: %v", factorIds, result, want)
	}
}

// TestMapVeteranCardIdToCharaId checks the chara ID of a card ID
func TestMapVeteranCardIdToCharaId(t *testing.T) {
	dataStore := setup(t)

	veteranCardId := []int{101801}
	result := dataStore.VeteranCardChara(veteranCardId)
	want := []int{1018}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("MapVeteranCardIdToCharaName(%d) == \"%d\", want: %d", veteranCardId, result, want)
	}
}

// TestMapVeteranCardIdToCharaName checks the chara name of a card ID
func TestMapVeteranCardIdToCharaName(t *testing.T) {
	dataStore := setup(t)

	veteranCardId := []int{101801}
	result := dataStore.VeteranCardCharaName(veteranCardId)
	want := []string{"Air Groove"}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("MapVeteranCardIdToCharaName(%d) == \"%s\", want: %s", veteranCardId, result, want)
	}
}

// TestMapVeteranCardIdToCharaTitle checks the title and name of a card ID
func TestMapVeteranCardIdToCharaTitle(t *testing.T) {
	dataStore := setup(t)

	veteranCardId := []int{101801}
	result := dataStore.VeteranCardCharaTitle(veteranCardId)
	want := []string{"[Empress Road] Air Groove"}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("MapVeteranCardIdToCharaName(%d) == \"%s\", want: %s", veteranCardId, result, want)
	}
}
