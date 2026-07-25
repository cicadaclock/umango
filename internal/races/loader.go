package races

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/sync/errgroup"
)

func LoadRacesFolder(directoryPath string) (TeamTrialResultSet, error) {
	resultSet := TeamTrialResultSet{}
	var paths []string
	err := filepath.WalkDir(directoryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d == nil {
				return err
			}
			return nil
		}
		if d.IsDir() || filepath.Ext(d.Name()) != ".json" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return resultSet, err
	}

	// Pre-size the slice so order is retained
	parsed := make([]*TeamTrialResult, len(paths))
	var g errgroup.Group
	g.SetLimit(runtime.GOMAXPROCS(0))
	for i, path := range paths {
		g.Go(func() error {
			teamTrialResult, err := LoadRaces(path)
			// If parsing race results fails, just skip it
			if err != nil {
				return nil
			}
			// Check correctness of data
			if !teamTrialResult.IsValidData() {
				return nil
			}
			parsed[i] = &teamTrialResult
			return nil
		})
	}
	_ = g.Wait()
	resultSet.Set = make([]TeamTrialResult, 0, len(paths))
	for _, teamTrialResult := range parsed {
		if teamTrialResult != nil {
			resultSet.append(*teamTrialResult)
		}
	}
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
