package chron

import "time"

// InUTC returns t in UTC (same normalization as the legacy TimeOf).
func InUTC(t time.Time) time.Time {
	return t.UTC()
}

// Preserve returns t unchanged (use when AddDate should respect t.Location()).
func Preserve(t time.Time) time.Time {
	return t
}

// TimeOf is equivalent to InUTC; kept for migration from older chron APIs.
func TimeOf(t time.Time) time.Time {
	return InUTC(t)
}

// Now returns the current instant in UTC.
func Now() time.Time {
	return InUTC(time.Now())
}

// NewTime constructs a UTC instant from calendar fields (month is 1–12 as in time.Month).
func NewTime(year int, month time.Month, day, hour, min, sec, nano int) time.Time {
	return time.Date(year, month, day, hour, min, sec, nano, time.UTC)
}

// ZeroTime is the zero value of time.Time in UTC.
func ZeroTime() time.Time {
	return time.Time{}.UTC()
}

// ZeroValue is the zero instant as a Chron (UTC).
func ZeroValue() time.Time {
	return TimeOf(time.Time{})
}

// ZeroYear is January 1 of calendar year 0 in UTC.
func ZeroYear() time.Time {
	return NewYear(0)
}

// ZeroUnix is the Unix epoch in UTC.
func ZeroUnix() time.Time {
	return TimeOf(time.Unix(0, 0))
}

// see: https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go
// and time.Unix() implementation
var unixToInternal = int64((1969*365 + 1969/4 - 1969/100 + 1969/400) * 24 * 60 * 60)

var maxTime = time.Unix(1<<63-1-unixToInternal, 999999999).UTC()
var minTime = time.Unix(-1*int64(^uint(0)>>1)-1+unixToInternal, 0).UTC()

// MaxValue is the latest representable instant in UTC.
func MaxValue() time.Time {
	return TimeOf(maxTime)
}

// MinValue is the earliest representable instant in UTC.
func MinValue() time.Time {
	return TimeOf(minTime)
}
