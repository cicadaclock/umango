package data

import (
	"reflect"
	"testing"
)

func setup(t *testing.T) *DataStore {
	dataStore, err := Init()
	if err != nil {
		t.Errorf("init dataStore: %v", err)
	}
	return dataStore
}

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
		"Power ★★★",
		"Mile ★★",
		"Oka Sho ★",
		"Yasuda Kinen ★",
		"Corner Recovery ○ ★",
		"Nimble Navigator ★",
		"Shifting Gears ★★★",
		"Murmur ★",
		"Lucky Seven ★★★",
	}
	result := dataStore.MapFactorNames(factorIds)
	if !reflect.DeepEqual(result, want) {
		t.Errorf("factor ids: %v\nGot: %v\nWant: %v", factorIds, result, want)
	}
}

func TestMapVeteranCardIdToCharaName(t *testing.T) {
	dataStore := setup(t)

	veteranCardId := []int{101801}
	result := dataStore.MapVeteranCardIdToCharaName(veteranCardId)
	want := []string{"Air Groove"}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("MapVeteranCardIdToCharaName(%d) == \"%s\", want: %s", veteranCardId, result, want)
	}
}
