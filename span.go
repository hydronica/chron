package chron

// Span is a half-open interval [Start, End) in UTC.
// End is exclusive: the first instant NOT in the span.
type Span struct {
	Start Chron
	End   Chron
}

// NewSpan builds an arbitrary range. It errors if start is not before end.
func NewSpan(start, end Chron) (Span, error) {
	if !start.Before(end.Time) {
		return Span{}, errInvalidSpan
	}
	return Span{Start: start, End: end}, nil
}

// MustSpan panics if the range is invalid.
func MustSpan(start, end Chron) Span {
	s, err := NewSpan(start, end)
	if err != nil {
		panic(err)
	}
	return s
}

// IsZero reports whether both endpoints are zero.
func (s Span) IsZero() bool {
	return s.Start.IsZero() && s.End.IsZero()
}

// Contains reports whether c is inside [Start, End).
func (s Span) Contains(c Chron) bool {
	return !c.Before(s.Start.Time) && c.Before(s.End.Time)
}

// ContainsSpan reports whether other is fully inside this span.
func (s Span) ContainsSpan(other Span) bool {
	return !other.Start.Before(s.Start.Time) && !s.End.Before(other.End.Time)
}

// Overlaps reports whether this span shares any instant with other.
func (s Span) Overlaps(other Span) bool {
	return s.Start.Before(other.End.Time) && other.Start.Before(s.End.Time)
}

// Adjacent reports whether other starts exactly at this End or vice versa.
func (s Span) Adjacent(other Span) bool {
	return s.End.Equal(other.Start.Time) || other.End.Equal(s.Start.Time)
}
