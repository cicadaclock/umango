package veteran

import (
	"testing"
)

func TestLoadVeteranList(t *testing.T) {
	list, err := LoadVeteranList("../../testdata/single_veteran.json")
	if err != nil {
		t.Errorf("loading veteran list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("length of test veteran list == %d, want 1", len(list))
	}
}
