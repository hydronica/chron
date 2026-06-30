# chron v1 — `chron.Chron` and `chron.Span` design

Design notes for a Go time helper library. This document covers the two core types:

- **`chron.Chron`** — a thin wrapper around `time.Time` (instants, with optional precision metadata)
- **`chron.Span`** — a concrete start/end interval for calendar buckets and containment checks
- **`chron.Duration`** — what to add: calendar units (month, week) and clock units (hour), applied correctly per anchor

Precision helpers and parsing are covered inline below.

**v1 is interfaceless.** No `chron.Time` or `chron.Span` interface. Behavior lives on
concrete structs and package-level functions.

---

## What problem are we solving?

Go gives us one type — `time.Time` — for many different meanings of "time." In practice
we use it for all of the following, often interchangeably:

| Use case | What we mean | Typical precision | Comparison semantics |
|----------|--------------|-------------------|---------------------|
| Event log timestamp | An instant | nanoseconds | Point ordering |
| Postgres `timestamptz` | An instant (stored UTC) | microseconds | Point ordering |
| "Today's report" | A calendar day | day | Interval overlap |
| Credit card expiry | A month on a calendar | month | Interval overlap |
| Hourly rollup bucket | A clock hour | hour | Interval overlap |
| Holiday / business day | A named calendar date | day | Interval overlap |
| "Same time tomorrow" | Calendar + clock arithmetic | mixed | Needs `AddDate`, not just `Add` |

`time.Time` handles the first row well. Everything else works, but only if the developer
remembers implicit rules: truncate before comparing, use `AddDate` for months, always
normalize to UTC before storage, pick the right parse format for each API.

Those rules are easy to get wrong and hard to see in code review.

### Concrete failure modes

**1. Point comparison when interval comparison was intended**

```go
feb := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
if event.Before(feb) { /* ... */ }
```

If `event` is `2026-02-01 08:00 UTC`, this is false — even though the event clearly
falls *within* February. The developer meant "is this event in February?" but wrote
instant ordering.

**2. Calendar arithmetic split across two APIs**

```go
t.AddDate(1, 3, 0)           // years, months, days — calendar-aware
t.Add(30 * 24 * time.Hour)   // fixed duration — not the same as "1 month"
```

There is no single value you can pass around that means "1 year, 3 months, and 4 hours
from anchor *t*." Config, UI copy, and billing logic often express durations this way.

**3. Precision is invisible in the type**

```go
type Subscription struct {
    ExpiresAt time.Time // month precision? day? instant?
}
```

Nothing in the type tells callers whether `ExpiresAt` is "end of Feb 2026" or
"2026-02-28T23:59:59Z" or "first instant of March." Serialization and DB round-trips
silently preserve or destroy that intent.

**4. Parse/format mismatch**

`time.RFC3339` is the JSON default. Real inputs include `"2026-02"`, `"02/2026"`,
`"2026-02-04 15:04:05"`, and database driver formats. Each caller rebuilds ad-hoc
parsers.

**5. Timezone confusion at boundaries**

Storing local-midnight-as-UTC, or comparing times in different locations, produces
off-by-one-day bugs around calendar boundaries. The fix is a convention (store UTC,
convert for display), but nothing enforces it.

---

## Design goals for v1

1. **Interop first.** `chron.Chron` wraps `time.Time`; it does not replace it. Any
   function that accepts `time.Time` should accept `chron.Chron` via `AsTime()` or
   implicit conversion patterns the ecosystem already uses.

2. **Explicit precision where it prevents bugs.** Not a lattice of nine struct types —
   a single wrapper that can *carry* precision metadata and offer truncation helpers.

3. **One way to add time.** A single `c.Add(d Duration)` replaces the stdlib split
   between `AddDate` and `Add`. `chron.Duration` names *what* you mean (a month, a week,
   an hour); the library picks the right implementation for the anchor instant.

4. **`Chron` is the receiver.** If `Chron` is the subject of the operation, it is a
   method on `Chron` — mirroring `time.Time.Add`. Package functions are for constructors
   (`Now`, `Parse`), `Duration` builders (`Months`, `Years`), and cases with no natural
   receiver (`NewSpan`).

