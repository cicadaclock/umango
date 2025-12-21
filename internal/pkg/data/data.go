// Interface for handling the data from master.mdb in memory

package data

import (
	"fmt"

	"github.com/cicadaclock/umango/internal/pkg/db"
)

// Anything we want to store as a single source of data
type DataStore struct {
	CardData                  map[int]int
	SuccessionRelations       map[int]int
	SuccessionRelationMembers map[int][]int
	// Text mapping
	FactorNames   map[int]string
	VeteranCardId map[int]string
	CharaNames    map[int]string
}

// Load DB tables into memory
func Init() (*DataStore, error) {
	dataStore := DataStore{}
	var err error

	// Open DB connection
	db, err := db.Open()
	if err != nil {
		return &dataStore, fmt.Errorf("opening master.mdb: %w", err)
	}
	defer db.SqlDB.Close()

	// Store DB results into memory
	chCount := 6
	chCardData := make(chan map[int]int)
	chSuccessionRelations := make(chan map[int]int)
	chSuccessionRelationMembers := make(chan map[int][]int)
	chFactorNames := make(chan map[int]string)
	chVeteranCardId := make(chan map[int]string)
	chCharaNames := make(chan map[int]string)
	errCh := make(chan error)

	go db.CardData(chCardData, errCh)
	go db.SuccessionRelations(chSuccessionRelations, errCh)
	go db.SuccessionRelationMembers(chSuccessionRelationMembers, errCh)
	go db.TextDataFactors(chFactorNames, errCh)
	go db.TextDataVeteranCardId(chVeteranCardId, errCh)
	go db.TextDataCharaName(chCharaNames, errCh)

	for range chCount {
		select {
		case err := <-errCh:
			close(errCh)
			return &dataStore, fmt.Errorf("load db data: %w", err)
		case dataStore.CardData = <-chCardData:
		case dataStore.SuccessionRelations = <-chSuccessionRelations:
		case dataStore.SuccessionRelationMembers = <-chSuccessionRelationMembers:
		case dataStore.FactorNames = <-chFactorNames:
		case dataStore.VeteranCardId = <-chVeteranCardId:
		case dataStore.CharaNames = <-chCharaNames:
		}
	}

	return &dataStore, nil
}

func (dataStore *DataStore) MapFactorNames(ids []int) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.FactorNames[id])
	}
	return result
}
