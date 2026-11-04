// Package data holds the master.mdb tables that the application reads, loaded into memory once.
package data

import (
	"fmt"

	"github.com/cicadaclock/umango/internal/db"
	"golang.org/x/sync/errgroup"
)

// skill_data rarity values of the skills that upgrade another skill
const (
	skillRarityWhite = 1
	skillRarityGold  = 2
)

// DataStore holds the master.mdb lookup tables in memory
type DataStore struct {
	// Card ID to chara ID
	cardData map[int]int
	// Succession relation type to affinity points
	successionRelations map[int]int
	// Chara ID to the succession relation types that include the chara
	successionRelationMembers map[int][]int
	// Factor ID to factor type
	factorType map[int]int
	// Raw score ID to base score
	ttRawScores map[int]int
	// Race instance ID to course distance in meters
	courseDistances map[int]int
	// Skill ID to skill point cost before any hint discount
	skillNeedPoints map[int]int
	// Skill ID to rarity
	skillRarities map[int]int

	// Skill ID to skill name
	skillNames map[int]string
	// Factor ID to factor name
	factorNames map[int]string
	// Card ID to chara title and name
	veteranCardId map[int]string
	// Chara ID to chara name
	charaNames map[int]string
	// Race track ID to race track name
	raceTrackNames map[int]string

	// Races from which team trials draw
	teamTrialPool []db.TeamTrialRace
	// Combinations of season, weather and ground condition that races draw
	raceConditions []db.RaceCondition
	// Skills that any uma can learn with skill points
	generalSkills []db.GeneralSkill
	// Base scores of a team trial race and the conditions that earn them
	scoreRules []db.ScoreRule
}

// Init opens master.mdb and loads every lookup table into memory concurrently
func Init() (*DataStore, error) {
	dataStore := DataStore{}
	var err error

	db, err := db.Open()
	if err != nil {
		return &dataStore, fmt.Errorf("opening master.mdb: %w", err)
	}
	defer db.SqlDB.Close()

	var g errgroup.Group
	g.Go(func() (err error) { dataStore.cardData, err = db.CardData(); return })
	g.Go(func() (err error) { dataStore.successionRelations, err = db.SuccessionRelations(); return })
	g.Go(func() (err error) { dataStore.successionRelationMembers, err = db.SuccessionRelationMembers(); return })
	g.Go(func() (err error) { dataStore.skillNames, err = db.TextDataSkills(); return })
	g.Go(func() (err error) { dataStore.factorNames, err = db.TextDataFactors(); return })
	g.Go(func() (err error) { dataStore.veteranCardId, err = db.TextDataVeteranCardId(); return })
	g.Go(func() (err error) { dataStore.charaNames, err = db.TextDataCharaName(); return })
	g.Go(func() (err error) { dataStore.raceTrackNames, err = db.TextDataRaceTracks(); return })
	g.Go(func() (err error) { dataStore.factorType, err = db.SuccessionFactors(); return })
	g.Go(func() (err error) { dataStore.ttRawScores, err = db.TeamStadiumRawScores(); return })
	g.Go(func() (err error) { dataStore.courseDistances, err = db.RaceInstanceDistances(); return })
	g.Go(func() (err error) { dataStore.skillNeedPoints, err = db.SkillNeedPoints(); return })
	g.Go(func() (err error) { dataStore.skillRarities, err = db.SkillRarities(); return })
	g.Go(func() (err error) { dataStore.teamTrialPool, err = db.TeamTrialPool(); return })
	g.Go(func() (err error) { dataStore.raceConditions, err = db.RaceConditions(); return })
	g.Go(func() (err error) { dataStore.generalSkills, err = db.GeneralSkills(); return })
	g.Go(func() (err error) { dataStore.scoreRules, err = db.ScoreRules(); return })
	if err := g.Wait(); err != nil {
		return &dataStore, fmt.Errorf("load db data: %w", err)
	}

	return &dataStore, nil
}

