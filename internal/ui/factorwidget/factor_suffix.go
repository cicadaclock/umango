package factorwidget

// SuffixType represents how a factor_widget should appear
type SuffixType int

const (
	StarSuffix SuffixType = iota // Show factor suffixes as stars
	IntSuffix                    // Show factor suffixes as integers
	SizeSuffix
)

func (ft SuffixType) Int() int {
	return int(ft)
}
