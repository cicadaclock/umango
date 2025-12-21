// Interface for querying master.mdb directly

package db

import (
	"database/sql"
	"fmt"

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
	db := DB{}
	dbPath, err := DBPath()
	if err != nil {
		return nil, fmt.Errorf("getting master.mdb path: %w", err)
	}
	sqlDb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening master.mdb at %s: %w", dbPath, err)
	}
	db.SqlDB = sqlDb
	return &db, nil
}

// Map card_id to chara_id from card_data
func (db *DB) CardData(c chan map[int]int, errCh chan error) {
	result := make(map[int]int, 100)
	rows, err := db.SqlDB.Query("SELECT t.id, t.chara_id FROM card_data AS t")
	if err != nil {
		errCh <- fmt.Errorf("query card_data rows: %w", err)
		return
	}
	defer rows.Close()
	var id int
	var chara_id int
	for rows.Next() {
		err := rows.Scan(&id, &chara_id)
		if err != nil {
			errCh <- fmt.Errorf("scanning card_data rows, %w", err)
			return
		}
		result[id] = chara_id
	}
	c <- result
}

// Map relation_type to relation_point from succession_relation
func (db *DB) SuccessionRelations(c chan map[int]int, errCh chan error) {
	result := make(map[int]int, 1000)
	rows, err := db.SqlDB.Query("SELECT t.relation_type, t.relation_point FROM succession_relation AS t")
	if err != nil {
		errCh <- fmt.Errorf("query succession_relation rows: %w", err)
		return
	}
	defer rows.Close()
	var relation_type int
	var relation_point int
	for rows.Next() {
		err := rows.Scan(&relation_type, &relation_point)
		if err != nil {
			errCh <- fmt.Errorf("scanning succession_relation rows, %w", err)
			return
		}
		result[relation_type] = relation_point
	}
	c <- result
}

// Map chara_id to []relation_type from succession_relation_member
func (db *DB) SuccessionRelationMembers(c chan map[int][]int, errCh chan error) {
	// Get unique chara_ids
	charaIds := make([]int, 0, 200)
	rows, err := db.SqlDB.Query("SELECT t.chara_id FROM succession_relation_member AS t GROUP BY t.chara_id")
	if err != nil {
		errCh <- fmt.Errorf("query succession_relation_member for all chara_id: %w", err)
		return
	}
	defer rows.Close()
	var chara_id int
	for rows.Next() {
		err := rows.Scan(&chara_id)
		if err != nil {
			errCh <- fmt.Errorf("scanning succession_relation_member rows, %w", err)
			return
		}
		charaIds = append(charaIds, chara_id)
	}

	// Get array of relation_type for each chara_id
	result := make(map[int][]int, len(charaIds))
	query, err := db.SqlDB.Prepare("SELECT t.relation_type FROM succession_relation_member AS t WHERE t.chara_id = ?")
	if err != nil {
		errCh <- fmt.Errorf("prepare succession_relation_member query: %w", err)
		return
	}
	defer query.Close()
	for _, id := range charaIds {
		rows2, err := query.Query(id)
		// rows2, err := db.SqlDB.Query("SELECT t.relation_type FROM succession_relation_member AS t WHERE t.chara_id = 1001")
		if err != nil {
			errCh <- fmt.Errorf("query succession_relation_member rows where chara_id = %d: %w", id, err)
			return
		}
		defer rows2.Close()
		var relation_type int
		relationTypeList := make([]int, 0, 200)
		for rows2.Next() {
			err := rows2.Scan(&relation_type)
			if err != nil {
				errCh <- fmt.Errorf("scanning succession_relation_member rows, %w", err)
				return
			}
			relationTypeList = append(relationTypeList, relation_type)
		}
		result[id] = relationTypeList
	}
	c <- result
}

// Map index to text from text_data
func (db *DB) textData(category, minIndex, maxIndex int, between bool) (map[int]string, error) {
	if minIndex > maxIndex {
		return nil, fmt.Errorf("minimum index larger than max index")
	}

	result := make(map[int]string, 2000)
	query := fmt.Sprintf("SELECT t.'index', text FROM text_data AS t WHERE (t.category = %d)", category)
	if maxIndex > 0 {
		if between {
			query += fmt.Sprintf(" AND (t.'index' BETWEEN %d AND %d)", minIndex, maxIndex)
		} else {
			query += fmt.Sprintf(" AND (t.'index' NOT BETWEEN %d AND %d)", minIndex, maxIndex)
		}
	}

	rows, err := db.SqlDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query text_data rows: %w", err)
	}
	defer rows.Close()
	var index int
	var text string
	for rows.Next() {
		err := rows.Scan(&index, &text)
		if err != nil {
			return nil, fmt.Errorf("scanning text_data rows, %w", err)
		}
		result[index] = text
	}
	return result, nil
}

// Map skill_id to text from text_data
func (db *DB) TextDataFactors(c chan map[int]string, errCh chan error) {
	result, err := db.textData(textDataFactors, 0, 0, false)
	if err != nil {
		errCh <- fmt.Errorf("get factors (skill sparks): %w", err)
		return
	}
	c <- result
}

// Map chara_id to text from text_data
func (db *DB) TextDataCharaName(c chan map[int]string, errCh chan error) {
	result, err := db.textData(textDataCharaName, 0, 0, false)
	if err != nil {
		errCh <- fmt.Errorf("get chara id: %w", err)
		return
	}
	c <- result
}

// Map card_id to text from text_data
func (db *DB) TextDataVeteranCardId(c chan map[int]string, errCh chan error) {
	result, err := db.textData(textDataCardId, 0, 0, false)
	if err != nil {
		errCh <- fmt.Errorf("get card id: %w", err)
		return
	}
	c <- result
}
