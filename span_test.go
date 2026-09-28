package chron

import (
	"testing"
	"time"

	"github.com/hydronica/trial"
)

func TestSpan_Contains(t *testing.T) {
	type input struct {
		span  Span
		chron Chron
	}

	fn := func(in input) (bool, error) {
		return in.span.Contains(in.chron), nil
	}

	feb := Date(2026, 2, 15).Span(Month)
	point := NewSpan(Date(2026, 2, 1), Date(2026, 2, 1))
	reverse := NewSpan(Date(2026, 3, 1), Date(2026, 1, 1))

	cases := trial.Cases[input, bool]{
		"event inside month": {
			Input: input{
				span:  feb,
				chron: FromTime(time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC)),
			},
			Expected: true,
		},
		"last instant of month": {
			Input: input{
				span:  feb,
				chron: feb.End,
			},
			Expected: true,
		},
		"march first not in february": {
			Input: input{
				span:  feb,
				chron: Date(2026, 3, 1),
			},
			Expected: false,
		},
		"instant before month": {
			Input: input{
				span:  feb,
				chron: FromTime(time.Date(2026, 1, 31, 23, 0, 0, 0, time.UTC)),
			},
			Expected: false,
		},
		"point span contains itself": {
			Input:    input{span: point, chron: Date(2026, 2, 1)},
			Expected: true,
		},
		"point span excludes other day": {
			Input:    input{span: point, chron: Date(2026, 2, 2)},
			Expected: false,
		},
		"reverse span contains middle": {
			Input: input{
				span:  reverse,
				chron: Date(2026, 2, 1),
			},
			Expected: true,
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
				a: NewSpan(Date(2026, 1, 1), Date(2026, 2, 1)),
				b: NewSpan(Date(2026, 1, 15), Date(2026, 3, 1)),
			},
			Expected: true,
		},
		"shared endpoint overlaps": {
			Input: input{
				a: NewSpan(Date(2026, 1, 1), Date(2026, 2, 1)),
				b: NewSpan(Date(2026, 2, 1), Date(2026, 3, 1)),
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
		"contiguous months": {
			Input: input{
				a: Date(2026, 1, 15).Span(Month),
				b: Date(2026, 2, 15).Span(Month),
			},
			Expected: true,
		},
		"shared endpoint not adjacent": {
			Input: input{
				a: NewSpan(Date(2026, 1, 1), Date(2026, 2, 1)),
				b: NewSpan(Date(2026, 2, 1), Date(2026, 3, 1)),
			},
			Expected: false,
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
		return NewSpan(in.start, in.end), nil
	}

	cases := trial.Cases[input, Span]{
		"reverse span": {
			Input: input{
				start: Date(2026, 2, 1),
				end:   Date(2026, 1, 1),
			},
			Expected: Span{Start: Date(2026, 2, 1), End: Date(2026, 1, 1)},
		},
		"point span": {
			Input: input{
				start: Date(2026, 2, 1),
				end:   Date(2026, 2, 1),
			},
			Expected: Span{Start: Date(2026, 2, 1), End: Date(2026, 2, 1)},
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestSpanBetween(t *testing.T) {
	got := SpanBetween(Date(2026, 3, 1), Date(2026, 1, 1))
	want := Span{Start: Date(2026, 1, 1), End: Date(2026, 3, 1)}
	if got != want {
		t.Fatalf("SpanBetween(Mar1, Jan1) = %+v, want %+v", got, want)
	}
}

func TestSpan_Each(t *testing.T) {
	type input struct {
		span Span
		step Duration
	}

	fn := func(in input) ([]string, error) {
		var out []string
		for c := range in.span.Each(in.step) {
			out = append(out, c.String())
		}
		return out, nil
	}

	cases := trial.Cases[input, []string]{
		"days through january": {
			Input: input{
				span: NewSpan(Date(2026, 1, 1), Date(2026, 1, 3)),
				step: Days(1),
			},
			Expected: []string{"2026-01-01", "2026-01-02", "2026-01-03"},
		},
		"point span": {
			Input: input{
				span: NewSpan(Date(2026, 1, 1), Date(2026, 1, 1)),
				step: Days(1),
			},
			Expected: []string{"2026-01-01"},
		},
		"reverse days": {
			Input: input{
				span: NewSpan(Date(2026, 1, 3), Date(2026, 1, 1)),
				step: Days(1),
			},
			Expected: []string{"2026-01-03", "2026-01-02", "2026-01-01"},
		},
		"zero step": {
			Input: input{
				span: NewSpan(Date(2026, 1, 1), Date(2026, 1, 3)),
				step: Duration{},
			},
			Expected: nil,
		},
	}
	trial.New(fn, cases).SubTest(t)
}
