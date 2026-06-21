package races

import (
	"testing"
)

func TestLoadRaceResultsForEmptyPath(t *testing.T) {
	_, err := LoadRaces("")
	if err == nil {
		t.Error("want error for empty path")
	}
}
