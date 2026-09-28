package chron

import (
	"strings"
	"time"
)

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
// Iota order runs fine → coarse (Nanosecond … Year), so finer means smaller enum value.
func (p Precision) Less(p2 Precision) bool {
	return p < p2
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
	case MicroSecond:
		return Micros(1)
	case Millisecond:
		return Millis(1)
	default: // case Nanosecond
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
	case "2006-01-02T15:04:05":
		return Second
	case time.RFC3339:
		return Second
	case time.RFC3339Nano:
		return Nanosecond
	default:
		return precisionFromLayoutComponents(layout)
	}
}

func precisionFromLayoutComponents(layout string) Precision {
	hasYear := strings.Contains(layout, "2006")
	hasMonth := strings.Contains(layout, "01")
	hasDay := strings.Contains(layout, "02")
	hasHour := strings.Contains(layout, "15")
	hasMinute := strings.Contains(layout, "04")
	hasSecond := strings.Contains(layout, "05")
	hasFrac := strings.Contains(layout, "9") || strings.Contains(layout, "0")

	switch {
	case hasFrac:
		return Nanosecond
	case hasSecond:
		return Second
	case hasMinute:
		return Minute
	case hasHour:
		return Hour
	case hasDay:
		return Day
	case hasMonth:
		return Month
	case hasYear:
		return Year
	default:
		return Nanosecond
	}
}

func layoutForPrecision(p Precision) string {
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
	case Millisecond, MicroSecond, Nanosecond, Week, MondayWeek, SundayWeek:
		return time.RFC3339Nano
	default:
		return time.RFC3339Nano
	}
}
