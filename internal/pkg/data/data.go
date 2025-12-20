// Interface for handling the data from master.mdb in memory

package data

import "fmt"

// Anything we want to store as a global state of data
type DataStore struct {
	CardData                  map[int]int
	SuccessionRelations       map[int]int
	SuccessionRelationMembers map[int][]int
	// Text mapping
	SkillSparkNames map[int]string
	VeteranCardId   map[int]string
}

// Load DB tables into memory
func New(db DB) (*DataStore, error) {
	dataStore := DataStore{}
	var err error
	dataStore.CardData, err = db.CardData()
	if err != nil {
		return &dataStore, fmt.Errorf("loading card_data into memory: %w", err)
	}
	dataStore.SuccessionRelations, err = db.SuccessionRelations()
	if err != nil {
		return &dataStore, fmt.Errorf("loading succession_relation into memory: %w", err)
	}
	dataStore.SuccessionRelationMembers, err = db.SuccessionRelationMembers()
	if err != nil {
		return &dataStore, fmt.Errorf("loading succession_relation_member into memory: %w", err)
	}
	dataStore.SkillSparkNames, err = db.TextDataSkillSpark()
	if err != nil {
		return &dataStore, fmt.Errorf("loading text_data for skill sparks into memory: %w", err)
	}
	dataStore.VeteranCardId, err = db.TextDataVeteranCardId()
	if err != nil {
		return &dataStore, fmt.Errorf("loading text_data for card id into memory: %w", err)
	}
	return &dataStore, nil
}
