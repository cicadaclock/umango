// Interface for handling the data from veterans.json in memory

package veteran

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ichiban/soa"
)

const (
	// Left side legacy
	LegacyParentL       LegacyName = 10
	LegacyGrandparentL1 LegacyName = 11
	LegacyGrandparentL2 LegacyName = 12
	// Right side legacy
	LegacyParentR       LegacyName = 20
	LegacyGrandparentR1 LegacyName = 21
	LegacyGrandparentR2 LegacyName = 22
)

//go:generate go tool soagen
type Veteran struct {
	// Metadata
	CardId        int    `json:"card_id"`
	CreateTime    string `json:"create_time"`
	RankScore     int    `json:"rank_score"`
	FactorIdArray []int  `json:"factor_id_array"`

	// Racing stats
	Speed   int `json:"speed"`
	Stamina int `json:"stamina"`
	Power   int `json:"power"`
	Guts    int `json:"guts"`
	Wit     int `json:"wiz"`

	// Legacy/affinity calcs
	SuccessionCharaArray []SuccessionChara `json:"succession_chara_array"`
	WinSaddleIdArray     []int             `json:"win_saddle_id_array"`
	NicknameIdArray      []int             `json:"nickname_id_array"` // Epithets, maybe relevant?
}

type SuccessionChara struct {
	FactorIdArray    []int `json:"factor_id_array"`
	WinSaddleIdArray []int `json:"win_saddle_id_array"`
	PositionId       int   `json:"position_id"`
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
func (v *Veteran) SuccessionCharaFactors(l LegacyName) []int {
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

type LegacyName int

// Returns the integer representation of the legacy name
func (l LegacyName) Int() int {
	return int(l)
}
