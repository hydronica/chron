package chron

import "iter"

// Span is a closed interval [Start, End] in UTC.
// Both endpoints are inclusive. Start may equal End (point span) or follow End
// (reverse span). Membership and overlap use chronological bounds; Each walks
// from Start toward End.
type Span struct {
	Start Chron
	End   Chron
}

// NewSpan builds a span from start to end. Point and reverse spans are valid.
func NewSpan(start, end Chron) Span {
	return Span{Start: start, End: end}
}

// SpanBetween returns a chronologically ordered span (lo, hi) regardless of
// argument order. Direction is not preserved — use NewSpan for directed spans.
func SpanBetween(a, b Chron) Span {
	if a.After(b.Time) {
		return NewSpan(b, a)
	}
	return NewSpan(a, b)
}

// IsZero reports whether both endpoints are zero.
func (s Span) IsZero() bool {
	return s.Start.IsZero() && s.End.IsZero()
}

// bounds returns chronological min/max endpoints.
func (s Span) bounds() (lo, hi Chron) {
	if s.Start.After(s.End.Time) {
		return s.End, s.Start
	}
	return s.Start, s.End
}

// Contains reports whether c is inside the closed interval [lo, hi].
func (s Span) Contains(c Chron) bool {
	lo, hi := s.bounds()
	return !c.Before(lo.Time) && !c.After(hi.Time)
}

// ContainsSpan reports whether other is fully inside this span.
func (s Span) ContainsSpan(other Span) bool {
	lo, hi := s.bounds()
	olo, ohi := other.bounds()
	return !olo.Before(lo.Time) && !hi.Before(ohi.Time)
}

// Overlaps reports whether this span shares any instant with other.
// Closed intervals that share only an endpoint overlap.
func (s Span) Overlaps(other Span) bool {
	lo, hi := s.bounds()
	olo, ohi := other.bounds()
	return !hi.Before(olo.Time) && !ohi.Before(lo.Time)
}

// Adjacent reports whether the spans are contiguous without overlapping.
// Contiguous means hi_a + 1ns == lo_b (or swapped). Sharing an endpoint overlaps.
func (s Span) Adjacent(other Span) bool {
	if s.Overlaps(other) {
		return false
	}
	lo, hi := s.bounds()
	olo, ohi := other.bounds()
	return hi.Add(Nanos(1)).Equal(olo.Time) || ohi.Add(Nanos(1)).Equal(lo.Time)
}

// Each yields instants from Start toward End by step (inclusive).
// The caller passes a positive step; reverse spans advance with Neg(step).
// A zero step yields nothing. A point span yields Start once.
func (s Span) Each(step Duration) iter.Seq[Chron] {
	return func(yield func(Chron) bool) {
		if step.IsZero() {
			return
		}
		if s.Start.Equal(s.End.Time) {
			yield(s.Start)
			return
		}

		dir := step
		reverse := s.Start.After(s.End.Time)
		if reverse {
			dir = step.Neg()
		}

		cur := s.Start
		if !yield(cur) {
			return
		}
		for {
			next := cur.Add(dir)
			progressed := next.After(cur.Time)
			if reverse {
				progressed = next.Before(cur.Time)
			}
			if !progressed || !s.Contains(next) {
				return
			}
			if !yield(next) {
				return
			}
			cur = next
		}
	}
}
