// Interface for handling the data from master.mdb in memory

package data

import (
	"fmt"
	"strings"

	"github.com/cicadaclock/umango/internal/pkg/db"
)

// Anything we want to store as a single source of data
type DataStore struct {
	CardData                  map[int]int
	SuccessionRelations       map[int]int
	SuccessionRelationMembers map[int][]int
	FactorType                map[int]int

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
	chCount := 7
	chCardData := make(chan map[int]int)
	chSuccessionRelations := make(chan map[int]int)
	chSuccessionRelationMembers := make(chan map[int][]int)
	chFactorNames := make(chan map[int]string)
	chVeteranCardId := make(chan map[int]string)
	chCharaNames := make(chan map[int]string)
	chFactorType := make(chan map[int]int)
	errCh := make(chan error)

	go db.CardData(chCardData, errCh)
	go db.SuccessionRelations(chSuccessionRelations, errCh)
	go db.SuccessionRelationMembers(chSuccessionRelationMembers, errCh)
	go db.TextDataFactors(chFactorNames, errCh)
	go db.TextDataVeteranCardId(chVeteranCardId, errCh)
	go db.TextDataCharaName(chCharaNames, errCh)
	go db.SuccessionFactors(chFactorType, errCh)

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
		case dataStore.FactorType = <-chFactorType:
		}
	}

	return &dataStore, nil
}

func (dataStore *DataStore) MapFactorNames(ids []int) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		var level strings.Builder
		level.WriteString(" ")
		for range id % 100 {
			_, _ = level.WriteString("★")
		}
		result = append(result, dataStore.FactorNames[id]+level.String())
	}
	return result
}

func (dataStore *DataStore) MapVeteranCardIdName(veteranCardIds []int) []string {
	names := make([]string, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		names = append(names, dataStore.VeteranCardId[id])
	}
	return names
}

func (dataStore *DataStore) MapVeteranCardIdToCharaName(veteranCardIds []int) []string {
	names := make([]string, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		charaId := dataStore.CardData[id]
		charaName := dataStore.CharaNames[charaId]
		names = append(names, charaName)
	}
	return names
}

// Internal ID for distinguishing between different factor types
type FactorType int

const (
	FactorTypeBlue  FactorType = 1
	FactorTypeRed   FactorType = 2
	FactorTypeGreen FactorType = 3
	FactorTypeWhite FactorType = 4
	FactorTypeRace  FactorType = 5
)

func (ft FactorType) Int() int {
	return int(ft)
}
