package races

import "github.com/cicadaclock/umango/internal/veteran"

// Starting parameters for a single team trial round
type RaceStartParams struct {
	Round int `json:"round"`
	// race_instance id in master.mdb, which sets the course
	RaceInstanceId     int             `json:"race_instance_id"`
	RaceHorseDataArray []RaceHorseData `json:"race_horse_data_array"`
}

// An uma entered in a team trial round, similar to a veteran but with some fields named
// differently
type RaceHorseData struct {
	TeamId int `json:"team_id"`
	// Unique chara ID
	TrainedCharaId int `json:"trained_chara_id"`
	// Chara info (so we can tell apart alts)
	CardId int `json:"card_id"`
	// Starting gate
	FrameOrder   int             `json:"frame_order"` // ID in the race scenario blob
	SkillArray   []veteran.Skill `json:"skill_array"`
	RunningStyle RunStyle        `json:"running_style"`
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
