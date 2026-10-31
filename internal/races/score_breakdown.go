package races

// ScoreCategory groups score events and score bonuses by what earned them
type ScoreCategory int

// Score categories in display order: the bonuses first, then the score events
const (
	OpponentRatingScore ScoreCategory = iota
	AceScore
	StreakScore
	PlacementScore
	MarginScore
	TargetTimeScore
	LongShotScore
	UniqueSkillScore
	RareSkillScore
	NormalSkillScore
	StrongStartScore
	PositioningScore
	RushedScore
	// Score events whose raw score ID has no category
	OtherScore
	// Number of score categories
	ScoreCategoryCount
)

// String returns the display name of the score category
func (c ScoreCategory) String() string {
	switch c {
	case PlacementScore:
		return "Placement"
	case MarginScore:
		return "Margin"
	case TargetTimeScore:
		return "Target Time"
	case LongShotScore:
		return "Long Shot"
	case UniqueSkillScore:
		return "Unique Skill"
	case RareSkillScore:
		return "Rare Skills"
	case NormalSkillScore:
		return "Normal Skills"
	case StrongStartScore:
		return "Strong Start"
	case PositioningScore:
		return "Positioning"
	case RushedScore:
		return "Rushed"
	case OpponentRatingScore:
		return "Opponent Rating"
	case AceScore:
		return "Ace"
	case StreakScore:
		return "Streak"
	}
	return "Other"
}

// scoreCategoryOf returns the category of a team_stadium_raw_score ID. Each category is a
// contiguous range of IDs.
func scoreCategoryOf(rawScoreId int) ScoreCategory {
	switch {
	case rawScoreId < 1:
		return OtherScore
	case rawScoreId <= 21:
		return MarginScore
	case rawScoreId <= 22:
		return LongShotScore
	case rawScoreId <= 24:
		return RushedScore
	case rawScoreId <= 25:
		return StrongStartScore
	case rawScoreId <= 55:
		return UniqueSkillScore
	case rawScoreId <= 56:
		return RareSkillScore
	case rawScoreId <= 57:
		return NormalSkillScore
	case rawScoreId <= 59:
		return PositioningScore
	case rawScoreId <= 71:
		return PlacementScore
	case rawScoreId <= 91:
		return TargetTimeScore
	}
	return OtherScore
}

// bonusCategoryOf returns the category of a team_stadium_score_bonus ID. The support bonus has
// no category and stays in the category of its score event.
func bonusCategoryOf(scoreBonusId int) (ScoreCategory, bool) {
	switch scoreBonusId {
	case 1:
		return AceScore, true
	case 2:
		return OpponentRatingScore, true
	case 4, 5, 6, 7:
		return StreakScore, true
	}
	return OtherScore, false
}

// ScoreBreakdown holds the points of one uma in each score category
type ScoreBreakdown [ScoreCategoryCount]int

// add adds the points of score events. A bonus with its own category counts there, not in the
// category of its score event.
func (b *ScoreBreakdown) add(events []ScoreEvent) {
	for _, event := range events {
		points := event.Score
		for _, bonus := range event.BonusArray {
			if category, ok := bonusCategoryOf(bonus.ScoreBonusId); ok {
				points -= bonus.Score
				b[category] += bonus.Score
			}
		}
		b[scoreCategoryOf(event.RawScoreId)] += points
	}
}
