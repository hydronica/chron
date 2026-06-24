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

6. **Interval comparisons stay explicit.** Do not silently change `Before`/`After`
   semantics vs stdlib on `Chron`. Use `Span.Contains` and `Span.Overlaps` for
   membership and overlap.

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
    // weekStart overrides DefaultWeekStart for WeekSpan / StartOfWeek on this
    // Chron. Zero value = use DefaultWeekStart (ISO Monday).
    weekStart WeekStart
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
| `Precision` | enum (`int` or `uint8`) | Metadata on `Chron`: how to truncate and serialize |
| `WeekStart` | enum | Which weekday begins a week span (`Monday` ISO default, or `Sunday`) |
| `Unit` | enum | Named unit inside `Duration` (`Month`, `Week`, `Hour`, …) |

There is no `chron.Time` interface. Callers work with `Chron` and `Span` directly.

**`Chron` vs `Span`.** A `Chron` is a point (or anchor). A `Span` is the full window
around that anchor when you care about start *and* end — "all of February 2026", "the
ISO week containing this timestamp", "business hours today". You can pass a `Span` to
functions, store it in structs, and ask `Contains` without re-deriving boundaries every
time.

**Chron methods vs package functions.** Operations on an instant are methods:
`c.StartOfDay()`, `c.Add(chron.Months(1))`, `c.MonthSpan()`. Package functions are
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
| `Parse(layout, s string) (Chron, error)` | Parse with explicit layout |
| `ParseAuto(s string) (Chron, error)` | Try registered layouts (see Parsing) |

Constructors always store UTC. Passing a `time.Time` in another location converts via
`.UTC()` unless a future `FromTimeIn(loc)` variant is needed.

### Escape hatch

| Method | Purpose |
|--------|---------|
| `AsTime() time.Time` | Return underlying stdlib value for third-party APIs |
| Embedded `time.Time` methods | `Format`, `Unix`, `Year`, `Month`, … work as today |

### Precision metadata

| Method | Purpose |
|--------|---------|
| `Precision() Precision` | Current precision metadata |
| `WithPrecision(p Precision) Chron` | Same instant, different serialization intent |
| `WithWeekStart(w WeekStart) Chron` | Same instant; `WeekSpan()` uses `w` instead of default |
| `WeekStart() WeekStart` | Effective week start for this value (explicit or `DefaultWeekStart`) |

### Truncation and start-of

Return the **inclusive start** of the calendar unit containing the receiver (UTC).
These set `precision` on the result to match the unit.

| Method | Purpose |
|--------|---------|
| `Truncate(p Precision) Chron` | General form — zero sub-units per `p` |
| `StartOfYear() Chron` | Jan 1 00:00:00 UTC of this instant's year |
| `StartOfMonth() Chron` | 1st 00:00:00 UTC of this instant's month |
| `StartOfDay() Chron` | Midnight UTC of this instant's calendar day |
| `StartOfHour() Chron` | Top of the hour containing this instant |
| `StartOfWeek() Chron` | Start of the week containing this instant (uses effective `WeekStart`) |

`c.StartOfMonth()` is equivalent to `c.MonthSpan().Start` — use whichever reads better.

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
| `Hour` | `Minutes(15)` | +15 minutes | `Minute` |
| `Minute` | `Seconds(30)` | +30 seconds | `Second` |
| `Nanosecond` | `Hours(2)` | +2 hours | `Nanosecond` (see below) |
| `Nanosecond` | `Months(1)` | +1 calendar month | `Month` |

**Nanosecond floor for clock offsets.** Values at `Nanosecond` precision — the default
for `FromTime`, `Now`, and any full instant — stay `Nanosecond` when the offset is
**clock-only** (`Hours`, `Minutes`, `Seconds`, `Clock`). Adding two hours to an event
timestamp is still an event timestamp, not an "hour bucket." Calendar fields in `d`
(`Years`, `Months`, `Weeks`, `Days`) still update precision normally.

