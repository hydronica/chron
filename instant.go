package chron

import (
	"time"

	"tools/dura"
)

// Increment adds calendar years/months/days and a wall-clock duration (fuzzy + exact).
func Increment(t time.Time, d dura.Time) time.Time {
	return t.AddDate(d.Years(), d.Months(), d.Days()).Add(d.Duration())
}

// Decrement subtracts calendar years/months/days and a wall-clock duration.
func Decrement(t time.Time, d dura.Time) time.Time {
	return t.AddDate(-d.Years(), -d.Months(), -d.Days()).Add(-d.Duration())
}

// AddN adds n nanoseconds in UTC (legacy Chron.AddN).
func AddN(t time.Time, n int) time.Time {
	return TimeOf(t.Add(time.Duration(n)))
}
