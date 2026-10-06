package chron

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hydronica/trial"
)

// testAnchor is the shared instant for Truncate and Add table tests.
var testAnchor = FromTime(time.Date(2026, 2, 4, 15, 30, 45, 123456789, time.UTC))

func TestChron_Truncate(t *testing.T) {
	fn := func(p Precision) (string, error) {
		return testAnchor.Truncate(p).String(), nil
	}

	cases := trial.Cases[Precision, string]{
		"year": {
			Input:    Year,
			Expected: "2026",
		},
		"month": {
			Input:    Month,
			Expected: "2026-02",
		},
		"week uses default week start": {
			Input:    Week,
			Expected: "2026-02-02T00:00:00Z",
		},
		"monday week": {
			Input:    MondayWeek,
			Expected: "2026-02-02T00:00:00Z",
		},
		"sunday week": {
			Input:    SundayWeek,
			Expected: "2026-02-01T00:00:00Z",
		},
		"day": {
			Input:    Day,
			Expected: "2026-02-04",
		},
		"hour": {
			Input:    Hour,
			Expected: "2026-02-04T15",
		},
		"minute": {
			Input:    Minute,
			Expected: "2026-02-04T15:30:00Z",
		},
		"second": {
			Input:    Second,
			Expected: "2026-02-04T15:30:45Z",
		},
		"millisecond": {
			Input:    Millisecond,
			Expected: "2026-02-04T15:30:45.123Z",
		},
		"microsecond": {
			Input:    Microsecond,
			Expected: "2026-02-04T15:30:45.123456Z",
		},
		"nanosecond": {
			Input:    Nanosecond,
			Expected: "2026-02-04T15:30:45.123456789Z",
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestChron_Add(t *testing.T) {
	day := testAnchor.Truncate(Day)
	month := testAnchor.Truncate(Month)

	type addCase struct {
		base Chron
		d    Duration
	}
	type addOut struct {
		S string // String() — precision-aware form
		T string // AsTime UTC RFC3339Nano — calendar day for clamp cases
	}

	fn := func(c addCase) (addOut, error) {
		got := c.base.Add(c.d)
		return addOut{
			S: got.String(),
			T: got.AsTime().UTC().Format(time.RFC3339Nano),
		}, nil
	}

	cases := trial.Cases[addCase, addOut]{
		"year": {
			Input:    addCase{base: testAnchor, d: Years(1)},
			Expected: addOut{S: "2027-02-04T15:30:45.123456789Z", T: "2027-02-04T15:30:45.123456789Z"},
		},
		"month": {
			Input:    addCase{base: testAnchor, d: Months(1)},
			Expected: addOut{S: "2026-03-04T15:30:45.123456789Z", T: "2026-03-04T15:30:45.123456789Z"},
		},
		"week": {
			Input:    addCase{base: day, d: Weeks(1)},
			Expected: addOut{S: "2026-02-11", T: "2026-02-11T00:00:00Z"},
		},
		"day": {
			Input:    addCase{base: testAnchor, d: Days(5)},
			Expected: addOut{S: "2026-02-09T15:30:45.123456789Z", T: "2026-02-09T15:30:45.123456789Z"},
		},
		"hour": {
			Input:    addCase{base: day, d: Hours(3)},
			Expected: addOut{S: "2026-02-04T03", T: "2026-02-04T03:00:00Z"},
		},
		"minute": {
			Input:    addCase{base: day, d: Minutes(15)},
			Expected: addOut{S: "2026-02-04T00:15:00Z", T: "2026-02-04T00:15:00Z"},
		},
		"second": {
			Input:    addCase{base: day, d: Seconds(30)},
			Expected: addOut{S: "2026-02-04T00:00:30Z", T: "2026-02-04T00:00:30Z"},
		},
		"millisecond": {
			Input:    addCase{base: day, d: Millis(250)},
			Expected: addOut{S: "2026-02-04T00:00:00.25Z", T: "2026-02-04T00:00:00.25Z"},
		},
		"microsecond": {
			Input:    addCase{base: day, d: Micros(500)},
			Expected: addOut{S: "2026-02-04T00:00:00.0005Z", T: "2026-02-04T00:00:00.0005Z"},
		},
		"nanosecond": {
			Input:    addCase{base: day, d: Nanos(100)},
			Expected: addOut{S: "2026-02-04T00:00:00.0000001Z", T: "2026-02-04T00:00:00.0000001Z"},
		},
		"month anchor plus days": {
			Input:    addCase{base: month, d: Days(5)},
			Expected: addOut{S: "2026-02-06", T: "2026-02-06T00:00:00Z"},
		},
		"month plus month keeps month": {
			Input:    addCase{base: month, d: Months(1)},
			Expected: addOut{S: "2026-03", T: "2026-03-01T00:00:00Z"},
		},
		"finer of nanosecond receiver and clock offset": {
			Input:    addCase{base: testAnchor, d: Hours(2)},
			Expected: addOut{S: "2026-02-04T17:30:45.123456789Z", T: "2026-02-04T17:30:45.123456789Z"},
		},
		// Month add clamps to last day of target month (never AddDate overflow).
		// Date is Day precision; Months is coarser → result stays Day.
		"jan 31 plus one month clamps to feb 28": {
			Input:    addCase{base: Date(2026, 1, 31), d: Months(1)},
			Expected: addOut{S: "2026-02-28", T: "2026-02-28T00:00:00Z"},
		},
		"jan 31 plus one month in leap year clamps to feb 29": {
			Input:    addCase{base: Date(2024, 1, 31), d: Months(1)},
			Expected: addOut{S: "2024-02-29", T: "2024-02-29T00:00:00Z"},
		},
		"mar 31 plus one month clamps to apr 30": {
			Input:    addCase{base: Date(2026, 3, 31), d: Months(1)},
			Expected: addOut{S: "2026-04-30", T: "2026-04-30T00:00:00Z"},
		},
		"mar 31 minus one month clamps to feb 28": {
			Input:    addCase{base: Date(2026, 3, 31), d: Months(-1)},
			Expected: addOut{S: "2026-02-28", T: "2026-02-28T00:00:00Z"},
		},
		"feb 29 plus one year clamps to feb 28": {
			Input:    addCase{base: Date(2024, 2, 29), d: Years(1)},
			Expected: addOut{S: "2025-02-28", T: "2025-02-28T00:00:00Z"},
		},
		"mar 3 minus one month": {
			Input:    addCase{base: Date(2025, 3, 3), d: Months(-1)},
			Expected: addOut{S: "2025-02-03", T: "2025-02-03T00:00:00Z"},
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestParse(t *testing.T) {
	fn := func(in string) (string, error) {
		c, err := Parse(in)
		return c.String(), err
	}

	cases := trial.Cases[string, string]{
		"year": {
			Input:    "2026",
			Expected: "2026",
		},
		"month": {
			Input:    "2026-02",
			Expected: "2026-02",
		},
		"day": {
			Input:    "2026-02-04",
			Expected: "2026-02-04",
		},
		"hour": {
			Input:    "2026-02-04T15",
			Expected: "2026-02-04T15",
		},
		"second without timezone": {
			Input:    "2026-02-04T15:30:45",
			Expected: "2026-02-04T15:30:45Z",
		},
		"RFC3339 zulu": {
			Input:    "2026-02-04T15:30:45Z",
			Expected: "2026-02-04T15:30:45Z",
		},
		"RFC3339Nano": {
			Input:    "2026-02-04T15:30:45.123456789Z",
			Expected: "2026-02-04T15:30:45.123456789Z",
		},
		"invalid input": {
			Input:       "not-a-date",
			ExpectedErr: errParseChron,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestParseFrom(t *testing.T) {
	type input struct {
		layout string
		value  string
	}
	fn := func(in input) (Chron, error) {
		return ParseFrom(in.layout, in.value)
	}
	cases := trial.Cases[input, Chron]{
		"registered day layout": {
			Input:    input{layout: "2006-01-02", value: "2026-02-04"},
			Expected: Chron{Time: trial.Day("2026-02-04"), precision: Day},
		},
		"custom layout defaults to nanosecond": {
			Input:    input{layout: "2006-01-02T15:04", value: "2026-02-04T15:30"},
			Expected: Chron{Time: time.Date(2026, 02, 04, 15, 30, 0, 0, time.UTC), precision: Minute},
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestChron_Sub(t *testing.T) {
	c := Date(2026, 3, 1).Sub(Months(1))
	want := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !c.Equal(want) {
		t.Fatalf("Date(2026,3,1).Sub(Months(1)) = %v, want %v", c.AsTime(), want)
	}
}

func TestChron_MarshalJSON(t *testing.T) {
	fn := func(c Chron) (string, error) {
		data, err := json.Marshal(c)
		return string(data), err
	}

	cases := trial.Cases[Chron, string]{
		"zero value": {
			Input:    Chron{},
			Expected: "null",
		},
		"day": {
			Input:    Date(2026, 2, 4),
			Expected: `"2026-02-04"`,
		},
		"month": {
			Input:    testAnchor.Truncate(Month),
			Expected: `"2026-02"`,
		},
		"year": {
			Input:    testAnchor.Truncate(Year),
			Expected: `"2026"`,
		},
		"hour": {
			Input:    testAnchor.Truncate(Hour),
			Expected: `"2026-02-04T15"`,
		},
		"second": {
			Input:    testAnchor.Truncate(Second),
			Expected: `"2026-02-04T15:30:45Z"`,
		},
		"millisecond": {
			Input:    testAnchor.Truncate(Millisecond),
			Expected: `"2026-02-04T15:30:45.123Z"`,
		},
		"nanosecond": {
			Input:    testAnchor,
			Expected: `"2026-02-04T15:30:45.123456789Z"`,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

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
			end:       s.End.AsTime().UTC().Format(time.RFC3339Nano),
			precision: s.Start.Precision(),
		}, nil
	}

	cases := trial.Cases[input, bounds]{
		"february month": {
			Input: input{anchor: Date(2026, 2, 15), precision: Month},
			Expected: bounds{
				start:     "2026-02",
				end:       "2026-02-28T23:59:59.999999999Z",
				precision: Month,
			},
		},
		"monday week": {
			Input: input{anchor: Date(2026, 2, 4), precision: MondayWeek},
			Expected: bounds{
				start:     "2026-02-02T00:00:00Z",
				end:       "2026-02-08T23:59:59.999999999Z",
				precision: MondayWeek,
			},
		},
		"sunday week": {
			Input: input{anchor: Date(2026, 2, 4), precision: SundayWeek},
			Expected: bounds{
				start:     "2026-02-01T00:00:00Z",
				end:       "2026-02-07T23:59:59.999999999Z",
				precision: SundayWeek,
			},
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestChron_EndOf(t *testing.T) {
	type input struct {
		anchor Chron
		p      Precision
	}

	fn := func(in input) (string, error) {
		return in.anchor.EndOf(in.p).String(), nil
	}

	cases := trial.Cases[input, string]{
		// Day/Month/Year precision .String() drops the sub-unit nanoseconds;
		// Week precisions keep RFC3339Nano because no shorter canonical layout exists.
		"day": {
			Input:    input{anchor: testAnchor, p: Day},
			Expected: "2026-02-04T23:59:59.999999999Z",
		},
		"hour": {
			Input:    input{anchor: testAnchor, p: Hour},
			Expected: "2026-02-04T15:59:59.999999999Z",
		},
		"minute": {
			Input:    input{anchor: testAnchor, p: Minute},
			Expected: "2026-02-04T15:30:59.999999999Z",
		},
		"second": {
			Input:    input{anchor: testAnchor, p: Second},
			Expected: "2026-02-04T15:30:45.999999999Z",
		},
		"millisecond": {
			Input:    input{anchor: testAnchor, p: Millisecond},
			Expected: "2026-02-04T15:30:45.123999999Z",
		},
		"microsecond": {
			Input:    input{anchor: testAnchor, p: Microsecond},
			Expected: "2026-02-04T15:30:45.123456999Z",
		},
		"nanosecond": {
			Input:    input{anchor: testAnchor, p: Nanosecond},
			Expected: "2026-02-04T15:30:45.123456789Z",
		},
		"week default monday": {
			Input:    input{anchor: testAnchor, p: Week},
			Expected: "2026-02-08T23:59:59.999999999Z",
		},
		"monday week": {
			Input:    input{anchor: testAnchor, p: MondayWeek},
			Expected: "2026-02-08T23:59:59.999999999Z",
		},
		"sunday week": {
			Input:    input{anchor: testAnchor, p: SundayWeek},
			Expected: "2026-02-07T23:59:59.999999999Z",
		},
		"month february non-leap": {
			Input:    input{anchor: Date(2026, 2, 15), p: Month},
			Expected: "2026-02-28",
		},
		"month february leap": {
			Input:    input{anchor: Date(2024, 2, 15), p: Month},
			Expected: "2024-02-29",
		},
		"month 31-day": {
			Input:    input{anchor: Date(2026, 1, 10), p: Month},
			Expected: "2026-01-31",
		},
		"year": {
			Input:    input{anchor: testAnchor, p: Year},
			Expected: "2026-12-31T23:59:59.999999999Z",
		},
	}
	trial.New(fn, cases).SubTest(t)
}