```go
event := chron.FromTime(loggedAt)     // Nanosecond (default)
event.Precision()                     // Nanosecond
event.Add(chron.Hours(2)).Precision() // Nanosecond — still a full instant

day := chron.Date(2026, 6, 1)         // Day
day.Add(chron.Hours(3)).Precision()   // Hour — caller moved to hour granularity
```

`c.StartOfMonth().Add(chron.Days(5))` is the common case: `StartOfMonth()` yields Jan 1
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

**`StartOf*` / `Truncate` / `WithPrecision`** set precision directly from the operation;
they do not use the rule above.

```go
jan := chron.Date(2026, 1, 15).StartOfMonth() // Jan 1, Month
jan.Precision()                               // Month
sixth := jan.Add(chron.Days(5))               // Jan 6, Day
sixth.Precision()                             // Day — original jan unchanged
```

### Span derivation

Build the calendar bucket **containing** this instant. Returns `[start, end)` as `Span`.

| Method | Purpose |
|--------|---------|
| `Span(p Precision) Span` | General form at precision `p`; `Week` uses effective `WeekStart` |
| `YearSpan() Span` | `[Jan 1, Jan 1 next year)` containing this instant |
| `MonthSpan() Span` | `[1st 00:00, 1st next month)` |
| `WeekSpan() Span` | Week containing this instant (effective `WeekStart`) |
| `WeekSpanFrom(w WeekStart) Span` | Same, but `w` for this call only — ignores `c.weekStart` |
| `DaySpan() Span` | `[midnight, midnight next day)` |
| `HourSpan() Span` | `[hour boundary, next hour)` |

#### Week boundaries (`WeekStart`)

Two week definitions are supported. **Default is ISO 8601 (Monday start).**

```go
type WeekStart uint8

const (
    MondayWeek WeekStart = iota // ISO 8601 — default
    SundayWeek                  // US-style weeks starting Sunday 00:00 UTC
)

var DefaultWeekStart = MondayWeek
```

| Mode | Week span `[Start, End)` | Notes |
|------|--------------------------|-------|
| `MondayWeek` | ISO week: Mon 00:00 UTC → next Mon 00:00 | Week-number rules follow ISO 8601 |
| `SundayWeek` | Sun 00:00 UTC → next Sun 00:00 | Calendar weeks aligned to Sunday |

**Where to set it**

| Mechanism | Scope | Use when |
|-----------|-------|----------|
| `DefaultWeekStart` | Package / process | App-wide locale default |
| `c.WithWeekStart(w)` | This `Chron` value | User preference stored on a record |
| `c.WeekSpanFrom(w)` | Single call | One-off report without mutating `c` |
| `FromTime(t, opts...)` / `ParseAuto` (future) | At construction | Ingest knows locale up front |

Zero `weekStart` on `Chron` means "use `DefaultWeekStart`" — constructors do not need
to set it unless overriding.

```go
// Default ISO week (Monday)
chron.Now().WeekSpan()

// App prefers Sunday weeks
chron.DefaultWeekStart = chron.SundayWeek

// Per-user setting on one anchor
userAnchor := chron.FromTime(t).WithWeekStart(chron.SundayWeek)
userAnchor.WeekSpan()

// Explicit for one query, anchor unchanged
c.WeekSpanFrom(chron.SundayWeek)
```

`Weeks(n)` in `Duration` remains **seven calendar days** via `AddDate` — it does not
follow `WeekStart`. Only **span boundaries** (`WeekSpan`, `StartOfWeek`) use `WeekStart`.

### Same-unit comparison (optional v1)

| Method | Purpose |
|--------|---------|
| `SameYear(other Chron) bool` | Same calendar year (UTC) |
| `SameMonth(other Chron) bool` | Same calendar month |
| `SameDay(other Chron) bool` | Same calendar day |

```go
anchor := chron.Now()
start := anchor.StartOfMonth()
later := anchor.Add(chron.Days(5))
if anchor.SameMonth(later) { /* ... */ }
```

### Instant comparison (stdlib semantics)

| Method | Semantics |
|--------|-----------|
| `Equal(other Chron) bool` | Same instant (nanosecond) |
| `BeforeInstant(other Chron) bool` | Strict point ordering |
| `AfterInstant(other Chron) bool` | Strict point ordering |

