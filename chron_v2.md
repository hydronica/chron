# chron v2 — Span and Duration design notes

Follow-on to [`chron_v1.md`](chron_v1.md). v1 covers core types and API surface; this
document records **v2 decisions**, **resolved open questions**, and **deferred items**
from implementation review.

---

## Span

### Real-world applications

Span is most valuable when a domain needs **both bounds** of a window, not just a
truncated start. Primary patterns:

| Domain | Pattern | Why Span |
|--------|---------|----------|
| SQL / API bounds | `WHERE at >= $1 AND at < $2` | Half-open `[Start, End)` maps directly to query params |
| Billing / subscriptions | `"2026-02"` → month window | `Parse` gives Chron intent; `Span(Month)` gives the full period |
| Overlap / conflict | booking vs maintenance window | `Overlaps` without manual endpoint math |
| Reporting | "events this week" | `Now().Span(Week)` → explicit start/end for logs and exports |
| Validity checks | "is this event in February?" | `Contains(event)` — membership, not precision-aware `Before` |
| Period stitching | consecutive billing cycles | `Adjacent` detects gapless period chains |

The primary ergonomic win is **`Chron.Span(p)` → Span**:

```go
feb := expiry.Span(expiry.Precision())
db.Query(`... WHERE at >= $1 AND at < $2`, feb.Start.AsTime(), feb.End.AsTime())
```

### Method tiers

| Method | Tier | Notes |
|--------|------|-------|
| `Contains` | Core | Primary membership check |
| `Overlaps` | Core | Scheduling, conflicts, maintenance windows |
| `NewSpan` / `MustSpan` | Core | Arbitrary ranges from two endpoints |
| `IsZero` | Core | Pairs with Chron zero semantics |
| `ContainsSpan` | Advanced | "Is sub-period fully inside parent?" (trial inside billing period) |
| `Adjacent` | Advanced | Period-chaining; defer emphasis until a concrete use case appears |

Deferred from v1 (add when needed): `Shift`, `Extend`, `Intersection`, Span JSON.

### Reverse spans — rejected

Do **not** support reverse spans (`Start` after `End`) as valid `Span` values.

- Half-open `[Start, End)` assumes `Start < End`; `Contains`, `Overlaps`, and SQL mapping
  all break or need special cases.
- A reversed interval is a **direction** or **signed offset**, not a span — that belongs
  on `Duration` / future `Chron` distance helpers.
- Current behavior (`NewSpan` errors when `!start.Before(end)`) is correct.

Callers that need "between A and B regardless of order" normalize at construction:

```go
func SpanBetween(a, b Chron) Span {
    if a.Before(b.Time) {
        return MustSpan(a, b)
    }
    return MustSpan(b, a)
}
```

### Span iteration — add in v2

Traversing a span by a fixed step is a high-value addition (e.g. every 3 days within a
month range).

Proposed API:

```go
// Each yields anchor points every step while still inside the span.
func (s Span) Each(step Duration) iter.Seq[Chron]  // Go 1.23+
```

Example:

```go
jan := MustSpan(Date(2026, 1, 1), Date(2026, 2, 1))
for d := range jan.Each(Days(3)) {
    // 2026-01-01, 2026-01-04, 2026-01-07, ...
}
```

Semantics:

| Rule | Behavior |
|------|----------|
| Interval | Half-open — yield while `Contains(d)`; never emit `End` |
| Step | Uses `Chron.Add(step)` — calendar steps are anchor-dependent |
| Precision | Yield with step's precision (e.g. `Days(3)` → `Day`) |
| Partial final step | Stop before `End`; do not extend the span |

Use cases: billing checkpoints, report bucket generation, trial milestone dates,
"every N days in range" without building a full scheduler.

---

## Duration

### Relationship to `time.Duration` and ISO 8601

These solve **different problems**:

| | `time.Duration` | `chron.Duration` |
|--|-----------------|------------------|
| Model | Fixed nanoseconds | Calendar components + clock |
| "Add 1 month" | Wrong if approximated as 30 days | Correct via `AddDate` |
| Primary use | Sleep, timeouts, profiling | Billing, trials, subscriptions, config |
| Standard | Go convention | ISO 8601 `PnYnMnDTnHnMnS` (+ chron extensions) |

Go deliberately keeps calendar math separate (`AddDate`) from elapsed time (`Add`).
chron unifies them behind one type — that is a real problem, not a redundant abstraction.

**ISO stance:** follow ISO for interchange (parse/marshal), with documented chron
extensions (bare `14d`, `ms`/`µs`/`ns`, optional leading `P`). Be explicit where chron
**intentionally diverges** from strict ISO (see composite `W` below).

### Zero duration

Align with `Chron` zero semantics:

| API | Zero behavior |
|-----|---------------|
| `String()` | `""` (not `"p0d"`) |
| `MarshalJSON` | `null` |
| Struct `omitempty` | Requires `*Duration` — Go never omits value-typed struct fields via `omitempty` |

### Lenient parse: composite calendar + week (`P1Y1W`, `P2W1D`)

ISO 8601 forbids mixing `W` with `Y`/`M`/`D`. Go constructors already allow composites
(`Years(1).Weeks(1)`) because all fields feed one `AddDate` call.

**Decision:** lenient parse, documented chron extension.

- **Parse:** accept `P1Y1W`, `P2W1D`, etc.
- **Marshal:** pick a canonical form (prefer ISO when possible, e.g. `p1y7d`; or emit
  chron extension `p1y1w` — decide at implement time and test round-trip).

### Fractional calendar years (`P1.5Y`)

