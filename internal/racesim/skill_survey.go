package racesim

import (
	"cmp"
	"math/rand/v2"
	"slices"
	"sync/atomic"
	"time"

	"github.com/cicadaclock/umango/internal/db"
)

const (
	// Races per course between reports of the skill summaries
	SkillSurveyBatchRaces = 10
	// Batches of a complete skill survey
	SkillSurveyBatches = 10
	// Races per course of a complete skill survey
	SkillSurveyRaces = SkillSurveyBatches * SkillSurveyBatchRaces
	// Batches that every skill runs before the survey continues with only the contenders
	surveyFirstStageBatches = 3
	// Skills of each score column that continue after the first stage
	surveyContendersPerColumn = 30
	// Lone Wolf, which skill surveys leave out
	excludedSkillId = 201641
)

// MaxRunners is the most runners in one race
const MaxRunners = 18

// team_stadium_raw_score condition types that a simulated race can earn
const (
	scoreMargin = iota + 1
	scoreLongShot
	scorePlace
	scoreTargetTime
	scoreRushed
	scoreRushedPerSecond
	scoreStrongStart
	scoreSkill
	scorePositioning
)

// skill_data rarities as a score condition. Higher rarities are unique skills.
const (
	scoreSkillNormal = iota + 1
	scoreSkillRare
	scoreSkillFirstUnique
)

// ScoreRules holds the base scores of a team trial race. The engine library reads this layout, so
// the fields stay in the order of its ScoreRules.
type ScoreRules struct {
	// By finish order, counted from 0
	Place [12]int32
	// By winning margin, from Nose to Distance
	Margin [21]int32
	// By the number of tenths of a second under the standard course time
	TargetTime [21]int32
	// By rarity from the first unique rarity, then by skill level from 1
	UniqueSkill [3][10]int32
	RareSkill   int32
	NormalSkill int32
	LongShot    int32
	// A long shot finishes within this place, counted from 1
	LongShotPlace int32
	// A long shot is this favorite or lower
	LongShotPopularity int32
	Rushed             int32
	RushedPerSecond    int32
	StrongStart        int32
	// Ahead of all runners of the same running style going into the mid-race and the late race
	Positioning [2]int32
}

// newScoreRules arranges the rows of team_stadium_raw_score. Rows of other condition types never
// occur in saved team trial results, and a simulated race does not earn them.
func newScoreRules(rows []db.ScoreRule) ScoreRules {
	var rules ScoreRules
	for i := range rows {
		value, other, score := rows[i].Value1, rows[i].Value2, int32(rows[i].Score)
		switch rows[i].ConditionType {
		case scoreMargin:
			rules.Margin[value] = score
		case scoreLongShot:
			rules.LongShot, rules.LongShotPlace, rules.LongShotPopularity = score, int32(value), int32(other)
		case scorePlace:
			rules.Place[value-1] = score
		case scoreTargetTime:
			rules.TargetTime[value] = score
		case scoreRushed:
			rules.Rushed = score
		case scoreRushedPerSecond:
			rules.RushedPerSecond = score
		case scoreStrongStart:
			rules.StrongStart = score
		case scoreSkill:
			switch value {
			case scoreSkillNormal:
				rules.NormalSkill = score
			case scoreSkillRare:
				rules.RareSkill = score
			default:
				rules.UniqueSkill[value-scoreSkillFirstUnique][other-1] = score
			}
		case scorePositioning:
			rules.Positioning[value-1] = score
		}
	}
	return rules
}

// RaceVariation is the seed, conditions and gates of one race of a skill trial. The engine library
// reads this layout, so the fields stay in the order of its RaceVariation.
type RaceVariation struct {
	Seed, Season, Weather, GroundCondition int32
	// Index in the saved race of the runner at each gate
	GateOrder [MaxRunners]uint8
}

// SkillTrial is the runner, course, skills and races of one run of the library
type SkillTrial struct {
	// Index in the saved race of the runner under test
	Runner                      int
	CourseSetId, RaceInstanceId int32
	// In ascending order
	SkillIds []int32
	// For each skill, the lower tier that the runner owns and the skill would replace, otherwise 0
	ReplacedSkillIds []int32
	Variations       []RaceVariation
	Rules            *ScoreRules
	// One per variation, stored by Engine.WatchSkills
	Baselines []BaselineOutcome
	// len(SkillIds) activation flags per variation, stored by Engine.WatchSkills. A flag has bit 0 set
	// when the skill activated and bit 1 when the lower tier that it would replace activated.
	Activations []uint8
}

// activationFiredSkill is the bit of an activation flag that is set when the skill activated
const activationFiredSkill = 1

