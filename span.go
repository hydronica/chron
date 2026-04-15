package chron

import (
	"fmt"
	"time"

	"tools/dura"
)

// Span is a half-open range [Start, End] on the time line (inclusive / inclusive).
type Span interface {
	Start() time.Time
	End() time.Time
}

// TimeSpan is a concrete Span with inclusive start and end instants.
type TimeSpan struct {
	S time.Time
	E time.Time
}

func (s TimeSpan) Start() time.Time { return s.S }
func (s TimeSpan) End() time.Time   { return s.E }

// InstantSpan is a degenerate span representing a single instant (nanosecond-wide).
func InstantSpan(t time.Time) TimeSpan {
	tt := InUTC(t)
	return TimeSpan{S: tt, E: tt}
}

// SpanBefore reports whether a ends strictly before b starts.
func SpanBefore(a, b Span) bool {
	return a.End().Before(b.Start())
}

// SpanAfter reports whether a starts strictly after b ends.
func SpanAfter(a, b Span) bool {
	return a.Start().After(b.End())
}

// SpanContains reports whether a overlaps b (legacy Contains on spans).
func SpanContains(a, b Span) bool {
	return !SpanBefore(a, b) && !SpanAfter(a, b)
}

// Contains is SpanContains with receiver a.
func (s TimeSpan) Contains(o Span) bool {
	return SpanContains(s, o)
}

// Interval is a span with an associated fuzzy length d (same value passed to NewInterval).
type Interval struct {
	start time.Time
	end   time.Time
	d     dura.Time
}

// NewInterval builds [start, start+d) closed-open style by backing off one nanosecond at the end,
// matching the previous Chron Interval behavior.
func NewInterval(start time.Time, d dura.Time) Interval {
	start = InUTC(start)
	if d.Duration() == time.Nanosecond && d.Years() == 0 && d.Months() == 0 && d.Days() == 0 {
		return Interval{start: start, end: start, d: dura.Nano}
	}
	end := Decrement(Increment(start, d), dura.Nano)
	return Interval{start: start, end: end, d: d}
}

func (i Interval) Start() time.Time { return i.start }
func (i Interval) End() time.Time     { return i.end }

func (i Interval) Duration() dura.Time { return i.d }

func (i Interval) Contains(o Span) bool {
	return SpanContains(i, o)
}

func (i Interval) Before(o Span) bool {
	return SpanBefore(i, o)
}

func (i Interval) After(o Span) bool {
	return SpanAfter(i, o)
}

func (i Interval) String() string {
	return fmt.Sprintf("start:%s, end:%s, len:%s", i.start, i.end, i.d)
}

// --- calendar bucket spans (legacy Start/End per precision) ---

func SpanOfYear(t time.Time) TimeSpan {
	y := YearOf(t)
	end := Decrement(AddYearsYear(y, 1), dura.Nano)
	return TimeSpan{S: y, E: end}
}

func SpanOfMonth(t time.Time) TimeSpan {
	m := MonthOf(t)
	end := Decrement(AddMonthsMonth(m, 1), dura.Nano)
	return TimeSpan{S: m, E: end}
}

func SpanOfDay(t time.Time) TimeSpan {
	d := DayOf(t)
	end := Decrement(AddDaysDay(d, 1), dura.Nano)
	return TimeSpan{S: d, E: end}
}

func SpanOfHour(t time.Time) TimeSpan {
	h := HourOf(t)
	end := Decrement(h.Add(time.Hour), dura.Nano)
	return TimeSpan{S: h, E: end}
}

func SpanOfMinute(t time.Time) TimeSpan {
	m := MinuteOf(t)
	end := Decrement(m.Add(time.Minute), dura.Nano)
	return TimeSpan{S: m, E: end}
}

func SpanOfSecond(t time.Time) TimeSpan {
	s := SecondOf(t)
	end := Decrement(s.Add(time.Second), dura.Nano)
	return TimeSpan{S: s, E: end}
}

func SpanOfMilli(t time.Time) TimeSpan {
	m := MilliOf(t)
	end := Decrement(m.Add(time.Millisecond), dura.Nano)
	return TimeSpan{S: m, E: end}
}

func SpanOfMicro(t time.Time) TimeSpan {
	m := MicroOf(t)
	end := Decrement(m.Add(time.Microsecond), dura.Nano)
	return TimeSpan{S: m, E: end}
}
