// Package chron builds on time.Time for calendar arithmetic, fuzzy durations
// (import tools/dura), and time spans (Span, TimeSpan, Interval).
//
// # v2 model
//
// Chron is a type alias for time.Time—use it anywhere a time.Time is required.
// Precision “views” (year, month, day, …) are plain time.Time values produced
// by YearOf, MonthOf, etc. Behavior lives in package-level functions (Increment,
// SpanContains, …) because Go does not allow methods on a type alias to an
// external type.
//
// Migration (v1-style names): AsYear → YearOf; AsMonth → MonthOf; method
// Increment(t) → Increment(t, d); Contains → SpanContains (with InstantSpan
// or SpanOfYear / … as appropriate); AddYears on Chron → AddYearsChron;
// on a year-aligned instant use AddYearsYear, AddMonthsYear, etc.
package chron

import "time"

// Chron is an instant in time with nanosecond resolution; it is identical to
// time.Time for assignment, JSON, and sql integration.
type Chron = time.Time
