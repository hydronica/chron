package chron

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hydronica/trial"
)

func TestChron_Period(t *testing.T) {
	type input struct {
		c Chron
		p Period
	}
	type spanOut struct {
		Start string
		End   string
	}

	fn := func(in input) (spanOut, error) {
		s := in.c.Period(in.p)
		return spanOut{
			Start: s.Start.AsTime().UTC().Format(time.RFC3339Nano),
			End:   s.End.AsTime().UTC().Format(time.RFC3339Nano),
		}, nil
	}

	anchor := FromTime(time.Date(2026, 2, 4, 15, 30, 45, 123456789, time.UTC))
	mar31 := Date(2026, 3, 31)
	feb29 := Date(2024, 2, 29)

	cases := trial.Cases[input, spanOut]{
		"month to date": {
			Input: input{c: anchor, p: MonthToDate},
			Expected: spanOut{
				Start: "2026-02-01T00:00:00Z",
				End:   "2026-02-04T23:59:59.999999999Z",
			},
		},
		"period zero equals month to date": {
			Input: input{c: anchor, p: Period(0)},
			Expected: spanOut{
				Start: "2026-02-01T00:00:00Z",
				End:   "2026-02-04T23:59:59.999999999Z",
			},
		},
		"year to date": {
			Input: input{c: anchor, p: YearToDate},
			Expected: spanOut{
				Start: "2026-01-01T00:00:00Z",
				End:   "2026-02-04T23:59:59.999999999Z",
			},
		},
		"prev month from day 31": {
			Input: input{c: mar31, p: PrevMonth},
			Expected: spanOut{
				Start: "2026-02-01T00:00:00Z",
				End:   "2026-02-28T23:59:59.999999999Z",
			},
		},
		"prev month to date from day 31": {
			Input: input{c: mar31, p: PrevMonthToDate},
			Expected: spanOut{
				Start: "2026-02-01T00:00:00Z",
				End:   "2026-02-28T23:59:59.999999999Z",
			},
		},
		"prev year month to date from leap day": {
			Input: input{c: feb29, p: PrevYearMonthToDate},
			Expected: spanOut{
				Start: "2023-02-01T00:00:00Z",
				End:   "2023-02-28T23:59:59.999999999Z",
			},
		},
		"prev year to date from leap day": {
			Input: input{c: feb29, p: PrevYearToDate},
			Expected: spanOut{
				Start: "2023-01-01T00:00:00Z",
				End:   "2023-02-28T23:59:59.999999999Z",
			},
		},
		"last full week": {
			Input: input{c: anchor, p: LastFullWeek},
			Expected: spanOut{
				Start: "2026-01-26T00:00:00Z",
				End:   "2026-02-01T23:59:59.999999999Z",
			},
		},
		"unknown period uses month to date": {
			Input: input{c: anchor, p: Period(255)},
			Expected: spanOut{
				Start: "2026-02-01T00:00:00Z",
				End:   "2026-02-04T23:59:59.999999999Z",
			},
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestPeriod_MarshalJSON(t *testing.T) {
	fn := func(p Period) (string, error) {
		data, err := json.Marshal(p)
		return string(data), err
	}
	cases := trial.Cases[Period, string]{
		"month to date": {
			Input:    MonthToDate,
			Expected: `"month_to_date"`,
		},
		"year to date": {
			Input:    YearToDate,
			Expected: `"year_to_date"`,
		},
		"prev month": {
			Input:    PrevMonth,
			Expected: `"prev_month"`,
		},
		"prev month to date": {
			Input:    PrevMonthToDate,
			Expected: `"prev_month_to_date"`,
		},
		"prev year month to date": {
			Input:    PrevYearMonthToDate,
			Expected: `"prev_year_month_to_date"`,
		},
		"prev year to date": {
			Input:    PrevYearToDate,
			Expected: `"prev_year_to_date"`,
		},
		"last full week": {
			Input:    LastFullWeek,
			Expected: `"last_full_week"`,
		},
		"unknown value": {
			Input:       Period(255),
			ExpectedErr: errInvalidPeriod,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestPeriod_UnmarshalJSON(t *testing.T) {
	fn := func(raw string) (Period, error) {
		var p Period
		err := json.Unmarshal([]byte(raw), &p)
		return p, err
	}
	cases := trial.Cases[string, Period]{
		"null": {
			Input:    `null`,
			Expected: MonthToDate,
		},
		"month to date": {
			Input:    `"month_to_date"`,
			Expected: MonthToDate,
		},
		"year to date": {
			Input:    `"year_to_date"`,
			Expected: YearToDate,
		},
		"prev month": {
			Input:    `"prev_month"`,
			Expected: PrevMonth,
		},
		"prev month to date": {
			Input:    `"prev_month_to_date"`,
			Expected: PrevMonthToDate,
		},
		"prev year month to date": {
			Input:    `"prev_year_month_to_date"`,
			Expected: PrevYearMonthToDate,
		},
		"prev year to date": {
			Input:    `"prev_year_to_date"`,
			Expected: PrevYearToDate,
		},
		"last full week": {
			Input:    `"last_full_week"`,
			Expected: LastFullWeek,
		},
		"rejects integer": {
			Input:     `0`,
			ShouldErr: true,
		},
		"rejects unknown string": {
			Input:       `"fortnight"`,
			ExpectedErr: errInvalidPeriod,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestNthWeekday(t *testing.T) {
	type input struct {
		year    int
		month   time.Month
		weekday time.Weekday
		n       int
	}
	fn := func(in input) (string, error) {
		c, err := NthWeekday(in.year, in.month, in.weekday, in.n)
		return c.String(), err
	}
	cases := trial.Cases[input, string]{
		"third monday in january 2026": {
			Input:    input{year: 2026, month: time.January, weekday: time.Monday, n: 3},
			Expected: "2026-01-19",
		},
		"fifth monday in february 2026 errors": {
			Input:       input{year: 2026, month: time.February, weekday: time.Monday, n: 5},
			ExpectedErr: errInvalidWeekday,
		},
		"n less than one errors": {
			Input:       input{year: 2026, month: time.January, weekday: time.Monday, n: 0},
			ExpectedErr: errInvalidWeekday,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestLastWeekday(t *testing.T) {
	type input struct {
		year    int
		month   time.Month
		weekday time.Weekday
	}
	fn := func(in input) (string, error) {
		return LastWeekday(in.year, in.month, in.weekday).String(), nil
	}
	cases := trial.Cases[input, string]{
		"last monday in may 2026": {
			Input:    input{year: 2026, month: time.May, weekday: time.Monday},
			Expected: "2026-05-25",
		},
	}
	trial.New(fn, cases).SubTest(t)
}
