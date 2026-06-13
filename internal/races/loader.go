package races

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func LoadRaceResultsFolder(directoryPath string) ([]RaceResult, error) {
	// 20 TT samples, 5 race results per TT sample = 100 results
	allRaceResults := make([]RaceResult, 0, 100)
	filepath.WalkDir(directoryPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		} else if filepath.Ext(d.Name()) != ".json" {
			return nil
		}

		raceResults, err := LoadRaceResults(path)
		// If parsing race results fails, just skip it
		if err != nil {
			return nil
		}
		allRaceResults = append(allRaceResults, raceResults...)
		return nil
	})
	return allRaceResults, nil
}

func LoadRaceResults(path string) ([]RaceResult, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	var teamTrial struct {
		RaceResultArray []RaceResult `json:"race_result_array"`
	}
	if err := json.Unmarshal(file, &teamTrial); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return teamTrial.RaceResultArray, nil
}