// TotalSkillPoints returns the skill point cost of learned skills before any hint discount. A gold
// skill includes the cost of the white skill that it upgrades, and a double circle skill includes
// the cost of its single circle form.
func (dataStore *DataStore) TotalSkillPoints(skillIds []int) int {
	total := 0
	for _, id := range skillIds {
		total += dataStore.skillNeedPoints[id]
		switch dataStore.skillRarities[id] {
		case skillRarityGold:
			upgraded := id - 1
			if id%10 == 1 {
				upgraded = id + 1
			}
			total += dataStore.skillNeedPoints[upgraded]
		case skillRarityWhite:
			if id%10 == 1 && dataStore.skillRarities[id+1] == skillRarityWhite {
				total += dataStore.skillNeedPoints[id+1]
			}
		}
	}
	return total
}

// SkillNames returns the name of each skill ID
func (dataStore *DataStore) SkillNames(ids []int) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.skillNames[id])
	}
	return result
}

// TeamTrialPool returns the races from which team trials draw. The caller must not change them.
func (dataStore *DataStore) TeamTrialPool() []db.TeamTrialRace {
	return dataStore.teamTrialPool
}

// RaceConditions returns the combinations of season, weather and ground condition that races draw.
// The caller must not change them.
func (dataStore *DataStore) RaceConditions() []db.RaceCondition {
	return dataStore.raceConditions
}

// GeneralSkills returns the skills that any uma can learn with skill points. The caller must not
// change them.
func (dataStore *DataStore) GeneralSkills() []db.GeneralSkill {
	return dataStore.generalSkills
}

// ScoreRules returns the base scores of a team trial race and the conditions that earn them. The
// caller must not change them.
func (dataStore *DataStore) ScoreRules() []db.ScoreRule {
	return dataStore.scoreRules
}

// FactorNames returns the name of each factor ID
func (dataStore *DataStore) FactorNames(ids []int) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.factorNames[id])
	}
	return result
}

// FactorClass returns the factor type of each factor ID
func (dataStore *DataStore) FactorClass(ids []int) []int {
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.factorType[id])
	}
	return result
}

// TTRawScores returns the base score of each raw score ID
func (dataStore *DataStore) TTRawScores(ids []int) []int {
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		result = append(result, dataStore.ttRawScores[id])
	}
	return result
}

// FactorLevels returns the star level of each factor ID, held in its last two digits
func (dataStore *DataStore) FactorLevels(ids []int) []int {
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		result = append(result, id%100)
	}
	return result
}

// VeteranCardCharaTitle returns the title and name of each card ID, as "[Title] Name"
func (dataStore *DataStore) VeteranCardCharaTitle(veteranCardIds []int) []string {
	names := make([]string, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		names = append(names, dataStore.veteranCardId[id])
	}
	return names
}

// VeteranCardChara returns the chara ID of each card ID
func (dataStore *DataStore) VeteranCardChara(veteranCardIds []int) []int {
	ids := make([]int, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		charaId := dataStore.cardData[id]
		ids = append(ids, charaId)
	}
	return ids
}

// VeteranCardCharaName returns the chara name of each card ID, without the title
func (dataStore *DataStore) VeteranCardCharaName(veteranCardIds []int) []string {
	names := make([]string, 0, len(veteranCardIds))
	for _, id := range veteranCardIds {
		charaId := dataStore.cardData[id]
		charaName := dataStore.charaNames[charaId]
		names = append(names, charaName)
	}
	return names
}

// CardChara returns the chara ID of a card ID
func (dataStore *DataStore) CardChara(cardId int) int {
	return dataStore.cardData[cardId]
}

// CharaName returns the name of a chara ID
func (dataStore *DataStore) CharaName(charaId int) string {
	return dataStore.charaNames[charaId]
}

// RaceTrackName returns the name of a race track ID
func (dataStore *DataStore) RaceTrackName(raceTrackId int) string {
	return dataStore.raceTrackNames[raceTrackId]
}

// CourseDistance returns the course distance in meters of a race instance ID, or 0 when unknown
func (dataStore *DataStore) CourseDistance(raceInstanceId int) int {
	return dataStore.courseDistances[raceInstanceId]
}

// RelationMembers returns the succession relation types that include a chara ID
func (dataStore *DataStore) RelationMembers(charaId int) []int {
	return dataStore.successionRelationMembers[charaId]
}

// RelationPoint returns the affinity points of a succession relation type
func (dataStore *DataStore) RelationPoint(relationType int) int {
	return dataStore.successionRelations[relationType]
}
