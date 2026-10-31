package races

import "iter"

// DistanceType is the team trial category of a round, numbered as in the result files
type DistanceType int

const (
	NoneDistance DistanceType = iota
	Sprint
	Mile
	Medium
	Long
	Dirt
)

// RunStyle is the running style of an uma, numbered as in the result files
type RunStyle int

const (
	NoneStyle RunStyle = iota
	Front
	Pace
	Late
	End
)

// DistanceTypeIter yields each distance type in display order
func DistanceTypeIter() iter.Seq[DistanceType] {
	distances := []DistanceType{Sprint, Mile, Medium, Long, Dirt}
	return func(yield func(DistanceType) bool) {
		for i := range distances {
			if !yield(distances[i]) {
				return
			}
		}
	}
}

// RunStyleIter yields each running style in display order
func RunStyleIter() iter.Seq[RunStyle] {
	runStyles := []RunStyle{Front, Pace, Late, End}
	return func(yield func(RunStyle) bool) {
		for i := range runStyles {
			if !yield(runStyles[i]) {
				return
			}
		}
	}
}

// String returns the display name of the distance type
func (dt DistanceType) String() string {
	switch dt {
	case Sprint:
		return "Sprint"
	case Mile:
		return "Mile"
	case Medium:
		return "Medium"
	case Long:
		return "Long"
	case Dirt:
		return "Dirt"
	}
	return "n/a"
}

// String returns the display name of the running style
func (rs RunStyle) String() string {
	switch rs {
	case Front:
		return "Front"
	case Pace:
		return "Pace"
	case Late:
		return "Late"
	case End:
		return "End"
	}
	return ""
}