Names are intentionally verbose in v1 sketch to avoid overriding embedded
`Before`/`After` with different meaning. **Open question:** embed `time.Time` and use
`BeforeInstant` only, or override `Before`/`After` when `precision != Nanosecond`?
Recommendation: **keep stdlib `Before`/`After` as instant comparison always**; use
`Span` for intervals.

### Location / display

| Method | Purpose |
|--------|---------|
| `UTC() Chron` | Ensure UTC (idempotent for constructors) |
| `InLocation(loc *time.Location) time.Time` | For display only; returns `time.Time` to signal "leaving chron conventions" |

Storage stays UTC. Formatting for humans converts at the edge.

### Serialization and I/O

| Method | Purpose |
|--------|---------|
| `MarshalJSON() ([]byte, error)` | Format according to `precision` |
| `UnmarshalJSON([]byte) error` | Try registered formats; infer precision when possible |
| `Scan(src any) error` | `database/sql` driver |
| `Value() (driver.Value, error)` | `database/sql` driver |

**Parse format registry** (global or package-level):

```go
var ParseFormats = []string{
    time.RFC3339Nano,
    "2006-01-02",
    "2006-01",
    // append for app-specific formats
}
```

Unmarshal picks the first matching layout or returns a clear error. Precision is set
from the narrowest layout that matched (e.g. `"2006-01"` → `Month` precision).

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
| `Minutes(n int) Duration` | `Add(n * time.Minute)` | Fixed clock |
| `Seconds(n int) Duration` | `Add(n * time.Second)` | Fixed clock |
| `Milliseconds(n int) Duration` | `Add(n * time.Millisecond)` | Fixed clock |
| `Microseconds(n int) Duration` | `Add(n * time.Microsecond)` | Fixed clock |
| `Nanoseconds(n int) Duration` | `Add(n * time.Nanosecond)` | Fixed clock |
| `Clock(d time.Duration) Duration` | `Add(d)` | Raw stdlib duration (escape hatch) |

### Building `Duration`: constructors or parse

Callers produce a `Duration` in one of two ways; **`c.Add` always takes `Duration`**.
There is one apply path — no parallel `Add` overloads for strings.

| Source | API | Example |
|--------|-----|---------|
| Go code | Unit constructors | `chron.Months(3).Days(5)` |
| Config, query params, JSON | `ParseDuration` | `chron.ParseDuration("P1Y3M4DT12H")` |
| JSON struct field | `Duration.UnmarshalJSON` | string or structured form |

```go
// Both paths → same type → same Add
c.Add(chron.Weeks(2))
d, err := chron.ParseDuration("P2W")
c.Add(d)
```

Optional convenience when the string is inline (parse errors propagate):

```go
func (c Chron) AddParsed(s string) (Chron, error)
func MustDuration(s string) Duration   // panic on error — tests, init
```

Prefer `ParseDuration` at boundaries; use `AddParsed` only when it reads cleaner.

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

#### Chron extensions: `ms`, `us`, `ns`

For readability, chron also accepts and emits explicit sub-second units (common in
config, not strict ISO):

| Unit | Marshal (canonical) | Unmarshal (case-insensitive) | Maps to |
|------|---------------------|------------------------------|---------|
| milliseconds | `500ms` | `ms`, `MS`, `Ms`, … | `clock` |
| microseconds | `250us` | `us`, `US`, `µs`, `μs` | `clock` |
| nanoseconds | `100ns` | `ns`, `NS`, … | `clock` |

These appear only in the **time segment** (after `t` / `T`), alongside `h`, `m`, `s`.
Example: `pt4h30m5s500ms250us100ns` (marshal) ↔ same string with any casing on parse.

Prefer **`ms` / `us` / `ns` in marshal** when the offset is a whole number of those
units; use **fractional `s`** when that is the natural ISO form (e.g. `pt0.001s` for 1 ms
is also valid on marshal if sub-second is purely fractional).

#### Case rules

