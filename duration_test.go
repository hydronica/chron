package chron

import (
	"encoding/json"
	"testing"

	"github.com/hydronica/trial"
)

func durationEqual(a, b Duration) bool {
	anchor := Date(2020, 6, 15)
	return anchor.Add(a).Equal(anchor.Add(b).Time)
}

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
		"fractional year": {
			Input:       "P1.5Y",
			ExpectedErr: errInvalidDuration,
		},
		"week combined with year": {
			Input:       "P1Y1W",
			ExpectedErr: errInvalidDuration,
		},
	}
	trial.New(ParseDuration, cases).SubTest(t)
}

func TestDurationJSON(t *testing.T) {
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
			Expected: `{"length":"p0d"}`,
		},
		"14 days": {
			Input:    Days(14),
			Expected: `{"length":"p14d"}`,
		},
		"250 microseconds": {
			Input:    Micros(250),
			Expected: `{"length":"pt250µs"}`,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestDurationNegMul(t *testing.T) {
	d := Months(2).Mul(3)
	if d.String() != "p6m" {
		t.Fatalf("Months(2).Mul(3).String() = %q, want p6m", d.String())
	}
	if !durationEqual(d.Neg(), Months(-6)) {
		t.Fatal("Months(2).Mul(3).Neg() mismatch, want Months(-6)")
	}
}
