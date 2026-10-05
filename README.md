# chron

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/hydronica/chron)
[![CI](https://github.com/hydronica/chron/actions/workflows/ci.yml/badge.svg)](https://github.com/hydronica/chron/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/hydronica/chron/graph/badge.svg)](https://codecov.io/gh/hydronica/chron)

it's time :]

<img src="chron.png">

Chron is a UTC-first Go time library that wraps `time.Time` with precision metadata,
calendar-aware durations, and closed intervals. Use it where instants, buckets, and
fuzzy calendar math need to stay distinct — without leaving the stdlib ecosystem.

## Background

The original [dustinevan/chron](https://github.com/dustinevan/chron) tackled a familiar
problem: `time.Time` is used for everything — event timestamps, billing months, hourly
rollups, holiday dates — each with different precision and comparison semantics. The
first design split time into three interfaces (`chron.Time`, `dura.Time`, `chron.Span`)
with nine precision structs (`Year` … `Chron`) and cross-conversion methods.

This keeps the motivation but simplifies the model: four concrete types you can grep
and reason about — no interface lattice, no separate `dura` package.

| Type | Role |
|------|------|
| `Chron` | An instant with optional `Precision` metadata |
| `Duration` | Calendar + clock offsets (years, months, weeks, hours, …) |
| `Span` | Closed interval `[Start, End]` for buckets and containment |
| `Precision` | Truncation, serialization, and span granularity |

Chron still embeds `time.Time` for calculations, so accuracy matches the stdlib. All
constructors normalize to UTC. Instant ordering stays on the embedded `Before`,
`After`, `Equal`, and `Compare`. Bucket membership uses `Span`, not point comparison.

## Installation

```bash
go get github.com/hydronica/chron
```

Requires **Go 1.23+**.

## Quick start

### Chron — instants, truncate, end of unit, add

```go
import "github.com/hydronica/chron"

now := chron.Now()
feb := chron.Date(2026, 2, 1)
startOfMonth := now.Truncate(chron.Month)
endOfMonth := now.EndOf(chron.Month)
next := now.Add(chron.Months(1).Days(3).Hours(4))
parsed, _ := chron.Parse("2026-02")
```

`Truncate(p)` returns the inclusive start of the unit and stores `p` (including
`MondayWeek` and `SundayWeek`). `EndOf(p)` returns the last nanosecond of that unit
and keeps the receiver's precision. `Add` and `Sub` apply a `Duration` with
`AddDate` for calendar fields, then `Add` for clock fields.

`Add` sets the result precision to the finest unit in the offset. A `Nanosecond`
instant stays `Nanosecond` for clock-only offsets (`Hours`, `Minutes`, …). Calendar
fields still change precision (`Months(1)` on a full instant yields `Month`).

### Duration — calendar math and ISO 8601

```go
d := chron.Years(1).Months(3).Hours(12)
d, _ = chron.ParseDuration("p1y3m12h")
d.String() // "p1y3m12h"
```

`Duration` unifies what stdlib splits across `AddDate` and `Add`. `Weeks(n)` is
`n * 7` calendar days. A fixed 168-hour offset is `Clock` or `Hours(168)`.

`ParseDuration` returns an error. `MustDuration` parses the same grammar and panics
on invalid input. Accepted input is case-insensitive ISO 8601 plus chron extensions:
optional leading `p`, bare `14d` or `12h`, `ms` / `us` / `µs` / `ns`, mixed `P1Y1W`,
and fractional years (`P1.5Y` → 18 months). Fractional months and days are rejected.
Marshal is lowercase and canonical (`Days(14)` and `Weeks(2)` both stringify as
`p2w`; microseconds marshal as `µs`). Zero stringifies to `""` and marshals as JSON
`null`. `omitempty` needs a `*Duration`.

### Span — closed intervals for SQL, billing, and iteration

```go
window := expiry.Span(expiry.Precision()) // [Truncate, EndOf] inclusive
db.Query(`WHERE at >= $1 AND at <= $2`, window.Start.AsTime(), window.End.AsTime())

if window.Contains(event) { /* membership, not point ordering */ }
if window.Overlaps(other) { /* scheduling conflicts */ }

for day := range chron.NewSpan(start, end).Each(chron.Days(1)) {
    // inclusive walk; reverse spans walk Start → End
}
```

`Span` is closed: both `Start` and `End` are included. Point (`Start == End`) and
reverse (`Start` after `End`) spans are valid. `NewSpan` always succeeds.
`SpanBetween` returns chronological bounds and drops direction. `Each` walks from
`Start` toward `End` with a positive step; a zero step yields nothing.

`Contains` and `Overlaps` use sorted bounds. Sharing an endpoint counts as overlap.
`Adjacent` means contiguous without overlap (`hi + 1ns == lo`). `ContainsSpan`
checks that one span sits fully inside another. Gapless calendar buckets come from
`Chron.Span(p)` (last nanosecond of January plus one nanosecond is February 1).

### Precision and weeks

```go
chron.DefaultWeekStart = chron.MondayWeek // ISO week (default)

monday := now.Truncate(chron.MondayWeek)
sunday := now.Truncate(chron.SundayWeek)
week := now.Truncate(chron.Week) // uses DefaultWeekStart
```

`Week` uses `DefaultWeekStart`. `MondayWeek` and `SundayWeek` select explicit
boundaries and are stored as that precision on `Truncate`. `Weeks(n)` does not
follow week-boundary selectors.

`Parse` and `UnmarshalJSON` try `ParseFormats` in order and set precision from the
layout that matches. `ParseFrom` uses one layout. Coarser inputs anchor at the start
of that unit in UTC (`"2026-02"` is February 1). `String`, JSON, and SQL `Value`
emit one canonical form per precision. A zero `Chron` stringifies to `""` and
marshals as JSON `null`; SQL `Value` returns `NULL`.

| Precision | Canonical form |
|-----------|----------------|
| `Year` | `2026` |
| `Month` | `2026-02` |
| `Day` | `2026-02-15` |
| `Hour` | `2026-02-15T14` |
| `Minute`, `Second` | RFC3339, no fraction |
| `Millisecond` and finer, and week precisions | RFC3339Nano |

Append human layouts such as `Jan 2006` to `ParseFormats` during init if you need
them on read. They are never written back.

### Interop

- **`AsTime()`** — pass a `Chron` to any API that expects `time.Time`
- **Embedded `time.Time`** — `Format`, `Unix`, `Year`, `Month`, and other stdlib methods work directly
- **JSON and SQL** — `MarshalJSON` / `UnmarshalJSON`, plus `Scan` and `Value` for `database/sql`

## Out of scope

Chron is not a scheduler, a holiday calendar, or a `time.Duration` replacement for
sleeps and timeouts. There is no date-only type, no `SameYear` / `SameMonth` /
`SameDay` helpers, and no span JSON, `Shift`, `Extend`, or `Intersection`. Use
`Span.Contains` (or compare truncated instants) for calendar equality, and
`time.Time.Format` after `In` for display.

## Time zones

All times belong in UTC until a human needs to see them. Constructors (`Now`, `Date`,
`Parse`, `FromTime`) guarantee UTC internal storage. Use the promoted `In` method
(from embedded `time.Time`) for display only.

## Development

```bash
go test ./...
```

CI runs tests with the race detector on Go 1.23, 1.24, and 1.26 and uploads coverage
to [Codecov](https://codecov.io/gh/hydronica/chron) using the latest stable Go release.

## Issues

Open an issue on GitHub if you have questions, ideas, or bugs to report.