| Direction | Rule |
|-----------|------|
| **Marshal** (`FormatDuration`, `MarshalJSON`) | **Lowercase** designators: `p`, `y`, `m`, `d`, `w`, `t`, `h`, `s`, `ms`, `us`, `ns` |
| **Unmarshal** (`ParseDuration`, `UnmarshalJSON`) | **Case-insensitive** for all designators (`P`/`p`, `Y`/`y`, `PT`/`pt`, `MS`/`ms`, …) |

Strict ISO uppercase input (`P1Y2M3DT4H30M5S`) parses correctly; output is chron
canonical lowercase (`p1y2m3dt4h30m5s`).

**Ambiguity note:** `m` before `t` = months; `m` after `t` = minutes — same as ISO.
Extension units are multi-letter (`ms`, `us`, `ns`) so they do not clash with minutes.

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
| `PT4H500ms` / `pt4h500ms` | extension | `Hours(4)` + 500 ms |
| `-P1D` | as-is | `Neg()` on result |

Normalization before parse:

1. Trim space.
2. If the string does not start with `P`/`p` or `-P`/`-p`, prepend `p`.
3. Bare clock unit at start of period body → insert `t` (`12h` → `pt12h`).

**Errors:** fractional **calendar** components (`P1.5Y`, `P2.5M`); `W` combined with
`Y`/`M`/`D` in one string (ISO rule). Fractional **`s`** and **`ms`/`us`/`ns`** are
supported.

| Function | Purpose |
|----------|---------|
| `ParseDuration(s string) (Duration, error)` | ISO + extensions → `Duration` |
| `FormatDuration(d Duration) string` | `Duration` → lowercase canonical string |
| `MustDuration(s string) Duration` | Parse or panic |
| `(Duration) MarshalJSON()` | `FormatDuration` → JSON string |
| `(Duration) UnmarshalJSON([]byte) error` | JSON string → `ParseDuration` |

Zero duration marshals as `"p0d"` (or `"pt0s"` — pick one at implement time).

```go
type TrialConfig struct {
    Length chron.Duration `json:"length"`
}

// Unmarshal accepts: "14d", "P14D", "pt500ms", "PT0.5S"
// Marshal emits:       "p14d", "pt500ms", "pt0.5s"
signupAt.Add(cfg.Length)
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
expires := signupAt.Add(chron.MustDuration("14D")) // P prefix optional

// Full ISO from config (any casing on input)
d, _ := chron.ParseDuration("P1Y3M4DT12H")
c.Add(d)

// Sub-second: strict ISO fractional s, or extension units
chron.MustDuration("PT0.001S")       // 1 ms
chron.MustDuration("pt4h500ms")      // marshal form; PT4H500MS also parses
chron.MustDuration("PT1.000000001S") // 1 ns via fractional seconds

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
`!t.Before(start) && t.Before(end)`.

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
Calendar buckets come from **`Chron` methods** (`c.MonthSpan()`).

| Function | Purpose |
|----------|---------|
| `NewSpan(start, end Chron) (Span, error)` | Arbitrary range; error if `!start.Before(end)` |
| `MustSpan(start, end Chron) Span` | Panic on invalid — for tests and constants |

Precision-based parsing:

```go
c, _ := chron.ParseAuto("2026-02")   // Chron with Precision == Month
feb := c.MonthSpan()                 // Span{2026-02-01, 2026-03-01}
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
feb := chron.Date(2026, 2, 15).MonthSpan()
event := chron.FromTime(someEventTime)
if feb.Contains(event) {
    // event in February 2026 (UTC)
}

week := chron.Now().WeekSpan()              // ISO Monday default
usWeek := chron.Now().WeekSpanFrom(chron.SundayWeek)
if week.Overlaps(maintenanceWindow) {
    // ...
}