**Decision:** parse `P1.5Y` as `{months: 18}` at parse time. Reject fractional
months and days. Internal representation stays integer-only.

---

## Normalization: weeks → days, years → months

This section resolves the open question: *why treat weeks and years differently when
collapsing internal fields?*

### Short answer

**For apply semantics, you can collapse both.** They are equally safe in a single
`AddDate` call. The reasons to keep separate fields — or to collapse internally with
careful `String()` / `Precision()` rules — are **representational**, not arithmetic.

### Why weeks → days collapse is obviously safe

`chron.Duration` applies calendar fields in one call:

```go
t := c.Time.AddDate(d.years, d.months, d.weeks*7+d.days).Add(d.clock)
```

`weeks` and `days` both contribute to the **third** `AddDate` argument only.
`Duration{weeks: 2, days: 3}` and `Duration{days: 17}` are identical for every anchor.

### Years → months is also safe for apply semantics

`years` and `months` contribute to the **first two** `AddDate` arguments. The question
is whether `AddDate(y, m, d)` always equals `AddDate(0, y*12+m, d)`.

Exhaustive testing over anchor dates (2020–2028, varying month/day) and offsets
(Y: 0–3, M: 0–24, D: 0–31) found **zero differences** in Go's `time.Time.AddDate`.

So `Duration{years: 1, months: 6}` and `Duration{months: 18}` produce the same instant
for every tested anchor when applied via a single `AddDate`.

**Correction from v1 review:** an earlier caution about Feb 29 / end-of-month edge cases
does not hold for Go's single-call `AddDate`. Collapsing years into months is
**arithmetically equivalent** to collapsing weeks into days.

### So why not always collapse years into months?

Not because apply would be wrong — but because **storage normalization changes behavior
outside of `Add`**:

| Concern | Separate fields | Collapsed to total months / total days |
|---------|-----------------|----------------------------------------|
| **`String()` / JSON** | `Years(1)` → `"p1y"`; `Months(12)` → `"p12m"` | Both become `"p12m"` or need decomposition rules |
| **`Precision()`** | `Years(1)` → `Year`; `Months(12)` → `Month` | Total 12 months → `Month` unless decomposed |
| **Parse round-trip** | `P1Y` should marshal back as `p1y`, not `p12m` | Requires canonical decomposition on output |
| **Author intent** | `P1Y3M` vs `P15M` — same apply, different meaning to humans/config | Normalization erases intent unless tracked |

Weeks → days has the **same representational tradeoffs**:

| Constructor | Apply | `String()` | `Precision()` |
|-------------|-------|------------|---------------|
| `Weeks(2)` | `AddDate(0,0,14)` | `"p2w"` | `Week` |
| `Days(14)` | `AddDate(0,0,14)` | `"p14d"` | `Day` |

Collapsing weeks into days internally would turn `Weeks(2)` into `Days(14)` unless
`String()` decomposes `days` back into `w` when `days % 7 == 0` (and no Y/M present).

**The same logic applies to years/months:** collapse is safe internally if output
decomposes back into the preferred units.

### Recommended internal model (v2)

Use **two collapsed calendar fields + clock**:

```go
type Duration struct {
    months int32         // years*12 + months (from input)
    days   int32         // weeks*7 + days (from input)
    clock  time.Duration
}
```

- **Apply:** `AddDate(0, months, days)` then `Add(clock)` — equivalent to current behavior.
- **Constructors** (`Years`, `Weeks`, …) accumulate into these totals.
- **`String()`** decomposes for output:
  - `months` → `y` + `m` (`months/12`, `months%12`)
  - `days` → `w` + `d` when ISO week form is preferred (`days/7`, `days%7` with rules)
- **`Precision()`** inspects the **decomposed** largest non-zero unit, not raw totals
  (so 14 days can still report `Week` when constructed via `Weeks(2)` — see below).

### Tracking author intent (optional)

If round-trip fidelity matters (`P1Y` must not become `p12m`), two approaches:

1. **Canonical `String()` only** — always decompose totals (simplest; may change wire
   shape on marshal).
2. **Preferred-unit flags** — e.g. a small bitfield recording whether input used `Y` or
   `M`, `W` or `D` (more complex; best round-trip).

Start with (1) unless config round-trip tests require (2).

### Struct size

Current layout (64-bit): **40 bytes** (4× `int` + `int64` clock).

Collapsed layout with `int32`:

```go
type Duration struct {
    months int32
    days   int32
    clock  time.Duration
}
// ~16 bytes
```

Down from 40 bytes. The size difference vs `time.Duration` (8 bytes) is the cost of
calendar semantics; collapsed `int32` form is sufficient unless profiles show otherwise.

### Summary table

| Normalization | Apply equivalent? | Safe to collapse internally? | Why keep separate on wire |
|---------------|-------------------|-------------------------------|---------------------------|
| weeks → days | Yes | Yes | `p2w` vs `p14d`; `Precision()` = `Week` vs `Day` |
| years → months | Yes | Yes | `p1y` vs `p12m`; `Precision()` = `Year` vs `Month` |

**Bottom line:** weeks and years should be treated the same — both can collapse for
storage and apply. The design choice is whether `String()`, `Precision()`, and parse
round-trip **decompose** totals back into the units the author expressed.

---

## Deferred (unchanged from v1)

- `Span.Shift`, `Extend`, `Intersection`
- Span JSON shape
- `SameYear` / `SameMonth` / `SameDay` helpers
- `Date` as separate type
- Business-day / holiday calendars

---

## Revision history

| Date | Notes |
|------|-------|
| 2026-07-02 | Initial v2 notes: Span tiers, iteration, reverse-span rejection, Duration zero/ISO/normalization |
