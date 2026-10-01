package chron

import (
	"encoding/json"
	"fmt"
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
	fn := func(src any) (string, error) {
		var c Chron
		if err := c.Scan(src); err != nil {
			return "", err
		}
		return c.String(), nil
	}

	cases := trial.Cases[any, string]{
		"nil": {
			Input:    nil,
			Expected: "",
		},
		"time.Time": {
			Input:    time.Date(2026, 2, 4, 15, 30, 45, 0, time.UTC),
			Expected: "2026-02-04T15:30:45Z",
		},
		"string day": {
			Input:    "2026-02-04",
			Expected: "2026-02-04",
		},
		"bytes RFC3339": {
			Input:    []byte("2026-02-04T15:30:45Z"),
			Expected: "2026-02-04T15:30:45Z",
		},
		"unsupported type": {
			Input:       42,
			ExpectedErr: fmt.Errorf("chron: cannot scan int into Chron"),
		},
	}
	trial.New(fn, cases).SubTest(t)
}
