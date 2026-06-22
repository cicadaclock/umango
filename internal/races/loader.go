package races

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func LoadRacesFolder(directoryPath string) (TeamTrialResultSet, error) {
	resultSet := TeamTrialResultSet{}
	filepath.WalkDir(directoryPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		} else if filepath.Ext(d.Name()) != ".json" {
			return nil
		}

		teamTrialResult, err := LoadRaces(path)
		// If parsing race results fails, just skip it
		if err != nil {
			return nil
		}
		if len(teamTrialResult.RaceResultArray) != len(teamTrialResult.RaceStartParamsArray) {
			return nil
		}

		resultSet.append(teamTrialResult)
		return nil
	})

	return resultSet, nil
}

func LoadRaces(path string) (TeamTrialResult, error) {
	var teamTrialRace TeamTrialResult
	if path == "" {
		return teamTrialRace, fmt.Errorf("empty path")
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return teamTrialRace, fmt.Errorf("read file %s: %w", path, err)
	}
	if err := json.Unmarshal(file, &teamTrialRace); err != nil {
		return teamTrialRace, fmt.Errorf("unmarshal: %w", err)
	}
	return teamTrialRace, nil
}
