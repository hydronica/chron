package chron

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hydronica/trial"
)

// testAnchor is the shared instant for Truncate and Add table tests.
var testAnchor = FromTime(time.Date(2026, 2, 4, 15, 30, 45, 123456789, time.UTC))

func TestTruncate(t *testing.T) {
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
			Input:    MicroSecond,
			Expected: "2026-02-04T15:30:45.123456Z",
		},
		"nanosecond": {
			Input:    Nanosecond,
			Expected: "2026-02-04T15:30:45.123456789Z",
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestAdd(t *testing.T) {
	day := testAnchor.Truncate(Day)
	month := testAnchor.Truncate(Month)

	type addCase struct {
		base Chron
		d    Duration
	}

	fn := func(c addCase) (string, error) {
		return c.base.Add(c.d).String(), nil
	}

	cases := trial.Cases[addCase, string]{
		"year": {
			Input:    addCase{base: testAnchor, d: Years(1)},
			Expected: "2027",
		},
		"month": {
			Input:    addCase{base: testAnchor, d: Months(1)},
			Expected: "2026-03",
		},
		"week": {
			Input:    addCase{base: day, d: Weeks(1)},
			Expected: "2026-02-11T00:00:00Z",
		},
		"day": {
			Input:    addCase{base: testAnchor, d: Days(5)},
			Expected: "2026-02-09",
		},
		"hour": {
			Input:    addCase{base: day, d: Hours(3)},
			Expected: "2026-02-04T03",
		},
		"minute": {
			Input:    addCase{base: day, d: Minutes(15)},
			Expected: "2026-02-04T00:15:00Z",
		},
		"second": {
			Input:    addCase{base: day, d: Seconds(30)},
			Expected: "2026-02-04T00:00:30Z",
		},
		"millisecond": {
			Input:    addCase{base: day, d: Millis(250)},
			Expected: "2026-02-04T00:00:00.25Z",
		},
		"microsecond": {
			Input:    addCase{base: day, d: Micros(500)},
			Expected: "2026-02-04T00:00:00.0005Z",
		},
		"nanosecond": {
			Input:    addCase{base: day, d: Nanos(100)},
			Expected: "2026-02-04T00:00:00.0000001Z",
		},
		"month anchor plus days": {
			Input:    addCase{base: month, d: Days(5)},
			Expected: "2026-02-06",
		},
		"nanosecond receiver clock-only keeps nanosecond": {
			Input:    addCase{base: testAnchor, d: Hours(2)},
			Expected: "2026-02-04T17:30:45.123456789Z",
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

func TestSub(t *testing.T) {
	c := Date(2026, 3, 1).Sub(Months(1))
	want := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !c.Equal(want) {
		t.Fatalf("Date(2026,3,1).Sub(Months(1)) = %v, want %v", c.AsTime(), want)
	}
}

func TestChronMarshalJSON(t *testing.T) {
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
