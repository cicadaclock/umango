package structs

import (
	"time"
)

type Uma struct {
	Id   int
	Name string
}

type RelationMember struct {
	A          int
	RelationId int
	UmaId      int
}

type Skill struct {
	Level   int `json:"level"`
	SkillId int `json:"skill_id"`
}

type RaceResult struct {
	GroundCondition int `json:"ground_condition"`
	Popularity      int `json:"popularity"`
	PrizeMoney      int `json:"prize_money"` // ? 0 for everything?
	ProgramId       int `json:"program_id"`  // ?
	ResultRank      int `json:"result_rank"` // Placement in the race
	ResultTime      int `json:"result_time"`
	RunningStyle    int `json:"running_style"`
	Turn            int `json:"turn"`
	Weather         int `json:"weather"`
}

// These are sparks, I think the grade data is being saved as 0 in the veteran lister?
type FactorInfo struct {
	FactorId int `json:"factor_id"` // Spark ID
	Grade    int `json:"level"`     // # of spark stars
}

type SupportCard struct {
	Exp             int `json:"exp"` // The support card's level
	LimitBreakCount int `json:"limit_break_count"`
	Position        int `json:"position"`
	SupportCardId   int `json:"support_card_id"`
}

type LegacyVeteranInfo struct {
	CardId           int          `json:"card_id"`             // Veteran card ID?
	FactorIdArray    []int        `json:"factor_id_array"`     // Sparks
	FactorInfoArray  []FactorInfo `json:"factor_info_array"`   // Sparks + Stars array
	OwnerViewerId    int          `json:"owner_viewer_id"`     // ? seems to always be 0
	PositionId       int          `json:"position_id"`         // ?
	Rank             int          `json:"chara_grade"`         // G=1, G+=2
	Rarity           int          `json:"rarity"`              // # of stars
	TalentLevel      int          `json:"talent_level"`        // Unique skill level
	WinSaddleIdArray []int        `json:"win_saddle_id_array"` // ? Shared race ID for affinity maybe?
}

type VeteranInfo struct {
	// Racing stats
	Speed   int `json:"speed"`
	Stamina int `json:"stamina"`
	Power   int `json:"power"`
	Guts    int `json:"guts"`
	Wit     int `json:"wiz"`
	// Racing Aptitudes
	// G=1, F=2, ...
	// Track
	Dirt int `json:"proper_ground_dirt"`
	Turf int `json:"proper_ground_turf"`
	// Distance
	Sprint int `json:"proper_distance_short"`
	Mile   int `json:"proper_distance_mile"`
	Medium int `json:"proper_distance_middle"`
	Long   int `json:"proper_distance_long"`
	// Style
	Front int `json:"proper_running_style_nige"`
	Pace  int `json:"proper_running_style_oikomi"`
	Late  int `json:"proper_running_style_sashi"`
	End   int `json:"proper_running_style_senko"`

	// Metadata
	ArriveRouteRaceId         int                 `json:"arrive_route_race_id"`          // ?
	CardId                    int                 `json:"card_id"`                       // Veteran card ID?
	CreateTime                time.Time           `json:"create_time"`                   // Same as register_time?
	RegisterTime              time.Time           `json:"register_time"`                 // Same as create_time?
	Seed                      int                 `json:"chara_seed"`                    // RNG seed for the career to determine stat/training hacks(?)
	Rank                      int                 `json:"chara_grade"`                   // G=1, G+=2
	RankScore                 int                 `json:"rank_score"`                    // Determines rank based on arbitrary cutoffs
	Rarity                    int                 `json:"rarity"`                        // # of stars
	Epithet                   int                 `json:"nickname_id"`                   //
	EpithetArray              []int               `json:"nickname_id_array"`             //
	Fans                      int                 `json:"fans"`                          //
	FactorIdArray             []int               `json:"factor_id_array"`               // Sparks
	FactorInfoArray           []FactorInfo        `json:"factor_info_array"`             // Sparks + Stars
	RaceResultList            []RaceResult        `json:"race_result_list"`              //
	IsLocked                  bool                `json:"is_locked"`                     //
	IsSaved                   bool                `json:"is_saved"`                      //
	OwnerTrainedCharaId       int                 `json:"owner_trained_chara_id"`        // ?
	OwnerViewerId             int                 `json:"owner_viewer_id"`               // ?
	RaceClothId               int                 `json:"race_cloth_id"`                 // ?
	RouteId                   int                 `json:"route_id"`                      // ?
	RunningStyle              int                 `json:"running_style"`                 // Default running style?
	ScenarioId                int                 `json:"scenario_id"`                   // 1=URA 2=Unity
	SingleplayerCharacterId   int                 `json:"single_mode_chara_id"`          // ID of the uma locally, should be the # of runs you did total on your account
	SuccessionCharacterArray  []LegacyVeteranInfo `json:"succession_chara_array"`        // Legacy, 6 umas
	SuccessionHistoryArray    []int               `json:"succession_history_array"`      // ?
	SuccessionNum             int                 `json:"succession_num"`                // ?
	SuccessionTrainedCharaId1 int                 `json:"succession_trained_chara_id_1"` // ?
	SuccessionTrainedCharaId2 int                 `json:"succession_trained_chara_id_2"` // ?
	SupportCardList           []SupportCard       `json:"support_card_list"`             //
	TalentLevel               int                 `json:"talent_level"`                  // Unique skill level
	TrainedCharaId            int                 `json:"trained_chara_id"`              // ?
	UseType                   int                 `json:"use_type"`                      // ?
	ViewerId                  int                 `json:"viewer_id"`                     // ?
	WinSaddleIdArray          []int               `json:"win_saddle_id_array"`           // ???
	Wins                      int                 `json:"wins"`                          //
}
