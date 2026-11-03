package racesim

import (
	"cmp"
	"slices"
)

// SkillHeatBins is the number of equal stretches into which a heat profile divides a course
const SkillHeatBins = 256

// SkillSpan is a stretch of one race in which one runner had one skill active. The engine library
// writes this layout, so the fields stay in the order of its SkillSpan.
type SkillSpan struct {
	// Index of the race in the seed list
	Race int32
	// Gate index of the runner
	Runner  int32
	SkillId int32
	// 1 when the skill has no time limit and lasts the whole race, otherwise 0
	Passive int32
	// Meters from the start at which the skill activated and at which the effect ended
	StartDistance, EndDistance float32
}

// SkillActivity summarizes where one runner had one skill active over the simulated races
type SkillActivity struct {
	SkillId int
	// Races in which the skill activated at least once
	RaceCount int
	// Mean meters from the start at which the skill activated
	MeanStart float64
	// The skill has no time limit in any activation, so it has no heat profile
	Passive bool
	// Races in which the skill was active somewhere in each bin of the course
	Heat [SkillHeatBins]uint16
}

// skillTally is a SkillActivity under construction
type skillTally struct {
	SkillActivity
	startSum  float64
	spanCount int
	// Last race counted, stored as the race index plus 1, overall and in each bin. Spans of one race
	// that overlap, such as the two details of a skill, count as one race.
	lastRace    int32
	lastRaceBin [SkillHeatBins]int32
}

// SummarizeSkills returns the skill activity of each runner in gate order, each ordered by mean
// start. The spans must be in race order.
func SummarizeSkills(spans []SkillSpan, runnerCount int, courseDistance float64) [][]SkillActivity {
	tallies := make([][]skillTally, runnerCount)
	for i := range spans {
		span := &spans[i]
		tally := findSkillTally(&tallies[span.Runner], int(span.SkillId))
		tally.startSum += float64(span.StartDistance)
		tally.spanCount++
		raceStamp := span.Race + 1
		if tally.lastRace != raceStamp {
			tally.lastRace = raceStamp
			tally.RaceCount++
		}
		if span.Passive != 0 {
			continue
		}
		tally.Passive = false
		last := skillHeatBin(float64(span.EndDistance), courseDistance)
		for bin := skillHeatBin(float64(span.StartDistance), courseDistance); bin <= last; bin++ {
			if tally.lastRaceBin[bin] != raceStamp {
				tally.lastRaceBin[bin] = raceStamp
				tally.Heat[bin]++
			}
		}
	}
	activities := make([][]SkillActivity, runnerCount)
	for runner := range tallies {
		activities[runner] = make([]SkillActivity, len(tallies[runner]))
		for i := range tallies[runner] {
			tally := &tallies[runner][i]
			tally.MeanStart = tally.startSum / float64(tally.spanCount)
			activities[runner][i] = tally.SkillActivity
		}
		slices.SortFunc(activities[runner], func(a, b SkillActivity) int {
			return cmp.Or(cmp.Compare(a.MeanStart, b.MeanStart), cmp.Compare(a.SkillId, b.SkillId))
		})
	}
	return activities
}

// findSkillTally returns the tally of a skill, adding it when absent. A runner has few enough
// skills that a linear search is the fastest lookup.
func findSkillTally(tallies *[]skillTally, skillId int) *skillTally {
	for i := range *tallies {
		if (*tallies)[i].SkillId == skillId {
			return &(*tallies)[i]
		}
	}
	*tallies = append(*tallies, skillTally{SkillActivity: SkillActivity{SkillId: skillId, Passive: true}})
	return &(*tallies)[len(*tallies)-1]
}

// skillHeatBin returns the bin of a heat profile that holds a distance
func skillHeatBin(distance, courseDistance float64) int {
	return min(max(int(distance/courseDistance*SkillHeatBins), 0), SkillHeatBins-1)
}
