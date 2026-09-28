# chron v2 — Span and Duration design notes

Follow-on to [`chron_v1.md`](chron_v1.md). This document is the **source of truth for
shipped v2 behavior**. Where it conflicts with v1 (half-open spans, reverse rejection,
duration zero/`P1.5Y`/mixed `W`), **v2 wins**.

---

## Span

### Model: closed inclusive

`Span` is a **closed interval `[Start, End]`** in UTC. Both endpoints are inclusive.

| Case | Valid? | Behavior |
|------|--------|----------|
| `Start.Before(End)` | Yes | Forward span |
| `Start.Equal(End)` | Yes | Point span — `Contains` that instant only |
| `Start.After(End)` | Yes | Reverse — stored as given; membership uses sorted bounds; `Each` walks Start → End |

`NewSpan(start, end Chron) Span` always succeeds. `MustSpan` was removed.

`SpanBetween(a, b)` returns chronologically ordered endpoints (`min`/`max`) and does not
preserve direction.

### Calendar buckets: Truncate + EndOf

```go
func (c Chron) Truncate(p Precision) Chron // inclusive start of unit
func (c Chron) EndOf(p Precision) Chron    // last nanosecond of unit
func (c Chron) Span(p Precision) Span      // NewSpan(Truncate(p), EndOf(p))
```

Full February example:

- `Truncate(Month)` → 2026-02-01 00:00:00 UTC  
- `EndOf(Month)` → 2026-02-28 23:59:59.999999999 UTC  
- `Span(Month)` covers the whole month; March 1 is outside  

### SQL

Use sorted bounds:

```sql
WHERE at >= $1 AND at <= $2  -- lo, hi := chronological Start/End
```

**Warning:** adjacent inclusive periods that share an endpoint **overlap**. Prefer
`Chron.Span(Month)` chaining (last nano of Jan + 1ns = Feb 1) for gapless periods;
`Adjacent` means contiguous without overlap (`hi + 1ns == lo`).

### Method tiers

| Method | Tier | Notes |
|--------|------|-------|
| `Contains` | Core | Closed membership on sorted bounds |
| `Overlaps` | Core | Shared endpoint ⇒ overlap |
| `NewSpan` | Core | No error; point and reverse allowed |
| `IsZero` | Core | Both endpoints zero |
| `Each` | Core | Directed inclusive walk (`iter.Seq`, Go 1.23+) |
| `SpanBetween` | Core | Order-normalizing constructor |
| `ContainsSpan` | Advanced | Sub-period fully inside parent |
| `Adjacent` | Advanced | Contiguous without overlap (1ns gap) |

Deferred: `Shift`, `Extend`, `Intersection`, Span JSON.

### Real-world applications

| Domain | Pattern | Why Span |
|--------|---------|----------|
| SQL / API bounds | `WHERE at >= $1 AND at <= $2` | Closed `[lo, hi]` |
| Billing / subscriptions | `"2026-02"` → month window | `Span(Month)` via Truncate+EndOf |
| Overlap / conflict | booking vs maintenance | `Overlaps` |
| Reporting / batching | every day in range | `Each(Days(1))` |
| Point / reverse batch | flowlord `At` / reverse | Point and reverse spans |

### Span.Each

```go
func (s Span) Each(step Duration) iter.Seq[Chron]
```

| Case | Behavior |
|------|----------|
| Zero step | Empty sequence |
| Point (`Start == End`) | Yield `Start` once |
| Forward | Yield while `Contains`, including `End` when landed on |
| Reverse | Advance with `Neg(step)`; caller passes a positive step |
| Non-progressing step | Stop (no infinite loop) |

```go
jan := NewSpan(Date(2026, 1, 1), Date(2026, 1, 31))
for d := range jan.Each(Days(1)) {
    // 2026-01-01 … 2026-01-31 inclusive
}
```

---

## Duration

### Relationship to `time.Duration` and ISO 8601

| | `time.Duration` | `chron.Duration` |
|--|-----------------|------------------|
| Model | Fixed nanoseconds | Collapsed calendar months/days + clock |
| "Add 1 month" | Wrong if approximated as 30 days | Correct via `AddDate` |
| Primary use | Sleep, timeouts | Billing, trials, config |
| Standard | Go convention | ISO 8601 + chron extensions |

### Zero duration

| API | Zero behavior |
|-----|---------------|
| `String()` | `""` |
| `MarshalJSON` | `null` |
| `omitempty` | Requires `*Duration` |

### Lenient parse (chron extensions)

- Accept `P1Y1W`, `P2W1D` (ISO forbids mixing `W` with `Y`/`M`/`D`).
- Parse `P1.5Y` as 18 months; reject fractional months/days.
- Marshal uses canonical decomposition: `months` → `y`+`m`, `days` → `w`+`d`
  (e.g. `Days(14)` and `Weeks(2)` both stringify as `p2w`; `P1Y1W` → `p1y1w`).

### Internal model (shipped)

```go
type Duration struct {
    months int32         // years*12 + months
    days   int32         // weeks*7 + days
    clock  time.Duration
}
```

Apply: `AddDate(0, months, days)` then `Add(clock)`.

Canonical `String()` only — no preferred-unit flags in v2.

---

## Supersedes v1

| Topic | v1 | v2 (shipped) |
|-------|----|--------------|
| Span interval | Half-open `[Start, End)` | Closed `[Start, End]` |
| Reverse / point | Invalid | Valid |
| `NewSpan` | `(Span, error)` + `MustSpan` | `Span` only |
| Month bucket end | Next month start (exclusive) | `EndOf(Month)` last nano |
| Zero `Duration` | `"p0d"` TBD | `""` / JSON `null` |
| `P1.5Y` | Reject | 18 months |
| Mixed `W`+`YMD` | Reject | Accept |

---

## Deferred (v3 candidates)

- `Span.Shift`, `Extend`, `Intersection`
- Span JSON shape
- `SameYear` / `SameMonth` / `SameDay` helpers
- `Date` as separate type
- Business-day / holiday calendars
- Preferred-unit Duration flags

---

## Revision history

| Date | Notes |
|------|-------|
| 2026-07-02 | Initial v2 notes: Span tiers, iteration, reverse-span rejection, Duration zero/ISO/normalization |
| 2026-09-28 | Shipped: inclusive Span, EndOf, Each, reverse/point; Duration collapse and I/O |
