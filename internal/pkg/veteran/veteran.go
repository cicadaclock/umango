package veteran

import (
	"encoding/json"
	"fmt"
	"os"
)

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

func LoadVeteranList(path string) ([]Veteran, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}
	var veteranList []Veteran
	err = json.Unmarshal(file, &veteranList)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling veteran list: %w", err)
	}
	return veteranList, nil
}
