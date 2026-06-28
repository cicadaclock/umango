package races

import "iter"

// I couldn't find these types in the master.mdb, presumably because they are
// calculated at runtime, so we define them here

// DistanceType represents the TT race type, including surface
//
// i.e. Dirts are miles, but are shown as dirt, so that's a unique DistanceType
type DistanceType int

const (
	NoneDistance DistanceType = iota
	Sprint
	Mile
	Medium
	Long
	Dirt
)

type RunStyle int

const (
	NoneStyle RunStyle = iota
	Front
	Pace
	Late
	End
)

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
