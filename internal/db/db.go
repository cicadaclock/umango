// Package db reads tables from the game's master.mdb SQLite database.
package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// text_data categories that the application reads
const (
	textDataCardId    = 4
	textDataCharaName = 6
	textDataSkills    = 47
	textDataFactors   = 147
	textDataTTScore   = 140
	textDataTTBonus   = 148
)

// DB is a read-only connection to master.mdb
type DB struct {
	SqlDB *sql.DB
}

// Open opens a read-only connection to master.mdb at the install path of this platform
func Open() (*DB, error) {
	dbPath, err := DBPath()
	if err != nil {
		return nil, fmt.Errorf("getting master.mdb path: %w", err)
	}
	// master.mdb is static, so immutable skips locking and journal checks
	dsn := "file:" + filepath.ToSlash(dbPath) + "?mode=ro&immutable=1"
	sqlDb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening master.mdb at %s: %w", dbPath, err)
	}
	return &DB{SqlDB: sqlDb}, nil
}

// CardData returns the chara ID of each card ID
func (db *DB) CardData() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, "SELECT t.id, t.chara_id FROM card_data AS t")
	if err != nil {
		return nil, fmt.Errorf("card_data: %w", err)
	}
	return result, nil
}

// SuccessionRelations returns the affinity points of each succession relation type
func (db *DB) SuccessionRelations() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, "SELECT t.relation_type, t.relation_point FROM succession_relation AS t")
	if err != nil {
		return nil, fmt.Errorf("succession_relation: %w", err)
	}
	return result, nil
}

// SuccessionRelationMembers returns the succession relation types that include each chara ID
func (db *DB) SuccessionRelationMembers() (map[int][]int, error) {
	rows, err := db.SqlDB.Query("SELECT t.chara_id, t.relation_type FROM succession_relation_member AS t")
	if err != nil {
		return nil, fmt.Errorf("query succession_relation_member rows: %w", err)
	}
	defer rows.Close()

	result := make(map[int][]int, 200)
	var charaId int
	var relationType int
	for rows.Next() {
		if err := rows.Scan(&charaId, &relationType); err != nil {
			return nil, fmt.Errorf("scanning succession_relation_member rows: %w", err)
		}
		result[charaId] = append(result[charaId], relationType)
	}
	return result, rows.Err()
}

// SuccessionFactors returns the factor type of each factor ID
func (db *DB) SuccessionFactors() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, "SELECT t.factor_id, t.factor_type FROM succession_factor AS t")
	if err != nil {
		return nil, fmt.Errorf("succession_factor: %w", err)
	}
	return result, nil
}

// TeamStadiumRawScores returns the base score of each team trial raw score ID
func (db *DB) TeamStadiumRawScores() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, "SELECT t.id, t.score FROM team_stadium_raw_score AS t")
	if err != nil {
		return nil, fmt.Errorf("team_stadium_raw_score: %w", err)
	}
	return result, nil
}

// RaceInstanceDistances returns the course distance in meters of each race instance ID
func (db *DB) RaceInstanceDistances() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, `SELECT ri.id, rcs.distance FROM race_instance AS ri
		JOIN race AS r ON r.id = ri.race_id
		JOIN race_course_set AS rcs ON rcs.id = r.course_set`)
	if err != nil {
		return nil, fmt.Errorf("race_instance distances: %w", err)
	}
	return result, nil
}

// TextDataSkills returns the name of each skill ID
func (db *DB) TextDataSkills() (map[int]string, error) {
	result, err := db.textData(textDataSkills, 0, 0, false)
	if err != nil {
		return nil, fmt.Errorf("get skills: %w", err)
	}
	return result, nil
}

// TextDataFactors returns the name of each factor ID
func (db *DB) TextDataFactors() (map[int]string, error) {
	result, err := db.textData(textDataFactors, 0, 0, false)
	if err != nil {
		return nil, fmt.Errorf("get factors (skill sparks): %w", err)
	}
	return result, nil
}

// TextDataCharaName returns the name of each chara ID
func (db *DB) TextDataCharaName() (map[int]string, error) {
	result, err := db.textData(textDataCharaName, 0, 0, false)
	if err != nil {
		return nil, fmt.Errorf("get chara id: %w", err)
	}
	return result, nil
}

// TextDataVeteranCardId returns the title and name of each card ID
func (db *DB) TextDataVeteranCardId() (map[int]string, error) {
	result, err := db.textData(textDataCardId, 0, 0, false)
	if err != nil {
		return nil, fmt.Errorf("get card id: %w", err)
	}
	return result, nil
}

// textData returns the text of each index in a text_data category. When maxIndex is above 0,
// only indices between minIndex and maxIndex are read, or only those outside when between is false.
func (db *DB) textData(category, minIndex, maxIndex int, between bool) (map[int]string, error) {
	if minIndex > maxIndex {
		return nil, fmt.Errorf("minimum index larger than max index")
	}

	query := fmt.Sprintf("SELECT t.'index', text FROM text_data AS t WHERE (t.category = %d)", category)
	if maxIndex > 0 {
		if between {
			query += fmt.Sprintf(" AND (t.'index' BETWEEN %d AND %d)", minIndex, maxIndex)
		} else {
			query += fmt.Sprintf(" AND (t.'index' NOT BETWEEN %d AND %d)", minIndex, maxIndex)
		}
	}
	return queryMap[string](db.SqlDB, query)
}

// queryMap runs a query of two columns and returns the second column keyed by the first
func queryMap[V int | string](sqlDB *sql.DB, query string) (map[int]V, error) {
	rows, err := sqlDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query rows: %w", err)
	}
	defer rows.Close()

	result := make(map[int]V, 1024)
	var key int
	var value V
	for rows.Next() {
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scanning rows: %w", err)
		}
		result[key] = value
	}
	return result, rows.Err()
}