// BaselineOutcome is what the runner under test did in one race with no added skill. The engine
// library writes this layout, so the fields stay in the order of its BaselineOutcome.
type BaselineOutcome struct {
	// Team trial score before bonuses
	Score int32
	// Finish order, counted from 0
	FinishOrder int32
}

// SkillOutcome is what one added skill changed in one race, against the baseline of that race. The
// engine library writes this layout, so the fields stay in the order of its SkillOutcome.
type SkillOutcome struct {
	// 1 when the skill won the race, -1 when it lost a won race, otherwise 0
	WinDelta int32
	// Change of the team trial score before bonuses
	ScoreDelta int32
}

// SkillSurveyConfig selects the runner and the pool of races of a skill survey
type SkillSurveyConfig struct {
	Engine *Engine
	// The team trial round whose field runs on every course
	Round *TeamRound
	// The runner under test
	Runner *TeamRunner
	// The courses of the pool
	Courses []db.TeamTrialRace
	// The conditions from which each race draws
	Conditions []db.RaceCondition
	// The skills that a runner can learn
	Skills []db.GeneralSkill
	// The rows of team_stadium_raw_score
	ScoreRules []db.ScoreRule
	// Returns the skill point cost of learned skills, before any hint discount
	SkillPoints func(skillIds []int) int
}

// SkillSummary is the result of one skill over a pool of courses. A rate is a share from 0 to 1.
type SkillSummary struct {
	SkillId int
	// The lower tier that the runner owns and the skill would replace, otherwise 0
	ReplacedSkillId int
	// Skill points that the runner would spend beyond the lower tier, before any hint discount
	SkillPoints int
	// Races per course behind the win and score changes. A skill outside the contenders stops early.
	RacesPerCourse int
	// Rate of races in which the skill activated, over all courses of the pool
	ProcRate float64
	// Mean change of the win rate
	WinDelta float64
	// Mean change of the team trial score before bonuses, over all races of the pool. A race in which
	// the skill did not activate counts as no change.
	ScoreDelta float64
}

// SkillSurveyBaseline is what the runner under test did with no added skill, over all courses
type SkillSurveyBaseline struct {
	// Mean team trial score before bonuses
	Score float64
	// Rate of races won
	WinRate float64
}

// SkillSurveyReport is the state of a skill survey in progress
type SkillSurveyReport struct {
	// Skills that a runner could learn and those that can activate on some course of the pool
	CandidateSkills, PossibleSkills int
	// Course runs finished and in the complete survey, not counting the watching races
	CourseRuns, TotalCourseRuns int
	// Estimated seconds until the survey completes, 0 before the first course run finishes
	SecondsLeft float64
	// Races per course behind the win and score changes. The proc rates and the baseline come from the
	// watching races, so they are complete at 0.
	RacesPerCourse int
	// The result so far, which only the reports after the watching races and after each batch hold
	Summaries []SkillSummary
	Baseline  SkillSurveyBaseline
}

// surveyTally adds up the outcomes of one skill over all courses
type surveyTally struct {
	activated, winDelta, scoreDelta int64
	// Skill points that the runner would spend beyond the lower tier
	skillPoints int
	// Races per course whose win and score changes are added
	racesPerCourse int
	// Whether the skill can activate on some course
	possible bool
	// Whether the skill runs no more races
	stopped bool
}

// courseTrial is the skills and races of one course
type courseTrial struct {
	trial SkillTrial
	// Index in the survey of each skill of the trial
	skillIndexes []int32
	// Outcomes of one batch
	outcomes []SkillOutcome
}

// skillSurvey adds up the outcomes of every skill over a pool of courses
type skillSurvey struct {
	skillIds, replacedIds []int32
	tallies               []surveyTally
	courseCount           int
	// Races per course of the batches that finished
	racesPerCourse int
	// Over all courses
	baselineScore, baselineWins int64
}

