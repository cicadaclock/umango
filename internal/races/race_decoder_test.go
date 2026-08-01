package races

import (
	"testing"
)

func TestDecodeRaceScenario(t *testing.T) {
	ttr, err := LoadRaces("../testdata/team_trial.json")
	if err != nil {
		t.Fatalf("LoadRaces: %v", err)
	}

	for i, raceResult := range ttr.RaceResultArray {
		scenario, err := raceResult.DecodeScenario()
		if err != nil {
			t.Fatalf("round %d: DecodeScenario: %v", i+1, err)
		}

		horses := ttr.RaceStartParamsArray[i].RaceHorseDataArray
		if len(scenario.HorseResults) != len(horses) {
			t.Errorf("round %d: got %d horse results, want %d", i+1, len(scenario.HorseResults), len(horses))
			continue
		}
		if len(scenario.Frames) == 0 {
			t.Errorf("round %d: no frames decoded", i+1)
			continue
		}

		// Finish orders are seen once and don't go out of bounds
		seen := make([]bool, len(scenario.HorseResults))
		for _, result := range scenario.HorseResults {
			if result.FinishOrder < 0 || int(result.FinishOrder) >= len(seen) || seen[result.FinishOrder] {
				t.Errorf("round %d: bad finish order %d", i+1, result.FinishOrder)
				continue
			}
			seen[result.FinishOrder] = true
		}

		// Frame times must ascend and horses only move forward
		previous := scenario.Frames[0]
		for _, frame := range scenario.Frames[1:] {
			if frame.Time <= previous.Time {
				t.Errorf("round %d: frame times not ascending: %f then %f", i+1, previous.Time, frame.Time)
				break
			}
			for h := range frame.Horses {
				if frame.Horses[h].Distance < previous.Horses[h].Distance {
					t.Errorf("round %d: horse %d moved backwards", i+1, h)
				}
			}
			previous = frame
		}

		// Winning horse has no finish diff
		for _, result := range scenario.HorseResults {
			if result.FinishOrder == 0 && result.FinishDiff != 0 {
				t.Errorf("round %d: winner has finish diff %f, want 0", i+1, result.FinishDiff)
			}
		}

		// Every skill activation must reference a skill the horse actually has
		activations := scenario.SkillActivations()
		if len(activations) == 0 {
			t.Errorf("round %d: no skill activations decoded", i+1)
		}
		for _, activation := range activations {
			if activation.HorseIndex < 0 || activation.HorseIndex >= len(horses) {
				t.Errorf("round %d: skill activation with bad horse index %d", i+1, activation.HorseIndex)
				continue
			}
			owned := false
			// Horses are parsed in the same order as frame order
			// I don't know if this order is guaranteed though
			if horses[activation.HorseIndex].FrameOrder != (activation.HorseIndex + 1) {
				t.Errorf("round %d: skill activation with bad horse index %d != %d", i+1, activation.HorseIndex, horses[activation.HorseIndex].FrameOrder)
				continue
			}
			for _, skill := range horses[activation.HorseIndex].SkillArray {
				if skill.SkillId == activation.SkillId {
					owned = true
					break
				}
			}
			if !owned {
				t.Errorf("round %d: horse %d activated skill %d it does not own",
					i+1, activation.HorseIndex, activation.SkillId)
			}
		}
	}
}

func TestDecodeRaceScenarioRejectsBadInput(t *testing.T) {
	if _, err := DecodeRaceScenario(""); err == nil {
		t.Error("want error for empty scenario")
	}
	if _, err := DecodeRaceScenario("not base64 at all!"); err == nil {
		t.Error("want error for invalid base64")
	}
	// Valid base64 of a valid gzip stream with invalid data
	if _, err := DecodeRaceScenario("H4sIAAAAAAAAA0tMSk4BAM3vwqkEAAAA"); err == nil {
		t.Error("want error for truncated binary payload")
	}
}
