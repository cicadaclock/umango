package races

import "iter"

type DistanceType int

const (
	_                   = iota
	Sprint DistanceType = iota
	Mile
	Medium
	Long
	Dirt
)

type RunStyle int

const (
	_              = iota
	Front RunStyle = iota
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
	return ""
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
