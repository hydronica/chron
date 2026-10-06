# chron

Go time values that track calendar precision, support month/year offsets, and check inclusive ranges.

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/hydronica/chron)
[![CI](https://github.com/hydronica/chron/actions/workflows/ci.yml/badge.svg)](https://github.com/hydronica/chron/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/hydronica/chron/graph/badge.svg)](https://codecov.io/gh/hydronica/chron)

<img src="chron.png" alt="chron" width="200">

Requires **Go 1.23+**. Inspired by [dustinevan/chron](https://github.com/dustinevan/chron).

## Installation

```bash
go get github.com/hydronica/chron
```

## Quick start

```go
import "github.com/hydronica/chron"

now := chron.Now()                                      // UTC, nanosecond precision
feb := chron.Date(2026, 2, 1)                           // midnight UTC, day precision
start := now.Truncate(chron.Month)                      // first instant of the month
end := now.EndOf(chron.Month)                           // last nanosecond of the month
next := now.Add(chron.Months(1).Days(3).Hours(4))

parsed, err := chron.Parse("2026-02")                   // month precision → 2026-02-01
if err != nil { /* ... */ }

signup := chron.Date(2026, 2, 1)
renewal := signup.Add(chron.MustDuration("P14D"))       // +14 calendar days

mid := chron.Date(2026, 1, 15)
mid.Add(chron.Months(1))                                // 2026-02-15 (same day next month)
```

## Types

| Type | Role |
|------|------|
| `Chron` | A moment in time with a `Precision` (day, month, …); embeds `time.Time` |
| `Duration` | An offset in years, months, weeks, days, and/or clock time |
| `Span` | A range from `Start` to `End` that includes both ends |
| `Precision` | How coarse the value is — used for truncate, format, and spans |
| `Period` | A named reporting window relative to a `Chron` (month-to-date, …) |

All constructors store UTC. To ask “is this event in February?”, build a month `Span` and call `Contains` — do not use `Before` / `After` alone.

## Chron

```go
c := chron.FromTime(t)   // converts to UTC; nanosecond precision
t = c.AsTime()           // underlying time.Time

c.Truncate(chron.Day)    // start of that day; result precision is Day
c.EndOf(chron.Day)       // last nanosecond of that day; keeps c's precision
c.Add(chron.Months(1))   // months clamped, then days, then clock
c.Sub(chron.Days(7))
c.Span(chron.Month)      // [start of month, end of month], both included
```

`Add` and `Sub` take `chron.Duration` only. Result precision is the finer of the receiver and the duration: day + month stays day (`2026-02-28`); day + seconds becomes second.

Compare two values with the embedded `time.Time` methods — pass `.Time` or `AsTime()`:

```go
a.Before(b.Time)
a.After(b.Time)
a.Equal(b.Time)
```

## Duration

`time.Duration` is a fixed number of nanoseconds. `chron.Duration` is a named offset: months (clamped to the end of the target month), days, and an optional clock part. Use `chron.Duration` when “one month” or “two weeks” should follow the calendar. Use `time.Duration` for sleeps, timeouts, and measuring elapsed time.

| | `time.Duration` | `chron.Duration` |
|--|-----------------|------------------|
| Storage | `int64` nanoseconds | months, days, and a clock `time.Duration` |
| Meaning | Fixed length | Calendar units from a starting `Chron`, plus optional clock |
| Typical use | `time.Sleep`, deadlines | `c.Add(chron.Months(1))`, config like `P1Y3M` |
| “One month” | Not representable | `Months(1)` |
| “One week” | Often `168 * time.Hour` | `Weeks(1)` → 7 calendar days |
| Parse | `time.ParseDuration` | `ParseDuration` / `MustDuration` |

### Construction

```go
d := chron.Years(1).Months(3).Hours(12)
d, err := chron.ParseDuration("P1Y3MT12H")  // T required before clock units
d = chron.MustDuration("P14D")               // panics on error
d = chron.Clock(90 * time.Minute)            // clock only, no calendar fields
```

`Years(n)` stores `n*12` months. `Weeks(n)` stores `n*7` days. You can chain methods on a value: `Years(1).Months(3).Hours(12)`.

Also available: `Neg`, `Mul`, `IsZero`, `Precision`, `String`, JSON marshal/unmarshal.

### How Add applies a duration

On `Chron.Add(d)`:

1. Apply months with end-of-month clamp (not Go’s `AddDate` overflow)
2. Apply days with `AddDate`
3. Apply the clock part with `Add`

`Jan 31` plus one month lands on the last day of February:

```go
chron.Date(2026, 1, 31).Add(chron.Months(1)) // 2026-02-28
chron.Date(2024, 1, 31).Add(chron.Months(1)) // 2024-02-29 (leap year)
```

### Parsing

`ParseDuration` accepts [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) period strings (`P…` / `p…`), with or without a `T` time section. Matching is case-insensitive.

| Input | Result |
|-------|--------|
| `P14D`, `14d` | 14 days |
| `PT12H`, `12h` | 12 hours |
| `P1Y3MT12H` | 1 year, 3 months, 12 hours |
| `P1Y1W` | 1 year and 1 week (chron extension) |
| `P1.5Y` | 18 months |
| `PT1.5S`, `250ms`, `250us` / `250µs`, `100ns` | clock fractions / sub-second units |
| `-P1D` | negative one day |
| `P0D`, `0d` | zero |

Rules:

- If `P` is missing, chron adds it (`14d` → `p14d`). Bare clock strings get `PT` (`12h` → `pt12h`).
- Clock units after calendar units need a `T` (`P1Y3MT12H`, not `P1Y3M12H`).
- Fractional months or days are rejected. Empty input and bare `P` / `PT` error.
- `MustDuration` panics on error. It is not named `Duration` because that name is already the type.

### String and JSON

`String` writes a lowercase form (`p1y2m3dt4h30m5s…`). Microseconds always use `µs`. On output, months split into years + months, and days into weeks + days — so `Days(14)` and `Weeks(2)` both print as `p2w`.

Zero duration: `String()` is `""`; JSON is `null`. For `json:",omitempty"` to skip a missing duration field, use `*chron.Duration`.

### Common mistakes

| Mistake | Instead |
|---------|---------|
| `30 * 24 * time.Hour` as “one month” | `chron.Months(1)` |
| `Hours(168)` for a calendar week | `Weeks(1)` |
| `Weeks(1)` for exactly 168 clock hours | `Hours(168)` or `Clock(7 * 24 * time.Hour)` |
| `time.ParseDuration("P1Y3M")` | `chron.ParseDuration` |
| Passing `chron.Duration` to `time.Sleep` | use `time.Duration` |
| Non-pointer `Duration` with `omitempty` | `*chron.Duration` |
| `ParseDuration("P1Y3M12H")` | `ParseDuration("P1Y3MT12H")` |

## Span

A `Span` includes both endpoints. `Start` may equal `End` (a single instant) or come after `End` (a reverse range). Membership and overlap still use chronological order; `Each` walks from `Start` toward `End`.

```go
window := expiry.Span(expiry.Precision())
db.Query(`WHERE at >= $1 AND at <= $2`, window.Start.AsTime(), window.End.AsTime())

if window.Contains(event) { /* inside the range */ }
if window.Overlaps(other) { /* any shared instant, including touching at an end */ }

for day := range chron.NewSpan(start, end).Each(chron.Days(1)) {
    // visits Start, then each step, through End when it lands on a step
}
```

| Method | Behavior |
|--------|----------|
| `NewSpan` | Builds the span as given (including reverse) |
| `SpanBetween` | Always returns earlier → later; loses reverse direction |
| `Each(step)` | Steps from `Start` toward `End`; zero step yields nothing |
| `Contains` / `Overlaps` | Use sorted ends; touching at one endpoint counts as overlap |
| `Adjacent` | Touching with a one-nanosecond gap, and not overlapping |
| `ContainsSpan` | `other` lies entirely inside this span |
| `Chron.Span(p)` | Full range for that unit (for example, all of February) |
| `Chron.Period(p)` | Named reporting window relative to this instant |

## Reporting spans

`Period` selects a closed reporting window from a `Chron`. To-date windows end at the last nanosecond of that calendar day (`EndOf(Day)`). The zero value is `MonthToDate`, so an unset `Period` is month-to-date.

```go
mtd := now.Period(chron.MonthToDate) // [1st of month, EndOf(Day)]
ytd := now.Period(chron.YearToDate)
prev := now.Period(chron.PrevMonth)  // full previous month

var p chron.Period
_ = json.Unmarshal([]byte(`"prev_year_to_date"`), &p) // same snake_case as Precision
span := now.Period(p)
```

| Constant | Window |
|----------|--------|
| `MonthToDate` | From the 1st of this month through `EndOf(Day)` |
| `YearToDate` | From Jan 1 through `EndOf(Day)` |
| `PrevMonth` | Full previous calendar month |
| `PrevMonthToDate` | Previous month, 1st through the same day (clamped) |
| `PrevYearMonthToDate` | Same month last year, 1st through the same day (clamped) |
| `PrevYearToDate` | Jan 1 last year through the same month/day (clamped) |
| `LastFullWeek` | The completed week before the week that contains this instant |

Prior-period day numbers clamp when the target month is shorter (31 March → 28 February; 29 February → 28 February). `LastFullWeek` follows `DefaultWeekStart`, the same as `Truncate(Week)`.

JSON encodes a `Period` as a snake_case string (`"month_to_date"`). `null` unmarshals to `MonthToDate`. An unknown string fails only in marshal/unmarshal; `Chron.Period` always returns a `Span`.

## Precision

```go
chron.DefaultWeekStart = chron.MondayWeek // default: Monday (ISO-style)

now.Truncate(chron.MondayWeek) // week starting Monday
now.Truncate(chron.SundayWeek) // week starting Sunday
now.Truncate(chron.Week)       // uses DefaultWeekStart
```

`Week` follows `DefaultWeekStart`. `MondayWeek` and `SundayWeek` pick a fixed start day. `Weeks(n)` on a `Duration` always means `n*7` calendar days — it does not align to Monday or Sunday week starts.

`Parse` and JSON unmarshal try `ParseFormats` in order and set precision from the layout that matched. `ParseFrom` uses one layout. A shorter string anchors at the start of that unit in UTC (`"2026-02"` → February 1). You can append layouts such as `Jan 2006` to `ParseFormats` during `init` for reading only; `String` / JSON output still use the forms below.

| Precision | `String` / JSON form |
|-----------|----------------------|
| `Year` | `2026` |
| `Month` | `2026-02` |
| `Day` | `2026-02-15` |
| `Hour` | `2026-02-15T14` |
| `Minute`, `Second` | RFC3339 (no fractional seconds) |
| Millisecond and finer; week values | RFC3339Nano |

A zero `Chron` prints as `""` and marshals as JSON `null`. For SQL, `Value` returns `NULL`.

## Working with time.Time and databases

- `AsTime()` when an API needs `time.Time`
- Promoted methods from the embedded `time.Time`: `Format`, `Unix`, `Year`, `Month`, `In`, …
- JSON: `MarshalJSON` / `UnmarshalJSON` on `Chron` and `Duration`
- SQL: `Scan` / `Value` on `Chron`
- `Now`, `Date`, `Parse`, and `FromTime` always store UTC; call `In` for display in another zone

## Limitations

- No job scheduling or holiday calendars — use `NthWeekday` / `LastWeekday` for nth-weekday math
- Sleeps and timeouts stay on `time.Duration`
- No separate date-only type; no `SameYear` / `SameMonth` / `SameDay` helpers — use `Span.Contains` or compare truncated values
- `Span` has no JSON encoding, and no `Shift`, `Extend`, or `Intersection`

## Development

```bash
go test ./...
```

CI runs race-detector tests on Go 1.23, 1.24, and 1.26 and uploads coverage to [Codecov](https://codecov.io/gh/hydronica/chron).

## Links

- [API docs](https://pkg.go.dev/github.com/hydronica/chron)
- [Issue tracker](https://github.com/hydronica/chron/issues)
