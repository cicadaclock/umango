// Interface for handling the data from master.mdb in memory

package data

import (
	"fmt"

	"github.com/cicadaclock/umango/internal/db"
	"golang.org/x/sync/errgroup"
)

// Anything we want to store as a single source of data
type DataStore struct {
	cardData                  map[int]int
	successionRelations       map[int]int
	successionRelationMembers map[int][]int
	factorType                map[int]int
	ttRawScores               map[int]int

	// Text mappings
	// Factor ID to factor name
	factorNames map[int]string
	// Veteran card ID to full chara title + name
	veteranCardId map[int]string
	// Veteran card ID to chara name
	charaNames map[int]string
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
	var g errgroup.Group
	g.Go(func() (err error) { dataStore.cardData, err = db.CardData(); return })
	g.Go(func() (err error) { dataStore.successionRelations, err = db.SuccessionRelations(); return })
	g.Go(func() (err error) { dataStore.successionRelationMembers, err = db.SuccessionRelationMembers(); return })
	g.Go(func() (err error) { dataStore.factorNames, err = db.TextDataFactors(); return })
	g.Go(func() (err error) { dataStore.veteranCardId, err = db.TextDataVeteranCardId(); return })
	g.Go(func() (err error) { dataStore.charaNames, err = db.TextDataCharaName(); return })
	g.Go(func() (err error) { dataStore.factorType, err = db.SuccessionFactors(); return })
	g.Go(func() (err error) { dataStore.ttRawScores, err = db.TeamStadiumRawScores(); return })
	if err := g.Wait(); err != nil {
		return &dataStore, fmt.Errorf("load db data: %w", err)
	}

	return &dataStore, nil
}

// Maps factor ID to factor name
func (dataStore *DataStore) FactorNames(ids []int) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.factorNames[id])
	}
	return result
}

// Maps factor ID to factor type (r/g/b/race/white)
func (dataStore *DataStore) FactorClass(ids []int) []int {
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.factorType[id])
	}
	return result
}

// Maps score ID to raw score value
func (dataStore *DataStore) TTRawScores(ids []int) []int {
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.ttRawScores[id])
	}
	return result
}

// Maps factor ID to factor level
func (dataStore *DataStore) FactorLevels(ids []int) []int {
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		result = append(result, id%100)
	}
	return result
}

// Maps veteran card ID to chara title. Example: "[Title] Name"
func (dataStore *DataStore) VeteranCardCharaTitle(veteranCardIds []int) []string {
	names := make([]string, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		names = append(names, dataStore.veteranCardId[id])
	}
	return names
}

// Maps veteran card ID to chara id
func (dataStore *DataStore) VeteranCardChara(veteranCardIds []int) []int {
	ids := make([]int, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		charaId := dataStore.cardData[id]
		ids = append(ids, charaId)
	}
	return ids
}

// Maps veteran card ID to chara name. Example: "Name"
func (dataStore *DataStore) VeteranCardCharaName(veteranCardIds []int) []string {
	names := make([]string, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		charaId := dataStore.cardData[id]
		charaName := dataStore.charaNames[charaId]
		names = append(names, charaName)
	}
	return names
}

// CardChara maps a card ID to its chara ID
func (dataStore *DataStore) CardChara(cardId int) int {
	return dataStore.cardData[cardId]
}

// CharaName maps a chara ID to its chara name
func (dataStore *DataStore) CharaName(charaId int) string {
	return dataStore.charaNames[charaId]
}

// RelationMembers maps a chara ID to its succession relation members
func (dataStore *DataStore) RelationMembers(charaId int) []int {
	return dataStore.successionRelationMembers[charaId]
}

// RelationPoint maps a succession relation type to its relation points
func (dataStore *DataStore) RelationPoint(relationType int) int {
	return dataStore.successionRelations[relationType]
}
