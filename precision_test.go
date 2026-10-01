package chron

import (
	"testing"

	"github.com/hydronica/trial"
)

func TestPrecision_Less(t *testing.T) {
	type input struct {
		p1 Precision
		p2 Precision
	}
	fn := func(in input) (bool, error) {
		return in.p1.Less(in.p2), nil
	}
	cases := trial.Cases[input, bool]{
		"same precision": {
			Input:    input{p1: Year, p2: Year},
			Expected: false,
		},
		"month less than year": {
			Input:    input{p1: Month, p2: Year},
			Expected: true,
		},
		"week less than month": {
			Input:    input{p1: Week, p2: Month},
			Expected: true,
		},
		"monday week less than week": {
			Input:    input{p1: MondayWeek, p2: Week},
			Expected: true,
		},
		"sunday week not less than monday week": {
			Input:    input{p1: SundayWeek, p2: MondayWeek},
			Expected: false,
		},
		"day less than sunday week": {
			Input:    input{p1: Day, p2: SundayWeek},
			Expected: true,
		},
		"hour less than day": {
			Input:    input{p1: Hour, p2: Day},
			Expected: true,
		},
		"minute less than hour": {
			Input:    input{p1: Minute, p2: Hour},
			Expected: true,
		},
		"second less than minute": {
			Input:    input{p1: Second, p2: Minute},
			Expected: true,
		},
		"millisecond less than second": {
			Input:    input{p1: Millisecond, p2: Second},
			Expected: true,
		},
		"microsecond less than millisecond": {
			Input:    input{p1: Microsecond, p2: Millisecond},
			Expected: true,
		},
		"nanosecond less than microsecond": {
			Input:    input{p1: Nanosecond, p2: Microsecond},
			Expected: true,
		},
		"nanosecond less than millisecond": {
			Input:    input{p1: Nanosecond, p2: Millisecond},
			Expected: true,
		},
		"nanosecond less than year": {
			Input:    input{p1: Nanosecond, p2: Year},
			Expected: true,
		},
		"year not less than month": {
			Input:    input{p1: Year, p2: Month},
			Expected: false,
		},
		"year not less than nanosecond": {
			Input:    input{p1: Year, p2: Nanosecond},
			Expected: false,
		},
		"millisecond not less than microsecond": {
			Input:    input{p1: Millisecond, p2: Microsecond},
			Expected: false,
		},
		"millisecond not less than nanosecond": {
			Input:    input{p1: Millisecond, p2: Nanosecond},
			Expected: false,
		},
		"hour not less than minute": {
			Input:    input{p1: Hour, p2: Minute},
			Expected: false,
		},
		"day not less than hour": {
			Input:    input{p1: Day, p2: Hour},
			Expected: false,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestPrecision_Duration(t *testing.T) {
	fn := func(p Precision) (string, error) {
		return p.Duration().String(), nil
	}
	cases := trial.Cases[Precision, string]{
		"year": {
			Input:    Year,
			Expected: "p1y",
		},
		"month": {
			Input:    Month,
			Expected: "p1m",
		},
		"week": {
			Input:    Week,
			Expected: "p1w",
		},
		"day": {
			Input:    Day,
			Expected: "p1d",
		},
		"hour": {
			Input:    Hour,
			Expected: "pt1h",
		},
		"microsecond": {
			Input:    Microsecond,
			Expected: "pt1µs",
		},
		"nanosecond": {
			Input:    Nanosecond,
			Expected: "pt1ns",
		},
	}
	trial.New(fn, cases).SubTest(t)
}
