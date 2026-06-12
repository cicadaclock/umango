// Interface for querying master.mdb directly

package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// text_data categories
	textDataCardId    = 4
	textDataCharaName = 6
	textDataFactors   = 147
)

type DB struct {
	SqlDB *sql.DB
}

// Create DB struct with connection to master.mdb
func Open() (*DB, error) {
	dbPath, err := DBPath()
	if err != nil {
		return nil, fmt.Errorf("getting master.mdb path: %w", err)
	}
	// master.mdb is static data, open it read-only immutable to skip locking
	// and journal checks
	dsn := "file:" + filepath.ToSlash(dbPath) + "?mode=ro&immutable=1"
	sqlDb, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening master.mdb at %s: %w", dbPath, err)
	}
	return &DB{SqlDB: sqlDb}, nil
}

// Map card_id to chara_id from card_data
func (db *DB) CardData() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, "SELECT t.id, t.chara_id FROM card_data AS t")
	if err != nil {
		return nil, fmt.Errorf("card_data: %w", err)
	}
	return result, nil
}

// Map relation_type to relation_point from succession_relation
func (db *DB) SuccessionRelations() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, "SELECT t.relation_type, t.relation_point FROM succession_relation AS t")
	if err != nil {
		return nil, fmt.Errorf("succession_relation: %w", err)
	}
	return result, nil
}

// Map chara_id to []relation_type from succession_relation_member
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

// Map factor_id to factor_type from succession_factor
func (db *DB) SuccessionFactors() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, "SELECT t.factor_id, t.factor_type FROM succession_factor AS t")
	if err != nil {
		return nil, fmt.Errorf("succession_factor: %w", err)
	}
	return result, nil
}

// Map skill_id to text from text_data
func (db *DB) TextDataFactors() (map[int]string, error) {
	result, err := db.textData(textDataFactors, 0, 0, false)
	if err != nil {
		return nil, fmt.Errorf("get factors (skill sparks): %w", err)
	}
	return result, nil
}

// Map chara_id to text from text_data
func (db *DB) TextDataCharaName() (map[int]string, error) {
	result, err := db.textData(textDataCharaName, 0, 0, false)
	if err != nil {
		return nil, fmt.Errorf("get chara id: %w", err)
	}
	return result, nil
}

// Map card_id to text from text_data
func (db *DB) TextDataVeteranCardId() (map[int]string, error) {
	result, err := db.textData(textDataCardId, 0, 0, false)
	if err != nil {
		return nil, fmt.Errorf("get card id: %w", err)
	}
	return result, nil
}

// Map index to text from text_data
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

// Scan the rows of a two column query into a map
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
