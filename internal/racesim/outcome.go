package racesim

// RunnerOutcome holds what one runner did in one simulated race. The engine library writes this
// layout, so the fields stay in the order of its RunnerOutcome.
type RunnerOutcome struct {
	// Finish order, counted from 0
	FinishOrder int32
	// Finish time in seconds of race simulation time, not the time that the game displays
	FinishTime float32
	// Meters short of the goal when the previous finisher finished, 0 for the winner
	FinishGap float32
	// Start delay in seconds
	StartDelay float32
	// 1 when the start delay cost the first frame of acceleration, otherwise 0
	LateStart int32
	// Meters from the start of the late race to the start of the last spurt, negative without a
	// last spurt
	SpurtDelay float32
	// Target speed of the last spurt in meters per second
	SpurtSpeed float32
	// HP at the start of the race
	StartHP float32
	// HP at the finish
	FinalHP float32
	// Meters short of the goal when HP first ran out, 0 when HP lasted to the goal
	HPOutDistance float32
	// HP that the runner needed after HP ran out
	HPShortfall float32
	// Seconds spent dueling
	DuelTime float32
	// Seconds in downhill mode before and in the late race
	DownhillTimeEarly, DownhillTimeLate float32
	// Seconds in the Pace Up, Pace Up EX, Speed Up and Overtake modes
	PaceUpTime float32
	// Seconds in Pace Down mode
	PaceDownTime float32
	// Meters lost to lane changes and outside lanes on corners
	WorldTransformLoss float32
}

// RunnerSummary holds the averages of one runner over the simulated races. A rate is a share from
// 0 to 1. An average over the races that meet a condition is 0 when no race meets it.
type RunnerSummary struct {
	// Mean place, counted from 1
	Finish float64
	// Rate of races won
	WinRate float64
	// Mean finish time in seconds
	FinishTime float64
	// Mean meters short of the goal when the previous finisher finished, over the races not won
	FinishGap float64
	// Mean start delay in seconds
	StartDelay float64
	// Rate of races with a late start
	LateStartRate float64
	// Rate of races without a last spurt
	NoSpurtRate float64
	// Mean last spurt delay in meters and target speed in meters per second, over the races with
	// a last spurt
	SpurtDelay, SpurtSpeed float64
	// Mean HP at the start of the race
	StartHP float64
	// Mean HP margin at the finish: the HP left, or the negative of the shortfall in a race where
	// HP ran out
	HPMargin float64
	// Rate of races in which HP ran out short of the goal
	HPOutRate float64
	// Mean meters short of the goal when HP ran out, over the races in which it did
	HPOutDistance float64
	// Mean seconds in each mode
	DuelTime, DownhillTimeEarly, DownhillTimeLate, PaceUpTime, PaceDownTime float64
	// Mean meters lost to lane changes and outside lanes on corners
	WorldTransformLoss float64
}

// Summarize averages the outcomes of races, which hold runnerCount elements per race in gate
// order, and returns one summary per runner in gate order
func Summarize(outcomes []RunnerOutcome, runnerCount int) []RunnerSummary {
	summaries := make([]RunnerSummary, runnerCount)
	for i := range outcomes {
		o, s := &outcomes[i], &summaries[i%runnerCount]
		s.Finish += float64(o.FinishOrder + 1)
		s.FinishTime += float64(o.FinishTime)
		s.FinishGap += float64(o.FinishGap)
		s.StartDelay += float64(o.StartDelay)
		s.LateStartRate += float64(o.LateStart)
		s.StartHP += float64(o.StartHP)
		s.DuelTime += float64(o.DuelTime)
		s.DownhillTimeEarly += float64(o.DownhillTimeEarly)
		s.DownhillTimeLate += float64(o.DownhillTimeLate)
		s.PaceUpTime += float64(o.PaceUpTime)
		s.PaceDownTime += float64(o.PaceDownTime)
		s.WorldTransformLoss += float64(o.WorldTransformLoss)
		if o.FinishOrder == 0 {
			s.WinRate++
		}
		if o.SpurtDelay < 0 {
			s.NoSpurtRate++
		} else {
			s.SpurtDelay += float64(o.SpurtDelay)
			s.SpurtSpeed += float64(o.SpurtSpeed)
		}
		if o.HPOutDistance > 0 {
			s.HPOutRate++
			s.HPOutDistance += float64(o.HPOutDistance)
			s.HPMargin -= float64(o.HPShortfall)
		} else {
			s.HPMargin += float64(o.FinalHP)
		}
	}
	raceCount := float64(len(outcomes) / runnerCount)
	for i := range summaries {
		s := &summaries[i]
		s.FinishGap = meanOver(s.FinishGap, raceCount-s.WinRate)
		s.SpurtDelay = meanOver(s.SpurtDelay, raceCount-s.NoSpurtRate)
		s.SpurtSpeed = meanOver(s.SpurtSpeed, raceCount-s.NoSpurtRate)
		s.HPOutDistance = meanOver(s.HPOutDistance, s.HPOutRate)
		s.Finish /= raceCount
		s.WinRate /= raceCount
		s.FinishTime /= raceCount
		s.StartDelay /= raceCount
		s.LateStartRate /= raceCount
		s.NoSpurtRate /= raceCount
		s.StartHP /= raceCount
		s.HPMargin /= raceCount
		s.HPOutRate /= raceCount
		s.DuelTime /= raceCount
		s.DownhillTimeEarly /= raceCount
		s.DownhillTimeLate /= raceCount
		s.PaceUpTime /= raceCount
		s.PaceDownTime /= raceCount
		s.WorldTransformLoss /= raceCount
	}
	return summaries
}

// meanOver returns the mean of a sum over a count of races, 0 when the count is 0
func meanOver(sum, count float64) float64 {
	if count == 0 {
		return 0
	}
	return sum / count
}
