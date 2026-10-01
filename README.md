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

Design rationale lives in [`chron_v1.md`](chron_v1.md) and [`chron_v2.md`](chron_v2.md)
(v2 is authoritative where they disagree).

Chron still embeds `time.Time` for calculations, so accuracy matches the stdlib. All
constructors normalize to UTC.

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

`Truncate(p)` returns the inclusive start of the unit; `EndOf(p)` returns the last
nanosecond of that unit. `Add` and `Sub` apply a `Duration` using `AddDate` for
calendar fields and `Add` for clock fields.

### Duration — calendar math and ISO 8601

```go
d := chron.Years(1).Months(3).Hours(12)
d, _ = chron.ParseDuration("p1y3m12h")
d.String() // "p1y3m12h"
```

`Duration` unifies what stdlib splits across `AddDate` and `Add`. Parse accepts ISO 8601
(`PnYnMnDTnHnMnS`) plus chron extensions (bare `14d`, `ms`/`µs`/`ns`, mixed `P1Y1W`,
fractional `P1.5Y` → 18 months). Zero durations marshal as JSON `null` and stringify
to `""`. Canonical marshal decomposes totals (`Days(14)` → `"p2w"`).

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
reverse (`Start` after `End`) spans are valid. Build ranges with `NewSpan` or
`SpanBetween` (ordered), or derive a calendar bucket via `c.Span(p)`.

### Precision and weeks

```go
chron.DefaultWeekStart = chron.MondayWeek // ISO week (default)

monday := now.Truncate(chron.MondayWeek)
sunday := now.Truncate(chron.SundayWeek)
week := now.Truncate(chron.Week) // uses DefaultWeekStart
```

`MondayWeek` and `SundayWeek` select explicit week boundaries; stored precision on
week selectors follows `Truncate`/`EndOf` input.

### Interop

- **`AsTime()`** — pass a `Chron` to any API that expects `time.Time`
- **Embedded `time.Time`** — `Format`, `Unix`, `Year`, `Month`, and other stdlib methods work directly; use `Span.Contains` for interval membership instead of overloading `Before`/`After`
- **`Parse` / `ParseFormats`** — registered layouts set precision from the match
- **JSON** — `MarshalJSON` encodes a precision-aware string, or `null` when zero
- **SQL** — `Scan` and `Value` for `database/sql` round-trips

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
