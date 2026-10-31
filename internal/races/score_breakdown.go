package races

// ScoreCategory groups the score events an uma earns in a team trial race by what earned them
type ScoreCategory int

// Score categories in display order: the bonuses, then the race result, skills and running
const (
	// Bonus points every event earns for the opponent team's rating
	OpponentRatingScore ScoreCategory = iota
	// Bonus points the events of the team's ace earn
	AceScore
	// Bonus points every event earns during a win streak
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
	// Score events whose ID the categories predate
	OtherScore
	ScoreCategoryCount
)

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

// scoreCategoryOf returns the category of a ScoreEvent.RawScoreId, an ID of master.mdb's
// team_stadium_raw_score, where each category spans a contiguous run of IDs
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

// bonusCategoryOf returns the category of a ScoreBonus.ScoreBonusId, an ID of master.mdb's
// team_stadium_score_bonus, or false for a bonus that stays in its event's category
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

// ScoreBreakdown holds points by ScoreCategory. Each event category includes the support bonus
// applied to its events, while the other bonuses count toward their own categories.
type ScoreBreakdown [ScoreCategoryCount]int

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
