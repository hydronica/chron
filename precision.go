package chron

// Precision describes truncation, serialization, and span bucket granularity.
// Week uses DefaultWeekStart at operation time; MondayWeek and SundayWeek
// select explicit week boundaries.
type Precision uint8

const (
	Nanosecond Precision = iota
	Millisecond
	MicroSecond
	Second
	Minute
	Hour
	Day
	MondayWeek // ISO week (Mon 00:00 UTC → next Mon)
	SundayWeek // US-style week (Sun 00:00 UTC → next Sun)
	Week       // Span/Truncate: uses DefaultWeekStart
	Month
	Year
)

// DefaultWeekStart is the week boundary used when the argument is Week.
var DefaultWeekStart = MondayWeek

// Less reports whether p is a finer (smaller) unit than p2.
// Iota order runs coarse → fine (Year … Nanosecond), so finer means larger enum value.
func (p Precision) Less(p2 Precision) bool {
	return p > p2
}

func (p Precision) Duration() Duration {
	switch p {
	case Year:
		return Years(1)
	case Month:
		return Months(1)
	case Week, MondayWeek, SundayWeek:
		return Weeks(1)
	case Day:
		return Days(1)
	case Minute:
		return Minutes(1)
	case Second:
		return Seconds(1)
	case MicroSecond:
		return Micros(1)
	case Millisecond:
		return Millis(1)
	default: // case NanoSecond
		return Nanos(1)
	}
}

// weekBoundary returns the week-boundary selector for span/truncate operations.
func weekBoundary(p Precision) Precision {
	switch p {
	case MondayWeek, SundayWeek:
		return p
	case Week:
		return DefaultWeekStart
	default:
		return p
	}
}