// RunSkillSurvey simulates each skill that the runner could learn on every course of the pool. A
// first pass of watching races gives the baselines and every proc rate. Batches of races then
// measure what each skill changes, on only the races where it activated. After the first stage only
// the contenders of each score column run on. It reports after the first pass, after each course run
// and after each batch, and returns when the survey completes or soon after stop is set. The seeds
// follow from the round's seed, so a repeated survey gives the same result.
func RunSkillSurvey(config *SkillSurveyConfig, stop *atomic.Bool, report func(SkillSurveyReport)) error {
	race, err := config.Engine.Prepare(config.Round.SavedRace)
	if err != nil {
		return err
	}
	defer config.Engine.Release(race)

	skillIds, replacedIds := candidateSkills(config.Skills, config.Runner.SkillIds)
	survey := skillSurvey{
		skillIds:    skillIds,
		replacedIds: replacedIds,
		tallies:     make([]surveyTally, len(skillIds)),
		courseCount: len(config.Courses),
	}
	rules := newScoreRules(config.ScoreRules)
	trials := make([]courseTrial, len(config.Courses))
	state := SkillSurveyReport{
		CandidateSkills: len(skillIds),
		TotalCourseRuns: SkillSurveyBatches * len(config.Courses),
	}
	allConditions := conditionVariations(config.Conditions, config.Round.RunnerCount)
	flags := make([]uint8, len(skillIds))
	seedSource := uint64(config.Round.Seed)
	totalRate := int64(0)
	for i := range config.Conditions {
		totalRate += int64(config.Conditions[i].Rate)
	}
	for k := range skillIds {
		survey.tallies[k].skillPoints = skillPointsBeyondTier(config.SkillPoints, skillIds[k], replacedIds[k])
	}
	for c := range trials {
		if stop.Load() {
			return nil
		}
		course := &config.Courses[c]
		whole := SkillTrial{
			Runner:           config.Runner.Index,
			CourseSetId:      int32(course.CourseSetId),
			RaceInstanceId:   int32(course.RaceInstanceId),
			SkillIds:         skillIds,
			ReplacedSkillIds: replacedIds,
			Variations:       allConditions,
			Rules:            &rules,
		}
		if err := config.Engine.PossibleSkills(race, &whole, flags); err != nil {
			return err
		}
		trials[c] = survey.newCourseTrial(&whole, flags)
		drawVariations(trials[c].trial.Variations, config, totalRate, rand.New(rand.NewPCG(seedSource, uint64(c))))
		if err := config.Engine.WatchSkills(race, &trials[c].trial); err != nil {
			return err
		}
		survey.addWatchingRaces(&trials[c])
	}
	state.PossibleSkills = survey.possibleSkillCount()
	state.Summaries = survey.summaries()
	state.Baseline = survey.baseline()
	report(state)
	state.Summaries = nil

	began := time.Now()
	plannedRaces, racesDone := survey.expectedRaces(trials), 0
	for batch := range SkillSurveyBatches {
		for c := range trials {
			if stop.Load() {
				return nil
			}
			t := &trials[c]
			view := t.batchView(batch)
			if err := config.Engine.RunSkillTrial(race, &view, t.outcomes); err != nil {
				return err
			}
			survey.addOutcomes(t)
			racesDone += countRaces(view.Activations)
			state.CourseRuns++
			racesLeft := max(plannedRaces-racesDone, 0)
			state.SecondsLeft = time.Since(began).Seconds() * float64(racesLeft) / float64(max(racesDone, 1))
			report(state)
		}
		survey.finishBatch()
		if batch+1 == surveyFirstStageBatches {
			survey.stopOutsideContenders(trials)
			plannedRaces = racesDone + racesAfterBatch(trials, batch)
		}
		state.RacesPerCourse = survey.racesPerCourse
		state.Summaries = survey.summaries()
		report(state)
		state.Summaries = nil
	}
	return nil
}

// candidateSkills returns the skills that a runner could still learn. A skill counts when the
// runner owns no tier of it at the same level or above. Each skill comes with the lower tier that
// it would replace, otherwise 0. A double circle skill does not count, because it earns the same
// activation score as its single circle form at a higher cost.
func candidateSkills(skills []db.GeneralSkill, owned []int) (skillIds, replacedIds []int32) {
	ownedTiers := make(map[int]db.GeneralSkill, len(owned))
	for i := range skills {
		held, ok := ownedTiers[skills[i].GroupId]
		if slices.Contains(owned, skills[i].Id) && (!ok || held.Tier < skills[i].Tier) {
			ownedTiers[skills[i].GroupId] = skills[i]
		}
	}
	skillIds = make([]int32, 0, len(skills))
	replacedIds = make([]int32, 0, len(skills))
	for i := range skills {
		held, ok := ownedTiers[skills[i].GroupId]
		if skills[i].Id == excludedSkillId || skills[i].DoubleCircle || (ok && held.Tier >= skills[i].Tier) {
			continue
		}
		skillIds = append(skillIds, int32(skills[i].Id))
		replacedIds = append(replacedIds, int32(held.Id))
	}
	return skillIds, replacedIds
}

