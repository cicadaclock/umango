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
	textDataRaceTrack = 31
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

// TeamTrialRace is one race of the team trial pool
type TeamTrialRace struct {
	RaceInstanceId, CourseSetId, Distance, Ground int
}

// RaceCondition is one combination of season, weather and ground condition that a race can draw
type RaceCondition struct {
	Season, Weather, GroundCondition int
	// Weight of this combination among all rows of race_condition
	Rate int
}

// GeneralSkill is one skill that any uma can learn with skill points
type GeneralSkill struct {
	Id int
	// Tiers of one skill share a group
	GroupId int
	// Tier within the group, counted from 1 for the lowest tier
	Tier int
}

// ScoreRule is one row of team_stadium_raw_score: a base score and the condition that earns it
type ScoreRule struct {
	ConditionType, Value1, Value2, Score int
}

// TeamTrialPool returns the races from which team trials draw, ordered by track and distance.
// master.mdb does not mark this pool. Race group 9 holds three instances of each course, and saved
// team trial results only ever use the third, whose ID is a multiple of 3.
func (db *DB) TeamTrialPool() ([]TeamTrialRace, error) {
	result, err := queryRows(db.SqlDB, `SELECT ri.id, rcs.id, rcs.distance, rcs.ground
		FROM race AS r
		JOIN race_instance AS ri ON ri.race_id = r.id
		JOIN race_course_set AS rcs ON rcs.id = r.course_set
		WHERE r."group" = 9 AND ri.id % 3 = 0
		ORDER BY rcs.race_track_id, rcs.distance, rcs.id`,
		func(rows *sql.Rows, r *TeamTrialRace) error {
			return rows.Scan(&r.RaceInstanceId, &r.CourseSetId, &r.Distance, &r.Ground)
		})
	if err != nil {
		return nil, fmt.Errorf("team trial pool: %w", err)
	}
	return result, nil
}

// RaceConditions returns the combinations of season, weather and ground condition that races draw
func (db *DB) RaceConditions() ([]RaceCondition, error) {
	result, err := queryRows(db.SqlDB, "SELECT t.season, t.weather, t.ground, t.rate FROM race_condition AS t WHERE t.rate > 0",
		func(rows *sql.Rows, c *RaceCondition) error {
			return rows.Scan(&c.Season, &c.Weather, &c.GroundCondition, &c.Rate)
		})
	if err != nil {
		return nil, fmt.Errorf("race_condition: %w", err)
	}
	return result, nil
}

// GeneralSkills returns the skills that any uma can learn with skill points: those with a cost and
// an ID in the general range, without uniques, inherited uniques or negative conditions
func (db *DB) GeneralSkills() ([]GeneralSkill, error) {
	result, err := queryRows(db.SqlDB, `SELECT s.id, s.group_id, s.group_rate FROM skill_data AS s
		JOIN single_mode_skill_need_point AS n ON n.id = s.id
		WHERE s.id / 100000 = 2 AND s.group_rate > 0
		ORDER BY s.id`,
		func(rows *sql.Rows, s *GeneralSkill) error {
			return rows.Scan(&s.Id, &s.GroupId, &s.Tier)
		})
	if err != nil {
		return nil, fmt.Errorf("general skills: %w", err)
	}
	return result, nil
}

// ScoreRules returns the base scores of a team trial race and the conditions that earn them
func (db *DB) ScoreRules() ([]ScoreRule, error) {
	result, err := queryRows(db.SqlDB, `SELECT t.condition_type, t.condition_value_1, t.condition_value_2, t.score
		FROM team_stadium_raw_score AS t`,
		func(rows *sql.Rows, r *ScoreRule) error {
			return rows.Scan(&r.ConditionType, &r.Value1, &r.Value2, &r.Score)
		})
	if err != nil {
		return nil, fmt.Errorf("team_stadium_raw_score rules: %w", err)
	}
	return result, nil
}

// SkillNeedPoints returns the skill point cost of each skill ID, before any hint discount
func (db *DB) SkillNeedPoints() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, "SELECT t.id, t.need_skill_point FROM single_mode_skill_need_point AS t")
	if err != nil {
		return nil, fmt.Errorf("single_mode_skill_need_point: %w", err)
	}
	return result, nil
}

// SkillRarities returns the rarity of each skill ID
func (db *DB) SkillRarities() (map[int]int, error) {
	result, err := queryMap[int](db.SqlDB, "SELECT t.id, t.rarity FROM skill_data AS t")
	if err != nil {
		return nil, fmt.Errorf("skill_data: %w", err)
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

// TextDataRaceTracks returns the name of each race track ID
func (db *DB) TextDataRaceTracks() (map[int]string, error) {
	result, err := db.textData(textDataRaceTrack, 0, 0, false)
	if err != nil {
		return nil, fmt.Errorf("get race tracks: %w", err)
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

// queryRows runs a query and returns one value per row, each filled by scan
func queryRows[T any](sqlDB *sql.DB, query string, scan func(rows *sql.Rows, value *T) error) ([]T, error) {
	rows, err := sqlDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query rows: %w", err)
	}
	defer rows.Close()

	var result []T
	for rows.Next() {
		var value T
		if err := scan(rows, &value); err != nil {
			return nil, fmt.Errorf("scanning rows: %w", err)
		}
		result = append(result, value)
	}
	return result, rows.Err()
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
