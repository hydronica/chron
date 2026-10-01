package chron

import (
	"encoding/json"
	"testing"

	"github.com/hydronica/trial"
)

func TestParseDuration(t *testing.T) {
	cases := trial.Cases[string, Duration]{
		"ISO 14 days": {
			Input:    "P14D",
			Expected: Days(14),
		},
		"bare 14 days": {
			Input:    "14d",
			Expected: Days(14),
		},
		"ISO 12 hours": {
			Input:    "PT12H",
			Expected: Hours(12),
		},
		"bare 12 hours": {
			Input:    "12h",
			Expected: Hours(12),
		},
		"ISO 2 weeks": {
			Input:    "P2W",
			Expected: Weeks(2),
		},
		"composite calendar and clock": {
			Input:    "P1Y3M4DT12H",
			Expected: Years(1).Months(3).Days(4).Hours(12),
		},
		"fractional seconds": {
			Input:    "PT0.001S",
			Expected: Millis(1),
		},
		"mixed clock units": {
			Input:    "pt4h500ms",
			Expected: Hours(4).Millis(500),
		},
		"microseconds ascii": {
			Input:    "250us",
			Expected: Micros(250),
		},
		"microseconds unicode": {
			Input:    "250µs",
			Expected: Micros(250),
		},
		"negative duration": {
			Input:    "-P1D",
			Expected: Days(-1),
		},
		"fractional year as months": {
			Input:    "P1.5Y",
			Expected: Months(18),
		},
		"week combined with year": {
			Input:    "P1Y1W",
			Expected: Years(1).Weeks(1),
		},
		"week combined with days": {
			Input:    "P2W1D",
			Expected: Weeks(2).Days(1),
		},
		"fractional month rejected": {
			Input:       "P1.5M",
			ExpectedErr: errInvalidDuration,
		},
	}
	trial.New(ParseDuration, cases).SubTest(t)
}

func TestDuration_String(t *testing.T) {
	fn := func(d Duration) (string, error) {
		return d.String(), nil
	}
	cases := trial.Cases[Duration, string]{
		"zero": {
			Input:    Duration{},
			Expected: "",
		},
		"two weeks": {
			Input:    Weeks(2),
			Expected: "p2w",
		},
		"fourteen days": {
			Input:    Days(14),
			Expected: "p2w",
		},
		"year and week": {
			Input:    Years(1).Weeks(1),
			Expected: "p1y1w",
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestDuration_MarshalJSON(t *testing.T) {
	type cfg struct {
		Length Duration `json:"length"`
	}

	fn := func(d Duration) (string, error) {
		data, err := json.Marshal(cfg{Length: d})
		return string(data), err
	}

	cases := trial.Cases[Duration, string]{
		"zero duration": {
			Input:    Duration{},
			Expected: `{"length":null}`,
		},
		"14 days": {
			Input:    Days(14),
			Expected: `{"length":"p2w"}`,
		},
		"250 microseconds": {
			Input:    Micros(250),
			Expected: `{"length":"pt250µs"}`,
		},
		"years decompose": {
			Input:    Years(1),
			Expected: `{"length":"p1y"}`,
		},
		"twelve months decompose to year": {
			Input:    Months(12),
			Expected: `{"length":"p1y"}`,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	fn := func(raw string) (Duration, error) {
		var d Duration
		err := json.Unmarshal([]byte(raw), &d)
		return d, err
	}

	cases := trial.Cases[string, Duration]{
		"null": {
			Input:    `null`,
			Expected: Duration{},
		},
		"empty string": {
			Input:    `""`,
			Expected: Duration{},
		},
		"14 days": {
			Input:    `"14d"`,
			Expected: Days(14),
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestDuration_Mul(t *testing.T) {
	type input struct {
		d Duration
		n int
	}
	fn := func(in input) (string, error) {
		return in.d.Mul(in.n).String(), nil
	}
	cases := trial.Cases[input, string]{
		"months times three": {
			Input:    input{d: Months(2), n: 3},
			Expected: "p6m",
		},
		"days times zero": {
			Input:    input{d: Days(5), n: 0},
			Expected: "",
		},
		"clock times two": {
			Input:    input{d: Hours(3), n: 2},
			Expected: "pt6h",
		},
		"negative scale": {
			Input:    input{d: Weeks(1), n: -2},
			Expected: "-p2w",
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestDuration_Neg(t *testing.T) {
	fn := func(d Duration) (string, error) {
		return d.Neg().String(), nil
	}
	cases := trial.Cases[Duration, string]{
		"negates months": {
			Input:    Months(6),
			Expected: "-p6m",
		},
		"negates clock": {
			Input:    Hours(2),
			Expected: "-pt2h",
		},
		"negates negative days": {
			Input:    Days(-3),
			Expected: "p3d",
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestDuration_Precision(t *testing.T) {
	fn := func(d Duration) (Precision, error) {
		return d.Precision(), nil
	}
	cases := trial.Cases[Duration, Precision]{
		"year": {
			Input:    Years(2),
			Expected: Year,
		},
		"month": {
			Input:    Months(3),
			Expected: Month,
		},
		"week": {
			Input:    Weeks(1),
			Expected: Week,
		},
		"day": {
			Input:    Days(2),
			Expected: Day,
		},
		"hour": {
			Input:    Hours(4),
			Expected: Hour,
		},
		"millisecond": {
			Input:    Millis(5),
			Expected: Millisecond,
		},
		"microsecond": {
			Input:    Micros(10),
			Expected: Microsecond,
		},
		"nanosecond": {
			Input:    Nanos(7),
			Expected: Nanosecond,
		},
		"calendar finest wins": {
			Input:    Years(1).Days(1),
			Expected: Day,
		},
	}
	trial.New(fn, cases).SubTest(t)
}
