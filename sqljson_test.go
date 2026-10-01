package chron

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hydronica/trial"
)

func TestChron_UnmarshalJSON(t *testing.T) {
	fn := func(raw string) (string, error) {
		var c Chron
		err := json.Unmarshal([]byte(raw), &c)
		if err != nil {
			return "", err
		}
		return c.String(), nil
	}

	cases := trial.Cases[string, string]{
		"null": {
			Input:    `null`,
			Expected: "",
		},
		"day": {
			Input:    `"2026-02-04"`,
			Expected: "2026-02-04",
		},
		"RFC3339 zulu": {
			Input:    `"2026-02-04T15:30:45Z"`,
			Expected: "2026-02-04T15:30:45Z",
		},
		"RFC3339Nano": {
			Input:    `"2026-02-04T15:30:45.123456789Z"`,
			Expected: "2026-02-04T15:30:45.123456789Z",
		},
		"invalid": {
			Input:       `"not-a-date"`,
			ExpectedErr: errParseChron,
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestChron_Value(t *testing.T) {
	fn := func(c Chron) (any, error) {
		return c.Value()
	}

	cases := trial.Cases[Chron, any]{
		"zero": {
			Input:    Chron{},
			Expected: nil,
		},
		"day": {
			Input:    Date(2026, 2, 4),
			Expected: "2026-02-04",
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestChron_Scan(t *testing.T) {
	type input struct {
		src any
	}

	fn := func(in input) (string, error) {
		var c Chron
		if err := c.Scan(in.src); err != nil {
			return "", err
		}
		return c.String(), nil
	}

	cases := trial.Cases[input, string]{
		"nil": {
			Input:    input{src: nil},
			Expected: "",
		},
		"time.Time": {
			Input: input{
				src: time.Date(2026, 2, 4, 15, 30, 45, 0, time.UTC),
			},
			Expected: "2026-02-04T15:30:45Z",
		},
		"string day": {
			Input:    input{src: "2026-02-04"},
			Expected: "2026-02-04",
		},
		"bytes RFC3339": {
			Input:    input{src: []byte("2026-02-04T15:30:45Z")},
			Expected: "2026-02-04T15:30:45Z",
		},
	}
	trial.New(fn, cases).SubTest(t)
}
