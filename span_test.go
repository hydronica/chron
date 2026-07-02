package chron

import (
	"testing"
	"time"

	"github.com/hydronica/trial"
)

func TestChron_Span(t *testing.T) {
	type input struct {
		anchor    Chron
		precision Precision
	}
	type bounds struct {
		start     string
		end       string
		precision Precision
	}

	fn := func(in input) (bounds, error) {
		s := in.anchor.Span(in.precision)
		return bounds{
			start:     s.Start.String(),
			end:       s.End.String(),
			precision: s.Start.Precision(),
		}, nil
	}

	cases := trial.Cases[input, bounds]{
		"february month": {
			Input: input{anchor: Date(2026, 2, 15), precision: Month},
			Expected: bounds{
				start:     "2026-02",
				end:       "2026-03",
				precision: Month,
			},
		},
		"monday week": {
			Input: input{anchor: Date(2026, 2, 4), precision: MondayWeek},
			Expected: bounds{
				start:     "2026-02-02T00:00:00Z",
				end:       "2026-02-09T00:00:00Z",
				precision: MondayWeek,
			},
		},
		"sunday week": {
			Input: input{anchor: Date(2026, 2, 4), precision: SundayWeek},
			Expected: bounds{
				start:     "2026-02-01T00:00:00Z",
				end:       "2026-02-08T00:00:00Z",
				precision: SundayWeek,
			},
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestSpan_Contains(t *testing.T) {
	type input struct {
		span  Span
		chron Chron
	}

	fn := func(in input) (bool, error) {
		return in.span.Contains(in.chron), nil
	}

	cases := trial.Cases[input, bool]{
		"event inside month": {
			Input: input{
				span:  Date(2026, 2, 15).Span(Month),
				chron: FromTime(time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC)),
			},
			Expected: true,
		},
		"instant before month": {
			Input: input{
				span:  Date(2026, 2, 15).Span(Month),
				chron: FromTime(time.Date(2026, 1, 31, 23, 0, 0, 0, time.UTC)),
			},
			Expected: false,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestSpan_Overlaps(t *testing.T) {
	type input struct {
		a Span
		b Span
	}

	fn := func(in input) (bool, error) {
		return in.a.Overlaps(in.b), nil
	}

	cases := trial.Cases[input, bool]{
		"overlapping spans": {
			Input: input{
				a: MustSpan(Date(2026, 1, 1), Date(2026, 2, 1)),
				b: MustSpan(Date(2026, 1, 15), Date(2026, 3, 1)),
			},
			Expected: true,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestSpan_Adjacent(t *testing.T) {
	type input struct {
		a Span
		b Span
	}

	fn := func(in input) (bool, error) {
		return in.a.Adjacent(in.b), nil
	}

	cases := trial.Cases[input, bool]{
		"touching spans": {
			Input: input{
				a: MustSpan(Date(2026, 1, 1), Date(2026, 2, 1)),
				b: MustSpan(Date(2026, 2, 1), Date(2026, 3, 1)),
			},
			Expected: true,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestNewSpan(t *testing.T) {
	type input struct {
		start Chron
		end   Chron
	}

	fn := func(in input) (Span, error) {
		return NewSpan(in.start, in.end)
	}

	cases := trial.Cases[input, Span]{
		"start after end": {
			Input: input{
				start: Date(2026, 2, 1),
				end:   Date(2026, 1, 1),
			},
			ExpectedErr: errInvalidSpan,
		},
	}
	trial.New(fn, cases).SubTest(t)
}
