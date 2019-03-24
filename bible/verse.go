package bible

const (
	defaultVerse = "001"
)

// Verse is representation
// of verse or verse range if end is set.
// End is expected to be non-zero else (if it's zero)
// it means that it represents only single verse.
type Verse struct {
	// Virtual indexes of verse.
	// If end is set to 0 it means that verse represents
	// single verse.
	start int

	// End value should be always greater than start or 0.
	end int
}

// IsSingle - returns true if end range is
// smaller then start.
func (v *Verse) IsSingle() bool {
	if v.end <= v.start {
		return true
	}
	return false
}

func (v *Verse) IsRange() bool {
	if v.start < v.end {
		return true
	}
	return false
}

func (v *Verse) Start() int {
	return v.start
}

func (v *Verse) End() int {
	return v.end
}

// Label represents string label of verse.
// Example: "001002003" - means book one, chapter two, verse three.
// Each token have reserver 3 digits.
// Verse token could contain alphabet literals like "01a".
type Label string

// GetChapter - returns chapter token with default "001".
func (l Label) GetChapter() string {
	if len(l) >= 6 {
		return string(l[3:6])
	}
	return defaultVerse
}

func (l Label) GetBook() string {
	if len(l) >= 3 {
		return string(l[:3])
	}
	return defaultVerse
}

func (l Label) GetVerse() string {
	if len(l) >= 9 {
		return string(l[6:9])
	}
	return defaultVerse
}
