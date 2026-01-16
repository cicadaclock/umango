// Interface for handling the data from veterans.json in memory
package veteran

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ichiban/soa"
)

const (
	// Main Legacy
	LegacyNameMain LegacyName = 0
	// Left side legacy parent
	LegacyNameParentL LegacyName = 10
	// Left side legacy grandparent 1
	LegacyNameGrandparentL1 LegacyName = 11
	// Left side legacy grandparent 2
	LegacyNameGrandparentL2 LegacyName = 12
	// Right side legacy
	LegacyNameParentR LegacyName = 20
	// Right side legacy grandparent 1
	LegacyNameGrandparentR1 LegacyName = 21
	// Right side legacy grandparent 2
	LegacyNameGrandparentR2 LegacyName = 22
)

//go:generate go tool soagen
type Veteran struct {
	// Metadata

	// Veteran ID that is unique locally to your account
	LocalVeteranId int    `json:"single_mode_chara_id"`
	CardId         int    `json:"card_id"`
	CreateTime     string `json:"create_time"`
	RankScore      int    `json:"rank_score"`
	// Sparks
	FactorIdArray []int `json:"factor_id_array"`

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

// Legacy parent and grandparent
type SuccessionChara struct {
	// Sparks
	FactorIdArray []int `json:"factor_id_array"`
	// Races placed 1st in
	WinSaddleIdArray []int `json:"win_saddle_id_array"`
	// Legacy position (parent 1 or 2, grandparent 1 or 2)
	PositionId int `json:"position_id"`
}

func Init(path string) (*VeteranSlice, error) {
	veterans, err := loadVeterans(path)
	if err != nil {
		return nil, fmt.Errorf("load veteran list: %w", err)
	}
	veteranSlice := soa.Make[VeteranSlice](0, len(veterans))
	for _, veteran := range veterans {
		veteranSlice = soa.Append(veteranSlice, veteran)
	}
	return &veteranSlice, nil
}

// Returns an array of factors for a given legacy
func (v *Veteran) LegacyFactor(l LegacyName) []int {
	for _, successionChara := range v.SuccessionCharaArray {
		if successionChara.PositionId == l.Int() {
			return successionChara.FactorIdArray
		}
	}
	return []int{}
}

func loadVeterans(path string) ([]Veteran, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	var veteranList []Veteran
	err = json.Unmarshal(file, &veteranList)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return veteranList, nil
}

// Internal ID for distinguishing between different legacy umas
type LegacyName int

// Returns the integer representation of the legacy name
func (l LegacyName) Int() int {
	return int(l)
}