// conditionVariations returns one race for each condition, with the runners in saved order
func conditionVariations(conditions []db.RaceCondition, runnerCount int) []RaceVariation {
	variations := make([]RaceVariation, len(conditions))
	for i := range conditions {
		variations[i].Season = int32(conditions[i].Season)
		variations[i].Weather = int32(conditions[i].Weather)
		variations[i].GroundCondition = int32(conditions[i].GroundCondition)
		for gate := range runnerCount {
			variations[i].GateOrder[gate] = uint8(gate)
		}
	}
	return variations
}

// newCourseTrial records which skills can activate on a course and returns the trial of those
// skills
func (s *skillSurvey) newCourseTrial(whole *SkillTrial, flags []uint8) courseTrial {
	count := 0
	for _, flag := range flags {
		count += int(flag)
	}
	t := courseTrial{
		trial: SkillTrial{
			Runner:           whole.Runner,
			CourseSetId:      whole.CourseSetId,
			RaceInstanceId:   whole.RaceInstanceId,
			SkillIds:         make([]int32, 0, count),
			ReplacedSkillIds: make([]int32, 0, count),
			Variations:       make([]RaceVariation, SkillSurveyRaces),
			Rules:            whole.Rules,
			Baselines:        make([]BaselineOutcome, SkillSurveyRaces),
			Activations:      make([]uint8, count*SkillSurveyRaces),
		},
		skillIndexes: make([]int32, 0, count),
		outcomes:     make([]SkillOutcome, count*SkillSurveyBatchRaces),
	}
	for k, flag := range flags {
		if flag == 0 {
			continue
		}
		s.tallies[k].possible = true
		t.trial.SkillIds = append(t.trial.SkillIds, whole.SkillIds[k])
		t.trial.ReplacedSkillIds = append(t.trial.ReplacedSkillIds, whole.ReplacedSkillIds[k])
		t.skillIndexes = append(t.skillIndexes, int32(k))
	}
	return t
}

// drawVariations draws the seed, conditions and gates of each race. The game draws season, weather
// and ground together by the rates of race_condition, and gates at random.
func drawVariations(variations []RaceVariation, config *SkillSurveyConfig, totalRate int64, generator *rand.Rand) {
	conditions, runnerCount := config.Conditions, config.Round.RunnerCount
	for i := range variations {
		v := &variations[i]
		v.Seed = generator.Int32()
		draw := generator.Int64N(totalRate)
		k := 0
		for draw >= int64(conditions[k].Rate) {
			draw -= int64(conditions[k].Rate)
			k++
		}
		v.Season = int32(conditions[k].Season)
		v.Weather = int32(conditions[k].Weather)
		v.GroundCondition = int32(conditions[k].GroundCondition)
		for gate := range runnerCount {
			v.GateOrder[gate] = uint8(gate)
		}
		generator.Shuffle(runnerCount, func(a, b int) {
			v.GateOrder[a], v.GateOrder[b] = v.GateOrder[b], v.GateOrder[a]
		})
	}
}

// batchView returns the trial narrowed to the races of one batch
func (t *courseTrial) batchView(batch int) SkillTrial {
	view := t.trial
	first, last := batch*SkillSurveyBatchRaces, (batch+1)*SkillSurveyBatchRaces
	count := len(t.trial.SkillIds)
	view.Variations = t.trial.Variations[first:last]
	view.Baselines = t.trial.Baselines[first:last]
	view.Activations = t.trial.Activations[first*count : last*count]
	return view
}

// addWatchingRaces adds the baselines and the activations of the watching races of one course
func (s *skillSurvey) addWatchingRaces(t *courseTrial) {
	for i := range t.trial.Baselines {
		s.baselineScore += int64(t.trial.Baselines[i].Score)
		if t.trial.Baselines[i].FinishOrder == 0 {
			s.baselineWins++
		}
	}
	count := len(t.skillIndexes)
	for i, flag := range t.trial.Activations {
		s.tallies[t.skillIndexes[i%count]].activated += int64(flag & activationFiredSkill)
	}
}

// skillPointsBeyondTier returns what a skill costs beyond the lower tier that it would replace
func skillPointsBeyondTier(skillPoints func([]int) int, skillId, replacedId int32) int {
	points := skillPoints([]int{int(skillId)})
	if replacedId != 0 {
		points -= skillPoints([]int{int(replacedId)})
	}
	return points
}

// countRaces returns the races that activation flags call for. A race runs when its flag is set.
func countRaces(activations []uint8) int {
	races := 0
	for _, flag := range activations {
		if flag != 0 {
			races++
		}
	}
	return races
}

