// Interface for handling the data from master.mdb in memory

package data

import (
	"fmt"

	"github.com/cicadaclock/umango/internal/db"
	"golang.org/x/sync/errgroup"
)

// Anything we want to store as a single source of data
type DataStore struct {
	CardData                  map[int]int
	SuccessionRelations       map[int]int
	SuccessionRelationMembers map[int][]int
	FactorType                map[int]int
	TTRawScores               map[int]int

	// Text mappings
	// Factor ID to factor name
	FactorNames map[int]string
	// Veteran card ID to full chara title + name
	VeteranCardId map[int]string
	// Veteran card ID to chara name
	CharaNames map[int]string
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
	g.Go(func() (err error) { dataStore.CardData, err = db.CardData(); return })
	g.Go(func() (err error) { dataStore.SuccessionRelations, err = db.SuccessionRelations(); return })
	g.Go(func() (err error) { dataStore.SuccessionRelationMembers, err = db.SuccessionRelationMembers(); return })
	g.Go(func() (err error) { dataStore.FactorNames, err = db.TextDataFactors(); return })
	g.Go(func() (err error) { dataStore.VeteranCardId, err = db.TextDataVeteranCardId(); return })
	g.Go(func() (err error) { dataStore.CharaNames, err = db.TextDataCharaName(); return })
	g.Go(func() (err error) { dataStore.FactorType, err = db.SuccessionFactors(); return })
	g.Go(func() (err error) { dataStore.TTRawScores, err = db.TeamStadiumRawScores(); return })
	if err := g.Wait(); err != nil {
		return &dataStore, fmt.Errorf("load db data: %w", err)
	}

	return &dataStore, nil
}

// Maps factor ID to factor name
func (dataStore *DataStore) MapFactorNames(ids []int) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.FactorNames[id])
	}
	return result
}

// Maps factor ID to factor type (r/g/b/race/white)
func (dataStore *DataStore) MapFactorTypes(ids []int) []int {
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.FactorType[id])
	}
	return result
}

// Maps score ID to raw score value
func (dataStore *DataStore) MapTTRawScores(ids []int) []int {
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.TTRawScores[id])
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
func (dataStore *DataStore) MapVeteranCardIdToCharaTitle(veteranCardIds []int) []string {
	names := make([]string, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		names = append(names, dataStore.VeteranCardId[id])
	}
	return names
}

// Maps veteran card ID to chara id
func (dataStore *DataStore) MapVeteranCardIdToCharaId(veteranCardIds []int) []int {
	ids := make([]int, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		charaId := dataStore.CardData[id]
		ids = append(ids, charaId)
	}
	return ids
}

// Maps veteran card ID to chara name. Example: "Name"
func (dataStore *DataStore) MapVeteranCardIdToCharaName(veteranCardIds []int) []string {
	names := make([]string, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		charaId := dataStore.CardData[id]
		charaName := dataStore.CharaNames[charaId]
		names = append(names, charaName)
	}
	return names
}
