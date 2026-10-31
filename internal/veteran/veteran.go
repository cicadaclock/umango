// Package veteran reads veterans.json files and calculates legacy affinity.
package veteran

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ichiban/soa"
)

// Veteran is one trained uma in a veterans.json file
//go:generate go tool soagen
type Veteran struct {
	// Veteran ID, unique within the player's account
	LocalVeteranId int `json:"trained_chara_id"`
	// Card ID, which identifies the chara and outfit
	CardId     int    `json:"card_id"`
	CreateTime string `json:"create_time"`
	RankScore  int    `json:"rank_score"`
	// Factor IDs of the veteran's own factors
	FactorIdArray []int   `json:"factor_id_array"`
	SkillArray    []Skill `json:"skill_array"`

	Speed   int `json:"speed"`
	Stamina int `json:"stamina"`
	Power   int `json:"power"`
	Guts    int `json:"guts"`
	Wit     int `json:"wiz"`

	// Legacy parents and grandparents of the veteran
	SuccessionCharaArray []SuccessionChara `json:"succession_chara_array"`
	// Win saddle IDs in ascending order
	WinSaddleIdArray []int `json:"win_saddle_id_array"`
	// Epithet IDs
	NicknameIdArray []int `json:"nickname_id_array"`
}

// Skill is one skill of an uma and its level
type Skill struct {
	SkillId int `json:"skill_id"`
	Level   int `json:"level"`
}

// SuccessionChara is one legacy parent or grandparent of a veteran
type SuccessionChara struct {
	// Card ID, which identifies the chara and outfit
	CardId int `json:"card_id"`
	// Factor IDs of this member's own factors
	FactorIdArray []int `json:"factor_id_array"`
	// Win saddle IDs in ascending order
	WinSaddleIdArray []int `json:"win_saddle_id_array"`
	// Position in the legacy tree; 10 and 20 are the veteran's two parents
	PositionId int `json:"position_id"`
}

// Decode reads a veterans.json document into a VeteranSlice
func Decode(r io.Reader) (*VeteranSlice, error) {
	var veterans []Veteran
	if err := json.NewDecoder(r).Decode(&veterans); err != nil {
		return nil, fmt.Errorf("decode veteran list: %w", err)
	}
	veteranSlice := soa.Make[VeteranSlice](0, len(veterans))
	for _, veteran := range veterans {
		veteranSlice = soa.Append(veteranSlice, veteran)
	}
	return &veteranSlice, nil
}