// racesAfterBatch returns the races that remain on all courses after a batch
func racesAfterBatch(trials []courseTrial, batch int) int {
	races := 0
	for c := range trials {
		t := &trials[c]
		races += countRaces(t.trial.Activations[(batch+1)*SkillSurveyBatchRaces*len(t.trial.SkillIds):])
	}
	return races
}

// expectedRaces estimates the races of the whole survey before the contenders are known: the first
// stage, then the share of the later races that the contenders would hold at most.
func (s *skillSurvey) expectedRaces(trials []courseTrial) int {
	later := racesAfterBatch(trials, surveyFirstStageBatches-1)
	total := racesAfterBatch(trials, -1)
	share := min(float64(2*surveyContendersPerColumn)/float64(max(s.possibleSkillCount(), 1)), 1)
	return total - later + int(float64(later)*share)
}

// finishBatch counts a batch that every course ran for the skills that still run
func (s *skillSurvey) finishBatch() {
	s.racesPerCourse += SkillSurveyBatchRaces
	for i := range s.tallies {
		if !s.tallies[i].stopped {
			s.tallies[i].racesPerCourse = s.racesPerCourse
		}
	}
}

// stopOutsideContenders stops each skill outside the leaders of both score columns. Its later races
// lose their flags, so the engine does not run them, and its result stays what the first stage found.
func (s *skillSurvey) stopOutsideContenders(trials []courseTrial) {
	indexes := make([]int, 0, len(s.tallies))
	for i := range s.tallies {
		if s.tallies[i].possible {
			indexes = append(indexes, i)
		}
	}
	keep := make([]bool, len(s.tallies))
	markLeaders := func(value func(*surveyTally) float64) {
		slices.SortStableFunc(indexes, func(a, b int) int {
			return cmp.Compare(value(&s.tallies[b]), value(&s.tallies[a]))
		})
		for _, i := range indexes[:min(surveyContendersPerColumn, len(indexes))] {
			keep[i] = true
		}
	}
	markLeaders(func(t *surveyTally) float64 { return float64(t.scoreDelta) })
	markLeaders(func(t *surveyTally) float64 { return float64(t.scoreDelta) / float64(max(t.skillPoints, 1)) })
	for i := range s.tallies {
		s.tallies[i].stopped = !keep[i]
	}
	firstLaterRow := surveyFirstStageBatches * SkillSurveyBatchRaces
	for c := range trials {
		t := &trials[c]
		count := len(t.skillIndexes)
		for k, index := range t.skillIndexes {
			if keep[index] {
				continue
			}
			for row := firstLaterRow; row < SkillSurveyRaces; row++ {
				t.trial.Activations[row*count+k] = 0
			}
		}
	}
}

// addOutcomes adds the outcomes of one batch of one course
func (s *skillSurvey) addOutcomes(t *courseTrial) {
	for i, k := range t.skillIndexes {
		tally := &s.tallies[k]
		for _, o := range t.outcomes[i*SkillSurveyBatchRaces : (i+1)*SkillSurveyBatchRaces] {
			tally.winDelta += int64(o.WinDelta)
			tally.scoreDelta += int64(o.ScoreDelta)
		}
	}
}

// possibleSkillCount returns the number of skills that can activate on some course
func (s *skillSurvey) possibleSkillCount() int {
	count := 0
	for i := range s.tallies {
		if s.tallies[i].possible {
			count++
		}
	}
	return count
}

// baseline returns the runner's results with no added skill
func (s *skillSurvey) baseline() SkillSurveyBaseline {
	races := float64(SkillSurveyRaces * s.courseCount)
	return SkillSurveyBaseline{Score: float64(s.baselineScore) / races, WinRate: float64(s.baselineWins) / races}
}

// summaries returns the result of each skill that can activate on some course. A course where a
// skill cannot activate counts as races without an activation or a change. The changes are 0 before
// the first batch finishes.
func (s *skillSurvey) summaries() []SkillSummary {
	summaries := make([]SkillSummary, 0, len(s.tallies))
	watched := float64(SkillSurveyRaces * s.courseCount)
	for i := range s.tallies {
		tally := &s.tallies[i]
		if !tally.possible {
			continue
		}
		measured := float64(max(tally.racesPerCourse, 1) * s.courseCount)
		summaries = append(summaries, SkillSummary{
			SkillId:         int(s.skillIds[i]),
			ReplacedSkillId: int(s.replacedIds[i]),
			SkillPoints:     tally.skillPoints,
			RacesPerCourse:  tally.racesPerCourse,
			ProcRate:        float64(tally.activated) / watched,
			WinDelta:        float64(tally.winDelta) / measured,
			ScoreDelta:      float64(tally.scoreDelta) / measured,
		})
	}
	return summaries
}
