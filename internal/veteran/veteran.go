// Interface for handling the data from veterans.json in memory
package veteran

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ichiban/soa"
)

//go:generate go tool soagen
type Veteran struct {
	// Metadata

	// Veteran ID that is unique locally to your account
	LocalVeteranId int `json:"trained_chara_id"`
	// Veteran card ID that determines the chara
	CardId     int    `json:"card_id"`
	CreateTime string `json:"create_time"`
	RankScore  int    `json:"rank_score"`
	// Sparks
	FactorIdArray []int `json:"factor_id_array"`
	// Skills
	SkillArray []Skill `json:"skill_array"`

	// Racing stats

	Speed   int `json:"speed"`
	Stamina int `json:"stamina"`
	Power   int `json:"power"`
	Guts    int `json:"guts"`
	Wit     int `json:"wiz"`

	// Legacy/affinity calcs

	// Legacy parent and grandparent umas
	SuccessionCharaArray []SuccessionChara `json:"succession_chara_array"`
	// Races placed 1st in
	WinSaddleIdArray []int `json:"win_saddle_id_array"`
	// Epithets
	NicknameIdArray []int `json:"nickname_id_array"`
}

type Skill struct {
	SkillId int `json:"skill_id"`
	Level   int `json:"level"`
}

// Legacy parent and grandparent
type SuccessionChara struct {
	// Veteran card ID that determines the chara
	CardId int `json:"card_id"`
	// Sparks
	FactorIdArray []int `json:"factor_id_array"`
	// Races placed 1st in
	WinSaddleIdArray []int `json:"win_saddle_id_array"`
	// Legacy position (parent 1 or 2, grandparent 1 or 2)
	PositionId int `json:"position_id"`
}

// Decode reads a veterans.json document into a VeteranSlice.
//
// Inputs:
// - r: Reader positioned at a JSON array of veterans
//
// Returns:
// - veterans: Decoded veterans, garbage collected and owned by the caller
// - err: JSON decoding failure
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
