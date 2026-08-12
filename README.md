# chron

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/hydronica/chron)
[![CI](https://github.com/hydronica/chron/actions/workflows/ci.yml/badge.svg)](https://github.com/hydronica/chron/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/hydronica/chron/graph/badge.svg)](https://codecov.io/gh/hydronica/chron)

it's time :]

<img src="chron.png">

Chron is a UTC-first Go time library that wraps `time.Time` with precision metadata,
calendar-aware durations, and half-open intervals. Use it where instants, buckets, and
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
| `Span` | Half-open interval `[Start, End)` for buckets and containment |
| `Precision` | Truncation, serialization, and span granularity |

Design rationale lives in [`chron_v1.md`](chron_v1.md) and [`chron_v2.md`](chron_v2.md).

Chron still embeds `time.Time` for calculations, so accuracy matches the stdlib. All
constructors normalize to UTC.

## Installation

```bash
go get github.com/hydronica/chron
```

Requires **Go 1.22+**. 

## Quick start

### Chron — instants, truncate, add

```go
import "github.com/hydronica/chron"

now := chron.Now()
feb := chron.Date(2026, 2, 1)
startOfMonth := now.Truncate(chron.Month)
next := now.Add(chron.Months(1).Days(3).Hours(4))
parsed, _ := chron.Parse("2026-02")
```

`Truncate(p)` returns the inclusive start of the unit containing the instant and sets
precision on the result. `Add` and `Sub` apply a `Duration` using `AddDate` for calendar
fields and `Add` for clock fields.

### Duration — calendar math and ISO 8601

```go
d := chron.Years(1).Months(3).Hours(12)
d, _ = chron.ParseDuration("p1y3m12h")
d.String() // "p1y3m12h"
```

`Duration` unifies what stdlib splits across `AddDate` and `Add`. Parse accepts ISO 8601
(`PnYnMnDTnHnMnS`) plus chron extensions (bare `14d`, `ms`/`µs`/`ns`). Zero durations
marshal as JSON `null` and stringify to `""`.

### Span — intervals for SQL, billing, and overlap

```go
window := expiry.Span(expiry.Precision()) // [Start, End)
db.Query(`WHERE at >= $1 AND at < $2`, window.Start.AsTime(), window.End.AsTime())

if window.Contains(event) { /* membership, not point ordering */ }
if window.Overlaps(other) { /* scheduling conflicts */ }
```

`Span` is half-open: `End` is exclusive. Build arbitrary ranges with `NewSpan` /
`MustSpan`, or derive a calendar bucket from any instant via `c.Span(p)`.

### Precision and weeks

```go
chron.DefaultWeekStart = chron.MondayWeek // ISO week (default)

monday := now.Truncate(chron.MondayWeek)
sunday := now.Truncate(chron.SundayWeek)
week := now.Truncate(chron.Week) // uses DefaultWeekStart
```

`MondayWeek` and `SundayWeek` select explicit week boundaries; stored precision
normalizes to `Week`.

### Interop

- **`AsTime()`** — pass a `Chron` to any API that expects `time.Time`
- **Embedded `time.Time`** — `Format`, `Unix`, `Year`, `Month`, and other stdlib methods work directly; use `Span.Contains` for interval membership instead of overloading `Before`/`After`
- **`Parse` / `ParseFormats`** — registered layouts set precision from the match
- **JSON** — `MarshalJSON` encodes a precision-aware string, or `null` when zero
- **SQL** — `Scan` and `Value` for `database/sql` round-trips

## Time zones

All times belong in UTC until a human needs to see them. Constructors (`Now`, `Date`,
`Parse`, `FromTime`) guarantee UTC internal storage. Use `InLocation` for display only.

## Development

```bash
go test ./...
```

CI runs tests with the race detector on every even go version starting with  Go 1.22 and uploads coverage
to [Codecov](https://codecov.io/gh/hydronica/chron) using the latest stable Go release.

## Issues

Open an issue on GitHub if you have questions, ideas, or bugs to report.
