package races

import "github.com/cicadaclock/umango/internal/veteran"

// RaceStartParams is the starting state of one team trial round
type RaceStartParams struct {
	Round int `json:"round"`
	// race_instance ID in master.mdb, which identifies the course
	RaceInstanceId     int             `json:"race_instance_id"`
	RaceHorseDataArray []RaceHorseData `json:"race_horse_data_array"`
}

// RaceHorseData is one uma entered in a team trial round
type RaceHorseData struct {
	// 1 for the player's team; other values are opponents
	TeamId int `json:"team_id"`
	// Veteran ID, unique within its owner's account
	TrainedCharaId int `json:"trained_chara_id"`
	// Card ID, which identifies the chara and outfit
	CardId int `json:"card_id"`
	// Gate number, counted from 1
	FrameOrder   int             `json:"frame_order"`
	SkillArray   []veteran.Skill `json:"skill_array"`
	RunningStyle RunStyle        `json:"running_style"`
}

// GetMyUmas returns the umas of the player's team in the round
func (rsp RaceStartParams) GetMyUmas() []RaceHorseData {
	umas := make([]RaceHorseData, 0, 3)
	for _, rhd := range rsp.RaceHorseDataArray {
		if rhd.ThatsMyUma() {
			umas = append(umas, rhd)
		}
	}
	return umas
}

// ThatsMyUma reports whether the uma belongs to the player's team
func (rhd RaceHorseData) ThatsMyUma() bool {
	return rhd.TeamId == 1
}
