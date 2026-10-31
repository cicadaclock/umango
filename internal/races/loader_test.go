package races

import (
	"testing"
)

// TestLoadRaceResultsForEmptyPath checks that an empty path returns an error
func TestLoadRaceResultsForEmptyPath(t *testing.T) {
	_, err := LoadRaces("")
	if err == nil {
		t.Error("want error for empty path")
	}
}