// Query-friendly bounds
db.Query(`... WHERE at >= $1 AND at < $2`, feb.Start.AsTime(), feb.End.AsTime())
```

### Span arithmetic (optional v1)

| Method | Purpose |
|--------|---------|
| `Shift(d Duration) Span` | `NewSpan(s.Start.Add(d), s.End.Add(d))` |
| `Extend(end Chron) Span` | Widen end (error if before current start) |
| `Intersection(other Span) (Span, bool)` | Common sub-interval, if any |

Defer `Extend` / `Intersection` if needed; `Shift` uses `Add` on both endpoints.

### Serialization (optional v1)

Spans may JSON as `{ "start": "...", "end": "..." }` or a single string when tied to
a calendar unit (`"2026-02"` → `MonthSpan()`). Exact shape TBD in `chron_v1_parsing.md`.

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

## Open questions

1. **JSON shape for `Chron`.** Should month-precision serialize as `"2026-02"` or always
   RFC3339? Proposal: serialize at declared precision; always accept multiple on input.

2. **Comparison API naming on `Chron`.** `BeforeInstant` vs overriding embedded methods —
   pick one story and document it prominently.

3. **Zero values.** `Chron{}` and `Span{}` invalid; provide `IsZero() bool` on both.

4. **`SameUnit` helpers.** Covered by optional `SameMonth` / `SameDay` — include in v1 or defer?

5. **`SameWeek` helper.** Should it respect `WeekStart` / `WeekSpanFrom` semantics?

---

## Resolved decisions

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

### Week span boundaries (`WeekStart`)

- **`MondayWeek`** (default): ISO 8601 weeks — Monday 00:00 UTC through next Monday
  (exclusive). `DefaultWeekStart = MondayWeek`.
- **`SundayWeek`**: Sunday 00:00 UTC through next Sunday (exclusive).
- **Configure at:** package default (`DefaultWeekStart`), per-`Chron` (`WithWeekStart`),
  or per call (`WeekSpanFrom(w)`).
- **`Weeks(n)` duration** is unchanged — always `AddDate(0,0,7n)` calendar days; only
  span/start-of-week boundaries use `WeekStart`.

### ISO 8601 duration strings (v1)

- **`ParseDuration` / `FormatDuration`** for ISO 8601 calendar + clock durations.
- **Sub-second (ISO):** fractional `s` only (`pt0.001s`, `pt1.5s`); no `MS`/`NS` in strict ISO.
- **Sub-second (chron extensions):** `ms`, `us`, `ns` in time segment; marshal lowercase;
  unmarshal case-insensitive (`500MS` OK).
- **Marshal:** lowercase designators (`p1y2m3dt4h30m5s500ms`); **unmarshal:** any case.
- **Lenient input:** leading `p` optional; bare clock units get `t`.
- **Single apply path:** constructors and parsed strings → `Duration` → `c.Add(d)`.
- **`p…w` weeks** → `Weeks(n)` (calendar days).
- **JSON:** `Duration` uses `MarshalJSON` / `UnmarshalJSON` with same rules.

---

## Example usage (target ergonomics)

```go
// Parse month intent → Chron; derive full window → Span
expiry, err := chron.ParseAuto("2026-02")
feb := expiry.MonthSpan()

// Domain logic: is now inside the billing month?
if feb.Contains(chron.Now()) {
    // still valid this month
}

// SQL / APIs want explicit bounds
db.Exec(`INSERT INTO periods (start_at, end_at) VALUES ($1, $2)`,
    feb.Start.AsTime(), feb.End.AsTime())

// Calendar math: feb.End is Mar 1 00:00 UTC (exclusive end = next period start)
nextMonth := feb.End
nextFeb := feb.Start.Add(chron.Years(1)).MonthSpan()
trialEnd := signupAt.Add(chron.MustDuration("P2W1D")) // or Weeks(2).Days(1) in code

// Explicit instant comparison (not span membership)
if chron.Now().BeforeInstant(deadline) {
    // ...
}
```

---

## Next documents

- `chron_v1_precision.md` — `Precision` enum, truncation table, serialization mapping
- `chron_v1_parsing.md` — format registry, ISO 8601 duration parse, strict vs lenient modes

## When we might add interfaces (later, if ever)

| Need | Likely approach |
|------|-----------------|
| Mock clock in tests | Small `Clock` interface in a `_test` helper or separate `clock` package |
| Custom parse plugins | Registry of `func(string) (Chron, error)` functions, not a `Parser` interface |

Default stance: add concrete types and functions first; extract an interface only when
two or more unrelated implementations exist and callers need to accept either.

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