5. **Intervals are first-class.** Calendar buckets ("the month of February", "the week
   containing this date") are represented as a concrete `Span` with explicit `Start` and
   `End`, not inferred from a truncated instant alone.

6. **Interval comparisons stay explicit.** Embedded `time.Time` keeps stdlib `Before` /
   `After` / `Equal` / `Compare` for instant ordering. Precision does not change those
   semantics. Use `Span.Contains`, `Span.Overlaps`, and calendar helpers (`SameMonth`,
   …) for buckets and membership — never overload the stdlib names.

7. **UTC in, UTC stored.** Constructors normalize to UTC. Display conversion is
   explicit (`InLocation`).

8. **Easy by default.** Common cases are one line. Escape hatches (`AsTime`, raw
   `time.Time` embed) are always available.

9. **Interfaceless.** Concrete types and functions only. The original chron library
   split time into three interfaces (`chron.Time`, `dura.Time`, `chron.Span`) with
   many implementations and cross-conversion methods. v1 collapses that into structs
   you can see and grep.

---

## What is `chron.Chron`?

```go
// Chron is an instant in time with nanosecond storage precision.
// It embeds time.Time and adds chron-specific behavior.
type Chron struct {
    time.Time
    // precision is optional metadata: how this value should be truncated,
    // compared, and serialized. Zero value = Nanosecond (full instant).
    precision Precision
}
```

`FromTime` and `Now` set `Nanosecond` precision by default — a converted `time.Time` is
treated as a full instant unless truncated or parsed with coarser intent.

`chron.Chron` is the **default instant type** for the library — the v1 equivalent of
the original `chron.Chron` nanosecond wrapper, simplified. It is **not** a replacement
for every `time.Time` in your codebase; use it at boundaries where precision, parsing,
or calendar logic matters.

### Type model (no interfaces)

The original library used three interfaces — `chron.Time`, `dura.Time`, `chron.Span` —
with nine precision structs (`Year` … `Chron`) each implementing conversion to every
other type. v1 replaces that lattice with a small set of concrete types:

| Type | Kind | Role |
|------|------|------|
| `Chron` | struct | Instants; optional `Precision` metadata for truncation and serialization |
| `Span` | struct | Half-open interval `[Start, End)` — the month, week, day, or arbitrary range |
| `Duration` | struct | Calendar + clock offset to apply via `Add` / `Sub` |
| `Precision` | enum (`uint8`) | Truncation, serialization, and `Span(p)` buckets. `MondayWeek` / `SundayWeek` are input-only week-boundary selectors; stored precision normalizes to `Week` |
| `Unit` | enum | Named unit inside `Duration` (`Month`, `Week`, `Hour`, …) — distinct from `Precision` |

There is no `chron.Time` interface. Callers work with `Chron` and `Span` directly.

**`Chron` vs `Span`.** A `Chron` is a point (or anchor). A `Span` is the full window
around that anchor when you care about start *and* end — "all of February 2026", "the
ISO week containing this timestamp", "business hours today". You can pass a `Span` to
functions, store it in structs, and ask `Contains` without re-deriving boundaries every
time.

**Chron methods vs package functions.** Operations on an instant are methods:
`c.Truncate(chron.Day)`, `c.Add(chron.Months(1))`, `c.Span(chron.Month)`. Package functions are
reserved for building values without a receiver (`Now`, `Parse`, `Months`, `NewSpan`).

If we later need a distinct struct (e.g. `Date` with no clock component), it will be
another concrete type with explicit conversion to/from `Chron` and `Span`, not a shared
interface.

---

## What `chron.Chron` implements

### Construction

| Function | Purpose |
|----------|---------|
| `Now() Chron` | Current instant, UTC |
| `FromTime(t time.Time) Chron` | Wrap existing `time.Time`, normalize to UTC |
| `Date(y, m, d int) Chron` | Calendar date at midnight UTC, precision `Day` |
| `Parse(s string) (Chron, error)` | Try registered layouts; set `precision` from match (see Parsing) |
| `ParseFrom(layout, s string) (Chron, error)` | Parse with explicit layout; set `precision` from layout |

**Explicit-layout name:** `ParseFrom` is the v1 name for the layout-parameter variant.
Alternatives if renamed later: `ParseLayout`, `ParseWithLayout`, `ParseInLayout`,
`ParseExact`.

Constructors always store UTC. Passing a `time.Time` in another location converts via
`.UTC()` unless a future `FromTimeIn(loc)` variant is needed.

### Escape hatch

| Method | Purpose |
|--------|---------|
| `AsTime() time.Time` | Return underlying stdlib value for third-party APIs |
| `IsZero() bool` | Invalid / unset — `true` when embedded `time.Time` is zero |
| Embedded `time.Time` methods | `Format`, `Unix`, `Year`, `Month`, … work as today |

### Truncation and precision

`Truncate(p Precision)` is the **only** way to set precision by calendar intent. It
returns the **inclusive start** of the unit containing the receiver (UTC) and sets
`precision` on the result. There is no separate `WithPrecision` — relabeling without
moving the instant would disagree with parse anchors and `Span(c.Precision())`.

**Week-boundary inputs.** `Truncate(Week)` uses `DefaultWeekStart`. `Truncate(MondayWeek)`
and `Truncate(SundayWeek)` use explicit ISO Monday or US Sunday boundaries. All three
store **`Week` precision** on the result — `MondayWeek` / `SundayWeek` are never retained
on `Chron`; they select boundaries only (same normalization for `Span(p).Start`).

| Method | Purpose |
|--------|---------|
| `Truncate(p Precision) Chron` | Start of unit for `p`; sets `precision` on result (week inputs → `Week`) |
| `Precision() Precision` | Read current precision metadata (set by parse, `Truncate`, or `Add`) |

`Truncate(chron.Month)` is equivalent to `c.Span(chron.Month).Start`.
`Truncate(chron.SundayWeek)` is equivalent to `c.Span(chron.SundayWeek).Start` — both
yield `Week` precision on the truncated instant.

Examples: `Truncate(Year)`, `Truncate(Month)`, `Truncate(Week)`, `Truncate(SundayWeek)`.

See [Span derivation](#span-derivation) for week boundary details.

### Add and subtract

One entry point for all offset arithmetic, like `time.Time.Add` but with `chron.Duration`.
**`c.Sub(d)`** is `c.Add(d.Neg())`.

| Method | Purpose |
|--------|---------|
| `Add(d Duration) Chron` | Apply `d` relative to this instant |
| `Sub(d Duration) Chron` | Subtract `d` from this instant |

```go
next := c.Add(chron.Months(1))              // calendar month (AddDate)
inTwoWeeks := c.Add(chron.Weeks(2))         // calendar weeks
later := c.Add(chron.Hours(4))              // fixed 4 hours (time.Add)
renewal := c.Add(chron.Years(1).Months(3).Hours(12))
```

See [What is `chron.Duration`?](#what-is-chronduration) for how each unit is applied.

### Precision after arithmetic

`Add` and `Sub` always return a **new** `Chron`; the receiver is unchanged (same as
`time.Time`). The returned value carries an updated `precision` that reflects what the
caller is now expressing — not a blind copy of the receiver's precision.

**Rule: result precision = finest unit in the applied offset**, stepping through the
full ladder: `Year` → `Month` → `Week` → `Day` → `Hour` → `Minute` → `Second` →
`Nanosecond`.

The instant moves via `AddDate` / `Add`; precision on the result matches the smallest
unit in `d`, because that is what the user signaled they care about.

| Receiver precision | Offset | Result instant | Result precision |
|--------------------|--------|----------------|------------------|
| `Month` (Feb 1 anchor) | `Days(5)` | Feb 6 | `Day` |
| `Month` | `Months(1)` | Mar 1 | `Month` |
| `Day` | `Hours(3)` | same day + 3h | `Hour` |
| `Hour` | `Min(15)` | +15 minutes | `Minute` |
| `Minute` | `Sec(30)` | +30 seconds | `Second` |
| `Nanosecond` | `Hours(2)` | +2 hours | `Nanosecond` (see below) |
| `Nanosecond` | `Months(1)` | +1 calendar month | `Month` |

**Nanosecond floor for clock offsets.** Values at `Nanosecond` precision — the default
for `FromTime`, `Now`, and any full instant — stay `Nanosecond` when the offset is
**clock-only** (`Hours`, `Min`, `Sec`, `Clock`). Adding two hours to an event
timestamp is still an event timestamp, not an "hour bucket." Calendar fields in `d`
(`Years`, `Months`, `Weeks`, `Days`) still update precision normally.

```go
event := chron.FromTime(loggedAt)     // Nanosecond (default)
event.Precision()                     // Nanosecond
event.Add(chron.Hours(2)).Precision() // Nanosecond — still a full instant

day := chron.Date(2026, 6, 1)         // Day
day.Add(chron.Hours(3)).Precision()   // Hour — caller moved to hour granularity
```

`c.Truncate(chron.Month).Add(chron.Days(5))` is the common case: truncate yields Jan 1
with `Month` precision; adding five days yields **Jan 6 with `Day` precision**. No
conflict — month precision on the receiver describes the starting intent; day precision
on the result describes the new intent.

**Month-precision anchor + one day.** If `c` is `"2026-02"` (`Month` precision, anchor
Feb 1) and the caller writes `c.Add(chron.Days(1))`, they mean a **specific calendar
day**, not "still the whole month." The result is Feb 2 with `Day` precision — not
February with `Month` precision.

**Composite offsets** (`Years(1).Months(3).Days(5).Hours(12)`): finest unit in the
composite wins (`Hour` here). Calendar fields still apply in full before clock. If the
receiver is `Nanosecond` and the composite includes only clock units, result stays
`Nanosecond`.

**`Truncate`** set precision directly from the operation; they do not use the finest-unit
rule below. **`Add` / `Sub`** update precision from the offset.

```go
jan := chron.Date(2026, 1, 15).Truncate(chron.Month) // Jan 1, Month
jan.Precision()                                      // Month
sixth := jan.Add(chron.Days(5))                      // Jan 6, Day
sixth.Precision()                                    // Day — original jan unchanged
```

### Span derivation

Build the calendar bucket **containing** this instant. Returns `[start, end)` as `Span`.
One entry point — **`Span(p Precision)`** — no separate interface or `WeekStart` type.

```go
type Precision uint8

const (
    Year Precision = iota
    Month
    Week          // Span/Truncate: uses DefaultWeekStart
    Day
    Hour
    Minute
    Second
    Millisecond
    Nanosecond

    // Week-boundary selectors — valid input to Truncate/Span; stored precision → Week
    MondayWeek    // ISO week (Mon 00:00 UTC → next Mon), explicit
    SundayWeek    // US-style week (Sun 00:00 UTC → next Sun), explicit
)

var DefaultWeekStart = MondayWeek // Precision value; used when arg is Week
```

| Call | Result |
|------|--------|
| `c.Span(chron.Year)` | `[Jan 1, Jan 1 next year)` |
| `c.Span(chron.Month)` | `[1st 00:00, 1st next month)` |
| `c.Span(chron.Week)` | Week containing `c`; uses `DefaultWeekStart` |
| `c.Span(chron.MondayWeek)` | ISO week, explicit (even if default is Sunday) |
| `c.Span(chron.SundayWeek)` | US-style week, explicit |
| `c.Span(chron.Day)` | `[midnight, midnight next day)` |
| `c.Span(chron.Hour)` | `[hour boundary, next hour)` |
| `c.Span(c.Precision())` | Bucket at this value's declared precision (e.g. parsed `"2026-02"`). If precision is `Week`, uses `DefaultWeekStart` |

`Span(p).Start` and `Truncate(p)` share the same boundary rules and the same precision
normalization: week-boundary selectors store `Week` on the resulting `Chron`.

`Span` stores only `Start` and `End`. Week boundary choice is input at construction time;
the interval is self-contained afterward.

**Why one enum, no interface.** `SpanUnit` existed only so Go could accept two types in
one method. A single `Precision` enum with week-boundary selectors at the end is simpler,
matches the interfaceless v1 goal, and keeps `c.Span(chron.SundayWeek)` without a wrapper
type. `Duration` keeps its own `Unit` enum — different concern (offsets, not buckets).

**Stored vs input precision.** Values stored on `Chron` (`parse`, `Truncate`, `Add`) are
always `Year` … `Nanosecond`. `MondayWeek` / `SundayWeek` are input-only aliases that
normalize to `Week` when truncating or when taking `Span(p).Start`.

#### Week boundaries

| Mode | Constant | Week span `[Start, End)` |
|------|----------|--------------------------|
| Default / explicit ISO | `Week` (via `DefaultWeekStart`) or `MondayWeek` | Mon 00:00 UTC → next Mon |
| Explicit US | `SundayWeek` | Sun 00:00 UTC → next Sun |

**Where week start is chosen**

| Mechanism | Scope | Use when |
|-----------|-------|----------|
| `DefaultWeekStart` | Package / process | App-wide default for `Span(Week)` / `Truncate(Week)` |
| `c.Span(chron.SundayWeek)` | Single call | Explicit boundaries; per-user locale from app state |
| App stores `Precision` on user/settings | Domain model | e.g. `SundayWeek`; pass to `Truncate(p)` or `Span(p)` — not on `Chron` |

```go
// Default ISO week (Monday) via Week + DefaultWeekStart
anchor := chron.Now().Truncate(chron.Week)   // precision → Week

// Explicit US week — boundary from SundayWeek, precision stored as Week
us := chron.Now().Truncate(chron.SundayWeek) // precision → Week, not SundayWeek

// Same via Span
usStart := chron.Now().Span(chron.SundayWeek).Start

// App prefers Sunday weeks globally
chron.DefaultWeekStart = chron.SundayWeek
chron.Now().Truncate(chron.Week)

// Per-user setting — app layer
userWeek := user.Settings.WeekStart // chron.SundayWeek (input selector)
c.Truncate(userWeek)

// Explicit ISO week regardless of DefaultWeekStart
c.Truncate(chron.MondayWeek)
```

`Weeks(n)` in `Duration` remains **seven calendar days** via `AddDate` — it does not
follow week boundary selectors. **`Span` / `Truncate`** with `Week`, `MondayWeek`, or
`SundayWeek` use week boundaries.

### Same-unit comparison (deferred — see [Future features](#future-features-and-options))

| Method | Purpose |
|--------|---------|
| `SameYear(other Chron) bool` | Same calendar year (UTC) |
| `SameMonth(other Chron) bool` | Same calendar month |
| `SameDay(other Chron) bool` | Same calendar day |

```go
anchor := chron.Now()
start := anchor.Truncate(chron.Month)
later := anchor.Add(chron.Days(5))
if anchor.SameMonth(later) { /* ... */ }
```

### Instant comparison (embedded `time.Time`)

`Chron` embeds `time.Time`. **No chron-specific comparison methods** — no
`BeforeInstant`, no precision-aware overrides of `Before` / `After` / `Equal`. Instant
ordering is always nanosecond point comparison, identical to stdlib.

Promoted from the embed (same signatures and semantics as `time.Time`):

| Method | Semantics |
|--------|-----------|
| `Equal(u time.Time) bool` | Same instant |
| `Before(u time.Time) bool` | Strict point ordering |
| `After(u time.Time) bool` | Strict point ordering |
| `Compare(u time.Time) int` | Stdlib three-way compare |

Compare two `Chron` values via the other's embedded instant:

```go
if event.Before(anchor.Time) { /* instant ordering */ }
if deadline.Equal(cutoff.Time) { /* ... */ }
```

`precision` is ignored for these calls — `"2026-02"` (month anchor) compared with
`Before` is still Feb 1 00:00 UTC vs the other instant, not "is inside February."
For month/day/year membership, use `Span` or `SameMonth` / `SameDay` / `SameYear`.

**Naming split:** chron-only APIs use names stdlib does not (`Span`, `SameMonth`,
`Contains`, `Overlaps`, …). `Span` methods are interval logic only and do not shadow
`time.Time`. The deliberate shadow is `Add(chron.Duration)` (calendar-aware), not
comparison.

### Location / display

| Method | Purpose |
|--------|---------|
| `UTC() Chron` | Ensure UTC (idempotent for constructors) |
| `InLocation(loc *time.Location) time.Time` | For display only; returns `time.Time` to signal "leaving chron conventions" |

Storage stays UTC. Formatting for humans converts at the edge.

### Serialization and I/O

**Asymmetric I/O (same pattern as `Duration`):** canonical ISO on marshal, lenient
registry on unmarshal. Human-friendly strings (`"Jan 2026"`, `"02/2026"`) are accepted
on **read** via `ParseFormats`; they are never canonical **write** output. Display for
people uses `Format` / `InLocation` at the UI edge — not JSON.

| Method | Purpose |
|--------|---------|
| `MarshalJSON() ([]byte, error)` | Format at declared `precision` (ISO canonical) |
| `UnmarshalJSON([]byte) error` | Try registered layouts; set `precision` from match |
| `Scan(src any) error` | `database/sql` driver |
| `Value() (driver.Value, error)` | `database/sql` driver |

**Marshal** — one canonical string per `precision`:

| `Precision` | JSON example | Layout |
|-------------|--------------|--------|
| `Year` | `"2026"` | `2006` |
| `Month` | `"2026-02"` | `2006-01` |
| `Day` | `"2026-02-15"` | `2006-01-02` |
| `Hour` | `"2026-02-15T14"` | `2006-01-02T15` |
| `Minute` / `Second` | RFC3339 (no fractional) | `time.RFC3339` |
| `Nanosecond` | RFC3339Nano | `time.RFC3339Nano` |

Full instants from `Now` / `FromTime` marshal as RFC3339Nano unless truncated.

**Unmarshal / parse** — precision is **implied by the layout that matched**, not passed
separately. `Parse`, `ParseFrom`, and `UnmarshalJSON` all use the same rule:

1. Try layouts (registry order for `Parse` / `UnmarshalJSON`; explicit layout for
   `ParseFrom`).
2. On success, set `precision` from the **finest unit in that layout** (e.g.
   `"2006-01"` → `Month`; `"2006-01-02"` → `Day`; RFC3339 with sub-second →
   `Nanosecond`).
3. Normalize the instant to UTC (anchor = start of the matched unit when coarser than
   nanosecond — e.g. `"2026-02"` → Feb 1 00:00:00 UTC).

```go
var ParseFormats = []string{
    time.RFC3339Nano,
    time.RFC3339,
    "2006-01-02T15:04:05",
    "2006-01-02T15",
    "2006-01-02",
    "2006-01",
    "2006",
    // app-specific read-only layouts, e.g. "Jan 2006", "01/02/2006"
}
```

`Parse` / `UnmarshalJSON` pick the **first** matching layout (register narrowest /
most specific layouts before broader ones when ambiguity matters). `ParseFrom(layout, s)` uses
the caller's layout and applies the same precision-from-layout mapping.

```go
c, _ := chron.Parse("2026-02")    // Month precision, anchor Feb 1 UTC
c, _ := chron.Parse("Jan 2026")   // same, if "Jan 2006" is in ParseFormats
c.MarshalJSON()                       // []byte(`"2026-02"`) — canonical, not "Jan 2026"
```

---

## What is `chron.Duration`?

`time.Duration` is always a fixed length in nanoseconds. That works for hours and
minutes, but **a month or a year has no fixed length** until you know the anchor instant
(Jan 31 + 1 month ≠ Feb 31). `chron.Duration` names the kind of offset so `c.Add(d)`
can route calendar units through `AddDate` and clock units through `time.Time.Add`.

### Structure

```go
// Duration is a sum of calendar and clock offsets.
// Calendar fields are applied via AddDate (order: years, months, days).
// Clock is applied last via Add (fixed nanoseconds).
type Duration struct {
    years, months, weeks, days int
    clock                      time.Duration // hours, minutes, seconds, sub-second
}
```

Weeks normalize to days (`weeks * 7`) via `AddDate` at apply time — not
`7 * 24 * time.Hour`. For a fixed 168-hour offset, use `Clock(7 * 24 * time.Hour)`.

```go
// Simple: one unit, one count
Months(3)   // Duration{months: 3}
Weeks(2)    // Duration{weeks: 2}
Hours(4)    // Duration{clock: 4 * time.Hour}

// Composite: chain or struct literal
Years(1).Months(3).Days(15).Hours(12)
NewDuration(years, months, weeks, days int, clock time.Duration) Duration
```

### Unit constructors

| Function | Applies as | Notes |
|----------|------------|-------|
| `Years(n int) Duration` | `AddDate(n, 0, 0)` | Calendar |
| `Months(n int) Duration` | `AddDate(0, n, 0)` | Calendar; length depends on `c` |
| `Weeks(n int) Duration` | `AddDate(0, 0, 7*n)` | Always calendar days, never `7×24h` fixed (see resolved decisions) |
| `Days(n int) Duration` | `AddDate(0, 0, n)` | Calendar days |
| `Hours(n int) Duration` | `Add(n * time.Hour)` | Fixed clock |
| `Min(n int) Duration` | `Add(n * time.Minute)` | Fixed clock |
| `Sec(n int) Duration` | `Add(n * time.Second)` | Fixed clock |
| `Millis(n int) Duration` | `Add(n * time.Millisecond)` | Fixed clock |
| `Micros(n int) Duration` | `Add(n * time.Microsecond)` | Fixed clock |
| `Nanos(n int) Duration` | `Add(n * time.Nanosecond)` | Fixed clock |
| `Clock(d time.Duration) Duration` | `Add(d)` | Raw stdlib duration (escape hatch) |

Sub-second constructors use readable names (`Millis`, `Micros`, `Nanos`) in Go code.
ISO duration **strings** use `ms`, `µs` (marshal), and `ns`; parse also accepts `us` for
microseconds.

### Building `Duration`: constructors or parse

Callers produce a `Duration` in one of two ways; **`c.Add` always takes `Duration`**.
There is one apply path — no parallel `Add` overloads for strings.

| Source | API | Example |
|--------|-----|---------|
| Go code | Unit constructors | `chron.Months(3).Days(5)` |
| Inline string (panic on error) | `Duration(s)` | `chron.Duration("14d")` |
| Boundaries / JSON (errors) | `ParseDuration` | `chron.ParseDuration("P1Y3M4DT12H")` |
| JSON struct field | `Duration.UnmarshalJSON` | string or structured form |

```go
// String constructor — panics on invalid input (tests, literals, c.Add inline)
c.Add(chron.Duration("P2W"))
c.Add(chron.Duration("pt4h500ms"))

// Explicit error handling at I/O boundaries
d, err := chron.ParseDuration(cfg.TrialLength)
c.Add(d)
```

`Duration(s string) Duration` shares the type name (valid Go); it wraps `ParseDuration`
and panics on failure. Prefer `ParseDuration` where errors should propagate.

### ISO 8601 duration parse and format

`ParseDuration`, `FormatDuration`, and `Duration` JSON marshal/unmarshal share one
grammar. Based on ISO 8601 `PnYnMnDTnHnMnS`, with chron extensions for sub-second units
and lowercase canonical output.

#### Sub-second: what ISO 8601 actually defines

ISO 8601 has **no `MS` or `NS` designators**. Sub-second precision uses a **decimal
fraction on `S`**:

| Meaning | Strict ISO | Maps to `clock` |
|---------|------------|-----------------|
| 500 ms | `PT0.5S` or `PT0,5S` | `500 * time.Millisecond` |
| 1 ms | `PT0.001S` | `time.Millisecond` |
| 1 ns | `PT0.000000001S` | `time.Nanosecond` |
| 1.5 s | `PT1.5S` | `1500 * time.Millisecond` |

Both `.` and `,` are accepted as the decimal separator on parse (ISO allows either).
**Marshal uses `.`** (ASCII period).

#### Chron extensions: `ms`, `µs`, `ns`

For readability, chron also accepts explicit sub-second units in the time segment (common
in config, not strict ISO):

| Unit | Marshal (canonical) | Unmarshal (case-insensitive) | Maps to |
|------|---------------------|------------------------------|---------|
| milliseconds | `500ms` | `ms`, `MS`, `Ms`, … | `clock` |
| microseconds | `250µs` | `us`, `US`, `µs`, `μs` | `clock` |
| nanoseconds | `100ns` | `ns`, `NS`, … | `clock` |

**Microseconds:** parse accepts **`us`** (ASCII) or **`µs`** / **`μs`** (Unicode mu
variants). **Marshal and `String()` always emit `µs`** (U+00B5) — never `us`.

These appear only in the **time segment** (after `t` / `T`), alongside `h`, `m`, `s`.
Example marshal: `pt4h30m5s500ms250µs100ns` ↔ parse accepts `250us`, `250µs`, or `250μs`.

Prefer **`ms` / `µs` / `ns` in marshal** when the offset is a whole number of those
units; use **fractional `s`** when that is the natural ISO form (e.g. `pt0.001s` for 1 ms
is also valid on marshal if sub-second is purely fractional).

#### Case rules

| Direction | Rule |
|-----------|------|
| **Marshal** (`FormatDuration`, `String`, `MarshalJSON`) | Lowercase designators: `p`, `y`, `m`, `d`, `w`, `t`, `h`, `s`, `ms`, **`µs`**, `ns` |
| **Unmarshal** (`ParseDuration`, `Duration(s)`, `UnmarshalJSON`) | Case-insensitive; **`us`** and **`µs`** / **`μs`** accepted for microseconds |

Strict ISO uppercase input (`P1Y2M3DT4H30M5S`) parses correctly; output is chron
canonical lowercase (`p1y2m3dt4h30m5s`).

**Ambiguity note:** `m` before `t` = months; `m` after `t` = minutes — same as ISO.
Extension units are multi-letter (`ms`, `µs`, `ns`; parse also `us`) so they do not clash with minutes.

#### Accepted input (lenient)

| Input | Normalized / behavior | Maps to |
|-------|----------------------|---------|
| `P1Y3M4DT12H30M5S` | designators → lowercase internally | composite calendar + clock |
| `p1y3m4dt12h30m5s` | as-is | same |
| `1Y3M4DT12H` | prepend `P` | leading `P`/`p` optional |
| `14D` / `14d` | `P14D` | `Days(14)` |
| `PT12H` / `pt12h` | as-is | `Hours(12)` |
| `12H` | `PT12H` | bare clock unit → prepend `PT` |
| `P2W` / `p2w` | as-is | `Weeks(2)` |
| `PT0.001S` | fractional seconds | 1 ms in `clock` |
| `PT4H500ms` / `pt4h500ms250us` | extension; `us`/`µs`/`μs` on parse | `Hours(4)` + 500 ms (+ µs if present) |
| `-P1D` | as-is | `Neg()` on result |

Normalization before parse:

1. Trim space.
2. If the string does not start with `P`/`p` or `-P`/`-p`, prepend `p`.
3. Bare clock unit at start of period body → insert `t` (`12h` → `pt12h`).

**Errors:** fractional **calendar** components (`P1.5Y`, `P2.5M`); `W` combined with
`Y`/`M`/`D` in one string (ISO rule). Fractional **`s`** and **`ms`/`µs`/`us`/`ns`** are
supported on parse.

| Function | Purpose |
|----------|---------|
| `Duration(s string) Duration` | Parse string; **panic** on error — literals, tests, inline `Add` |
| `ParseDuration(s string) (Duration, error)` | Same grammar; errors returned |
| `FormatDuration(d Duration) string` | Canonical string (`µs` for microseconds) |
| `(Duration) String() string` | Same as `FormatDuration` |
| `(Duration) MarshalJSON()` | `FormatDuration` → JSON string |
| `(Duration) UnmarshalJSON([]byte) error` | JSON string → `ParseDuration` |

Zero duration marshals as `"p0d"` (or `"pt0s"` — pick one at implement time).

```go
type TrialConfig struct {
    Length chron.Duration `json:"length"`
}

// Unmarshal accepts: "14d", "P14D", "pt500ms", "250us", "250µs", "PT0.5S"
// Marshal / String emits: "p14d", "pt500ms", "pt0.5s", "250µs" (never "us")
signupAt.Add(cfg.Length)
signupAt.Add(chron.Duration("P14D"))
```

### Apply order

`c.Add(d)` on the underlying UTC `time.Time`:

1. `AddDate(d.years, d.months, d.weeks*7 + d.days)`
2. `Add(d.clock)`

Same order stdlib callers use when they compose `AddDate` then `Add` manually.

### Comparison to `time.Duration`

| | `time.Duration` | `chron.Duration` |
|--|-----------------|------------------|
| Storage | `int64` nanoseconds | Struct: calendar ints + `time.Duration` |
| "Add 1 month" | Meaningless / wrong if approximated as 30 days | `Months(1)` → correct per anchor |
| "Add 1 week" | `7 * 24 * time.Hour` (fixed) | `Weeks(1)` → `AddDate(0,0,7)` (calendar days) |
| Pass to `time.Sleep` | Yes | Use `Clock()` portion or convert explicitly |
| JSON / config | Awkward for "P1Y3M" style | `ParseDuration` / `UnmarshalJSON` on `Duration` |

Do **not** make `chron.Duration` interchangeable with `time.Duration` — different types,
different meaning. Name the stdlib field `clock` inside the struct to keep the distinction
obvious.

### Operations on `Duration`

| Function | Purpose |
|----------|---------|
| `Neg() Duration` | Flip sign of all components (for `Sub`) |
| `Mul(n int) Duration` | Scale (`Months(2).Mul(3)` → 6 months) |
| `IsZero() bool` | No offset |

Chaining builds composites: `Years(1).Months(3)` merges fields into one struct.

### Examples

```go
// Jan 31 + 1 month = Feb 28/29 (calendar), not ~30 days
chron.Date(2026, 1, 31).Add(chron.Months(1))

// Billing: store the offset, apply when anchor is known
expires := signupAt.Add(chron.Duration("14D")) // panic if invalid

// Full ISO from config (any casing on input; us or µs for microseconds)
d, _ := chron.ParseDuration("P1Y3M4DT12H")
c.Add(d)

// Sub-second: strict ISO fractional s, or extension units
chron.Duration("PT0.001S")       // 1 ms
chron.Duration("pt4h500ms")      // marshal form; PT4H500MS also parses
chron.Duration("pt4h250µs")      // µs on input; us also parses
chron.Duration("PT1.000000001S") // 1 ns via fractional seconds

// Mixed — one call replaces AddDate + Add
c.Add(chron.Years(1).Months(3).Hours(12))
```

---

## What is `chron.Span`?

```go
// Span is a half-open interval [Start, End) in UTC.
// End is exclusive: the first instant NOT in the span.
type Span struct {
    Start Chron
    End   Chron
}
```

Half-open intervals match common range semantics (SQL `BETWEEN` pitfalls aside), avoid
"last nanosecond of the day" ambiguity, and make `Contains` straightforward:
`!t.Before(start.Time) && t.Before(end.Time)`.

### Why a separate type?

Many domains need the **full window**, not just a truncated start:

- Billing period: March 1 00:00 UTC → April 1 00:00 UTC
- Report range: "events this week" with explicit bounds for queries
- Subscription validity: parse `"2026-02"` → span covering all of February
- Overlap checks: does this booking overlap this maintenance window?

A `Chron` with `precision: Month` records *intent* for I/O; a `Span` is the computed
`[start, end)` you can pass to SQL, log, and compare against.

### Construction

Package-level functions when building a span from two endpoints or for validation.
Calendar buckets come from **`c.Span(...)`** (e.g. `c.Span(chron.Month)`).

| Function | Purpose |
|----------|---------|
| `NewSpan(start, end Chron) (Span, error)` | Arbitrary range; error if `!start.Before(end.Time)` |
| `MustSpan(start, end Chron) Span` | Panic on invalid — for tests and constants |

Precision-based parsing:

```go
c, _ := chron.Parse("2026-02")   // Chron with Precision == Month
feb := c.Span(c.Precision())     // Span{2026-02-01, 2026-03-01}
```

### Accessors

| Field / method | Purpose |
|----------------|---------|
| `Start Chron` | Inclusive start |
| `End Chron` | Exclusive end |
| `IsZero() bool` | Both endpoints zero |

### Membership and overlap

| Method | Purpose |
|--------|---------|
| `Contains(c Chron) bool` | Is instant inside `[Start, End)`? |
| `ContainsSpan(other Span) bool` | Is `other` fully inside this span? |
| `Overlaps(other Span) bool` | Do the intervals share any instant? |
| `Adjacent(other Span) bool` | Does `other` start exactly at this `End` (or vice versa)? |

Examples:

```go
feb := chron.Date(2026, 2, 15).Span(chron.Month)
event := chron.FromTime(someEventTime)
if feb.Contains(event) {
    // event in February 2026 (UTC)
}

week := chron.Now().Span(chron.Week)              // DefaultWeekStart (ISO Monday)
usWeek := chron.Now().Span(chron.SundayWeek)
if week.Overlaps(maintenanceWindow) {
    // ...
}

// Query-friendly bounds
db.Query(`... WHERE at >= $1 AND at < $2`, feb.Start.AsTime(), feb.End.AsTime())
```

### Span arithmetic (deferred — see [Future features](#future-features-and-options))

| Method | Purpose |
|--------|---------|
| `Shift(d Duration) Span` | `NewSpan(s.Start.Add(d), s.End.Add(d))` |
| `Extend(end Chron) Span` | Widen end (error if before current start) |
| `Intersection(other Span) (Span, bool)` | Common sub-interval, if any |

Defer `Extend` / `Intersection` until needed; `Shift` is the likely first addition.

### Serialization (deferred — see [Future features](#future-features-and-options))

Spans may JSON as `{ "start": "...", "end": "..." }` or a single string when tied to
a calendar unit (`"2026-02"` → `Span(Month)`). Wire shape TBD in `chron_v1_parsing.md`.

---

## What `chron.Chron` deliberately does *not* do (v1)

- **Not a scheduler or cron engine.** Out of scope; may consume `Chron` later.
- **Not a separate type per precision.** No `chron.Hour` struct; use `Chron` + `Precision`.
- **Not a `time.Time` replacement everywhere.** Use at domain boundaries.
- **Not interface-driven.** No `chron.Time` or `chron.Span` interface.
- **Not locale-aware formatting.** Use `time.Time.Format` after `InLocation` for i18n.
- **Not leap-second modeling.** Follow stdlib behavior.
- **Not a substitute for `time.Duration`.** Use stdlib for sleeps, timeouts, and fixed elapsed time; use `chron.Duration` when the *unit name* matters for calendar math.

---

## Future features and options

Items deferred from v1 or noted elsewhere in this doc. None block the initial
implementation; add when a concrete use case appears.

### Calendar comparison helpers

| Item | Notes |
|------|-------|
| `SameYear` / `SameMonth` / `SameDay` | Convenience over `Span` + `Contains` or truncation equality. Optional v1 — defer if `Span(Month).Contains(c)` is enough. |
| `SameWeek(other Chron, p Precision) bool` | Requires explicit week constant (`MondayWeek` / `SundayWeek` / `Week` + default). Compare via `Span(p)` on both anchors. |

### `Span` API and I/O

| Item | Notes |
|------|-------|
| `Span.Shift(d Duration)` | Move both endpoints by `d`. Straightforward; include when interval arithmetic is needed. |
| `Span.Extend(end Chron)` | Widen end; error if new end precedes start. Defer unless editing ranges in place is common. |
| `Span.Intersection(other)` | Common sub-interval `(Span, bool)`. Defer unless overlap logic needs narrowing, not just detection. |
| **Span JSON shape** | Object `{ "start", "end" }` vs single calendar string (`"2026-02"` → month span). Detail in `chron_v1_parsing.md`. |
| Span `MarshalJSON` / `UnmarshalJSON` | Optional v1; same asymmetric I/O rules as `Chron` when added. |

### Construction and parsing

| Item | Notes |
|------|-------|
| `FromTimeIn(t, loc)` | Preserve local wall time when wrapping `time.Time` instead of normalizing via `.UTC()` first. |
| **Strict vs lenient parse modes** | Global or per-call tightening of `ParseFormats` / `Parse`. `chron_v1_parsing.md`. |
| **Custom parse plugins** | Registry of `func(string) (Chron, error)` — not a `Parser` interface unless multiple backends exist. |
| `Add(chron.Duration("P2W1D"))` | Inline string duration via panic constructor |
| App-specific `ParseFormats` | Human layouts (`"Jan 2006"`, `"02/2026"`) on read only; never canonical marshal. |

### Additional types

| Item | Notes |
|------|-------|
| `Date` (date-only, no clock) | Separate concrete type with conversion to/from `Chron` and `Span` if day-precision `Chron` is insufficient in APIs. |
| Per-precision struct types | Original library had `Year` … `Chron` lattice; v1 uses one `Chron` + `Precision` enum. Revisit only if metadata proves inadequate. |

### Ecosystem and tooling

| Item | Notes |
|------|-------|
| **`chron_v1_precision.md`** | `Precision` enum, truncation table, layout ↔ precision mapping. |
| **`chron_v1_parsing.md`** | Format registry, duration grammar edge cases, Span wire format. |
| **Mock / injectable clock** | Small `Clock` interface in test helper or `clock` package — not in core v1. |
| **Scheduler / cron engine** | Out of scope; may consume `Chron` later. |
| **Business-day / holiday calendars** | Named calendar dates in problem table; needs a richer model than `Chron` + `Span`. |
| **Locale-aware formatting** | Stay in app layer via `InLocation` + `Format`; core remains UTC + ISO I/O. |

### Interfaces (only if forced)

v1 stays interfaceless. Extract an interface when two or more unrelated implementations
exist and callers must accept either — same stance as the original chron rewrite.

---

## Resolved decisions

### Chron JSON and parse I/O

- **Marshal:** precision-aware ISO strings (`"2026-02"` for month, not RFC3339 midnight
  and not human month names). Same canonical shape for `MarshalJSON`, `Value()`, and
  `Format` when using the precision's default layout.
- **Unmarshal / parse:** lenient — try registered layouts; accept human-friendly forms on
  input only.
- **Precision from layout:** whichever layout successfully parses the input sets
  `precision` automatically (`Parse`, `ParseFrom`, `UnmarshalJSON`, `Scan`). No separate
  precision field in JSON for the common `string` field case.
- **Display:** `"Jan 2026"` and locale-specific forms are out of scope for canonical I/O;
  use `Format` after `InLocation` at the application edge.

### Instant comparison (`Chron`)

- **Use embedded `time.Time` methods** — `Before`, `After`, `Equal`, `Compare` with
  stdlib semantics. No alternate names (`BeforeInstant`, …) and no precision-aware
  overrides.
- **Two `Chron` values:** `a.Before(b.Time)` (or compare underlying instants another
  way). Behavior matches `a.Time.Before(b.Time)`.
- **Intervals:** `Span.Contains` / `Overlaps` / `Adjacent` and calendar helpers
  (`SameMonth`, …). Those names do not overlap `time.Time`; do not use `Before` for
  bucket membership.

### Zero values (`Chron`, `Span`)

- **`Chron{}` and `Span{}` are invalid** — never a meaningful instant or interval.
  Constructors and parsers return populated values or an error.
- **`IsZero() bool` on both types** — primary API for callers:
  - `Chron`: `return c.Time.IsZero()` (delegates to embedded `time.Time`)
  - `Span`: `return s.Start.IsZero() && s.End.IsZero()`
- **`reflect`:** no registration or special support required. `reflect.Value.IsZero()`
  walks struct fields; embedded zero `time.Time` is already treated as zero, so
  `reflect.ValueOf(Chron{}).IsZero()` is `true`. **Reflect does not call your
  `IsZero()` method** — keep the method aligned with field-wise zero (same rule as
  above) so direct calls and reflection agree.
- **`encoding/json` `omitempty`:** value-typed struct fields are **never** omitted
  (`isEmptyValue` returns `false` for structs). Options for optional JSON fields:
  - `*chron.Chron` / `*chron.Span` (`nil` = absent), or
  - `MarshalJSON` emits JSON `null` when `IsZero()` (recommended for value fields).
- **`database/sql`:** `Value()` returns `NULL` when `IsZero()`; `Scan` of `NULL` leaves
  the target zero.

### Precision after arithmetic

When `c.Add(d)` or `c.Sub(d)` runs, the returned `Chron` gets a new instant and a new
`precision`. The receiver is never mutated.

- **Finer offset → finer precision** along the ladder through `Hour`, `Minute`, `Second`.
- **Same-granularity offset → same precision.** `Months(1)` on a month-precision value
  stays `Month`.
- **Nanosecond floor.** `FromTime` / `Now` default to `Nanosecond`. Clock-only offsets
  on a `Nanosecond` receiver keep `Nanosecond`; calendar offsets still step precision
  (e.g. `Nanosecond` + `Months(1)` → `Month`).
- **Rationale:** precision metadata tracks *what the caller is expressing now*, not a
  permanent label on the anchor. Serialization and I/O follow the result's precision.

### Weeks are calendar days, not fixed hours

`Weeks(n)` always applies as `AddDate(0, 0, 7*n)` — seven **calendar days** per week,
never `7 * 24 * time.Hour`. A week is a date concept, not a fixed elapsed interval.

For exactly 168 hours of clock time, use `Clock(7 * 24 * time.Hour)` or `Hours(168)`.
Those stay on the `Nanosecond` floor when added to a full instant.

```go
c.Add(chron.Weeks(1))                    // AddDate(0, 0, 7)
c.Add(chron.Clock(7 * 24 * time.Hour))   // fixed 168h — different meaning
```

### Week span boundaries

- **`Week`** — `Span(Week)` / `Truncate(Week)` use package `DefaultWeekStart` (`MondayWeek` by default); result precision is `Week`.
- **`MondayWeek` / `SundayWeek`** — input selectors for explicit boundaries; `Truncate` / `Span` normalize stored precision to `Week`.
- **Per-user locale** — app stores a week-boundary selector (`SundayWeek`, etc.); pass to `Truncate(p)` or `Span(p)` — not on `Chron`.
- **`Span` struct** — only `[Start, End)`; week choice is not retained after construction.
- **`Weeks(n)` duration** — unchanged; always `AddDate(0,0,7n)` calendar days.

### ISO 8601 duration strings (v1)

- **`ParseDuration` / `FormatDuration`** for ISO 8601 calendar + clock durations.
- **Sub-second (ISO):** fractional `s` only (`pt0.001s`, `pt1.5s`); no `MS`/`NS` in strict ISO.
- **Sub-second (chron extensions):** `ms`, `µs` (marshal), `ns` in time segment; parse
  also accepts `us` and Unicode `μs`; unmarshal case-insensitive (`500MS` OK).
- **Marshal / String:** lowercase designators (`p1y2m3dt4h30m5s500ms250µs`); **unmarshal:** any case.
- **Lenient input:** leading `p` optional; bare clock units get `t`.
- **Single apply path:** constructors, `Duration(s)`, or `ParseDuration` → `c.Add(d)`.
- **`Duration(s)`** — panic constructor for inline use; **`ParseDuration`** returns errors.
- **`p…w` weeks** → `Weeks(n)` (calendar days).
- **JSON:** `Duration` uses `MarshalJSON` / `UnmarshalJSON` with same rules.

---

## Example usage (target ergonomics)

```go
// Parse month intent → Chron; derive full window → Span
expiry, err := chron.Parse("2026-02")
feb := expiry.Span(expiry.Precision())

// Domain logic: is now inside the billing month?
if feb.Contains(chron.Now()) {
    // still valid this month
}

// SQL / APIs want explicit bounds
db.Exec(`INSERT INTO periods (start_at, end_at) VALUES ($1, $2)`,
    feb.Start.AsTime(), feb.End.AsTime())

// Calendar math: feb.End is Mar 1 00:00 UTC (exclusive end = next period start)
nextMonth := feb.End
nextFeb := feb.Start.Add(chron.Years(1)).Span(chron.Month)
trialEnd := signupAt.Add(chron.Duration("P2W1D")) // or Weeks(2).Days(1) in code

// Instant comparison (embedded time.Time — not span membership)
if chron.Now().Before(deadline.Time) {
    // ...
}
```

---

## Next documents

- `chron_v1_precision.md` — `Precision` enum, truncation table, serialization mapping
- `chron_v1_parsing.md` — format registry, ISO 8601 duration parse, strict vs lenient modes

See [Future features and options](#future-features-and-options) for the full deferred backlog.

---

## Revision history

| Date | Notes |
|------|-------|
| 2026-06-24 | Initial sketch: problems, goals, `Chron` API surface |
| 2026-06-24 | Interfaceless design: drop `chron.Time` / `dura.Time` interfaces |
| 2026-06-24 | Add concrete `Span` struct: calendar buckets, `Contains`, explicit start/end |
| 2026-06-24 | Replace `AddYears`/`AddMonths`/… with `c.Add(Duration)` and unit constructors |
| 2026-06-24 | `Chron` as receiver: methods for truncation, add, span derivation |
| 2026-06-24 | Precision after `Add`: finest unit in offset; Nanosecond floor for clock-only adds |
| 2026-06-24 | `Weeks(n)` always `AddDate(0,0,7n)`; fixed 168h via `Clock` / `Hours` |
| 2026-06-24 | ISO 8601 `ParseDuration` (optional `P`), constructors + `c.Add(d)` unified |
| 2026-06-24 | Duration I/O: fractional `s` + `ms`/`us`/`ns`; marshal lowercase, parse any case |
| 2026-06-24 | `WeekStart`: ISO Monday default; Sunday option via default, `WithWeekStart`, `WeekSpanFrom` |
| 2026-06-25 | Chron I/O: canonical ISO marshal; lenient parse; precision implied by matched layout |
| 2026-06-25 | Instant compare: embedded `time.Time` `Before`/`After`/`Equal`; `Span` for intervals |
| 2026-06-25 | Zero values: `IsZero()` on `Chron`/`Span`; reflect needs no hook; JSON `null` when zero |
| 2026-06-25 | Parse API: `Parse(s)` registry default; explicit layout via `ParseFrom(layout, s)` |
| 2026-06-25 | `Span(SpanUnit)`: `Precision` or `WeekStart`; drop named spans; `WeekStart` not on `Chron` |
| 2026-06-25 | Drop `WithPrecision` and `StartOf*`; `Truncate(p)` sets precision; keep `Precision()` read |
| 2026-06-25 | Merge `WeekStart` into `Precision`; drop `SpanUnit` interface; `Span(p Precision)` only |
| 2026-06-25 | Duration clock constructors: `Min`, `Sec`, `Millis`, `Micros`, `Nanos` |
| 2026-06-25 | `Duration(s)` panic constructor; parse `us`/`µs`/`μs`; marshal `String()` always `µs` |
