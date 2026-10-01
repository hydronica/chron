package chron

import (
	"time"
)

// Precision describes truncation, serialization, and span bucket granularity.
// Week uses DefaultWeekStart at operation time; MondayWeek and SundayWeek
// select explicit week boundaries.
type Precision uint8

const (
	Nanosecond Precision = iota
	Microsecond
	Millisecond
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
// Mutate only during package init, before concurrent Truncate/Span calls.
var DefaultWeekStart = MondayWeek

// Less reports whether p is a finer (smaller) unit than p2.
// Iota order runs fine → coarse (Nanosecond … Year), so finer means smaller enum value.
func (p Precision) Less(p2 Precision) bool {
	return p < p2
}

func (p Precision) String() string {
	switch p {
	case Nanosecond:
		return "nanosecond"
	case Microsecond:
		return "microsecond"
	case Millisecond:
		return "millisecond"
	case Second:
		return "second"
	case Minute:
		return "minute"
	case Hour:
		return "hour"
	case Day:
		return "day"
	case MondayWeek:
		return "monday_week"
	case SundayWeek:
		return "sunday_week"
	case Week:
		return "week"
	case Month:
		return "month"
	case Year:
		return "year"
	default:
		return ""
	}
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
	case Hour:
		return Hours(1)
	case Minute:
		return Minutes(1)
	case Second:
		return Seconds(1)
	case Millisecond:
		return Millis(1)
	case Microsecond:
		return Micros(1)
	default: // case Nanosecond
		return Nanos(1)
	}
}

// weekBoundary returns the week-boundary selector for span/truncate operations.
func (p Precision) weekBoundary() Precision {
	switch p {
	case MondayWeek, SundayWeek:
		return p
	case Week:
		return DefaultWeekStart
	default:
		return p
	}
}

func precisionFromLayout(layout string) Precision {
	switch layout {
	case "2006":
		return Year
	case "2006-01":
		return Month
	case "2006-01-02":
		return Day
	case "2006-01-02T15":
		return Hour
	case "2006-01-02T15:04":
		return Minute
	case time.RFC3339, "2006-01-02T15:04:05":
		return Second
	default: // time.RFC3339Nano also covers unrecognized / custom layouts
		return Nanosecond
	}
}

func (p Precision) layout() string {
	switch p {
	case Year:
		return "2006"
	case Month:
		return "2006-01"
	case Day:
		return "2006-01-02"
	case Hour:
		return "2006-01-02T15"
	case Minute, Second:
		return time.RFC3339
	case Millisecond, Microsecond, Nanosecond, Week, MondayWeek, SundayWeek:
		return time.RFC3339Nano
	default: // also covers unrecognized Precision values
		return time.RFC3339Nano
	}
}
