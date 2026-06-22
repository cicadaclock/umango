package races

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func LoadRacesFolder(directoryPath string) ([]TeamTrialRace, error) {
	allRaceResults := make([]TeamTrialRace, 0, 100)
	filepath.WalkDir(directoryPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		} else if filepath.Ext(d.Name()) != ".json" {
			return nil
		}

		teamTrialRaceResult, err := LoadRaces(path)
		// If parsing race results fails, just skip it
		if err != nil {
			return nil
		}
		if len(teamTrialRaceResult.RaceResultArray) != len(teamTrialRaceResult.RaceStartParamsArray) {
			return nil
		}

		allRaceResults = append(allRaceResults, teamTrialRaceResult)
		return nil
	})

	return allRaceResults, nil
}

func LoadRaces(path string) (TeamTrialRace, error) {
	var teamTrialRace TeamTrialRace
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
