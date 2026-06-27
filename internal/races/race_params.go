package races

import "github.com/cicadaclock/umango/internal/veteran"

// Starting parameters for a single team trial round
type RaceStartParams struct {
	Round           int `json:"round"`
	Season          int `json:"season"`
	Weather         int `json:"weather"`
	GroundCondition int `json:"ground_condition"`
	// Your team's evaluation score
	SelfEvaluate int `json:"self_evaluate"`
	// Opponent team's evaluation score
	OpponentEvaluate   int             `json:"opponent_evaluate"`
	RaceHorseDataArray []RaceHorseData `json:"race_horse_data_array"`
}

// Similar to veteran but some fields have different names and some are
// irrelevant to veterans so a new struct is used
type RaceHorseData struct {
	Distance    DistanceType
	TrainerName string `json:"trainer_name"`
	TeamId      int    `json:"team_id"`
	// Unique chara ID
	TrainedCharaId int `json:"trained_chara_id"`
	// Chara info (so we can tell apart alts)
	CardId      int `json:"card_id"`
	TalentLevel int `json:"talent_level"`
	// Starting gate
	FrameOrder int             `json:"frame_order"`
	SkillArray []veteran.Skill `json:"skill_array"`
	// Stats
	Stamina      int `json:"stamina"`
	Speed        int `json:"speed"`
	Pow          int `json:"pow"`
	Guts         int `json:"guts"`
	Wiz          int `json:"wiz"`
	RunningStyle int `json:"running_style"`
	RaceDressId  int `json:"race_dress_id"`
	FinalGrade   int `json:"final_grade"`
	// Distance aptitudes
	ProperDistanceShort  int `json:"proper_distance_short"`
	ProperDistanceMile   int `json:"proper_distance_mile"`
	ProperDistanceMiddle int `json:"proper_distance_middle"`
	ProperDistanceLong   int `json:"proper_distance_long"`
	// Running style aptitudes
	ProperRunningStyleNige   int `json:"proper_running_style_nige"`
	ProperRunningStyleSenko  int `json:"proper_running_style_senko"`
	ProperRunningStyleSashi  int `json:"proper_running_style_sashi"`
	ProperRunningStyleOikomi int `json:"proper_running_style_oikomi"`
	// Ground aptitudes
	ProperGroundTurf int `json:"proper_ground_turf"`
	ProperGroundDirt int `json:"proper_ground_dirt"`
	// Mood
	Motivation int `json:"motivation"`
}

// Filters RaceHorseDataArray for only our umas
func (rsp RaceStartParams) GetMyUmas() []RaceHorseData {
	// 3 umas per race
	umas := make([]RaceHorseData, 0, 3)
	for _, rhd := range rsp.RaceHorseDataArray {
		if rhd.ThatsMyUma() {
			umas = append(umas, rhd)
		}
	}
	return umas
}

// Checks if this uma belongs to our account
func (rhd RaceHorseData) ThatsMyUma() bool {
	return rhd.TeamId == 1
}
