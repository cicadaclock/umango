package data

import (
	"database/sql"
	"log"
)

type DB struct {
	SqlDB *sql.DB
}

type SkillData struct {
	Id   int
	Name string
}

type RelationMember struct {
	Id           int
	RelationType int
	CharaId      int
}

// func (db DB) SuccessionRelationMember()

// Map column values for 'index' to 'text' in table 'text_data'
func (db DB) TextDataMap(sc TextDataSearchCriteria) (map[int]string, error) {
	// The largest result we care about, skill sparks, has ~1000 rows, so
	// 2000 is enough headroom for future skills without overdoing allocation.
	// If allocation is a concern then enable this section to fit every
	// map to its exact size.
	count := 2000
	countRows := false
	if countRows {
		countQuery := sc.CountQuery()
		rows, err := db.SqlDB.Query(countQuery)
		if err != nil {
			return nil, err
		}
		if rows.Next() {
			rows.Scan(&count)
		}
	}

	result := make(map[int]string, count)
	query := sc.Query()
	rows, err := db.SqlDB.Query(query)
	if err != nil {
		return result, err
	}
	var skillData SkillData
	for rows.Next() {
		rows.Scan(&skillData.Id, &skillData.Name)
		result[skillData.Id] = skillData.Name
	}
	return result, nil
}

func LoadDatabaseData() {
	dbPath, err := DBPath()
	if err != nil {
		log.Fatalf("Getting master.mdb path: %v", err)
	}

	sqlDb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer sqlDb.Close()

	db := DB{
		SqlDB: sqlDb,
	}

	_, err = db.TextDataMap(skillName[0])
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
