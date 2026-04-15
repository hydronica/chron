package chron

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ScanTime implements sql.Scanner semantics for *time.Time (legacy Chron/Year Scan).
// If truncate is non-nil, it is applied after normalizing with TimeOf (e.g. YearOf for year columns).
func ScanTime(dst *time.Time, value any, truncate ...func(time.Time) time.Time) error {
	if dst == nil {
		return fmt.Errorf("ScanTime: nil destination")
	}
	if value == nil {
		*dst = ZeroValue()
		if len(truncate) > 0 && truncate[0] != nil {
			*dst = truncate[0](*dst)
		}
		return nil
	}
	if tt, ok := value.(time.Time); ok {
		*dst = TimeOf(tt)
		if len(truncate) > 0 && truncate[0] != nil {
			*dst = truncate[0](*dst)
		}
		return nil
	}
	return fmt.Errorf("unsupported Scan, storing %s into *time.Time", reflect.TypeOf(value))
}

// ValueTime implements driver.Valuer for time.Time (legacy Value).
func ValueTime(t time.Time) (driver.Value, error) {
	return InUTC(t), nil
}

// UnmarshalTimeJSON parses JSON string or null into *time.Time.
// If truncate is non-nil, the parsed instant is passed through truncate (e.g. YearOf).
func UnmarshalTimeJSON(dst *time.Time, data []byte, truncate func(time.Time) time.Time) error {
	if dst == nil {
		return fmt.Errorf("UnmarshalTimeJSON: nil destination")
	}
	if string(data) == "null" {
		*dst = ZeroValue()
		return nil
	}
	s := strings.Trim(string(data), `"`)
	t, err := Parse(s)
	if err != nil {
		*dst = ZeroValue()
		if truncate != nil {
			*dst = truncate(*dst)
		}
		return err
	}
	if truncate != nil {
		*dst = truncate(t)
	} else {
		*dst = TimeOf(t)
	}
	return nil
}

// MarshalTimeJSON encodes t as a JSON string (RFC3339Nano in UTC).
func MarshalTimeJSON(t time.Time) ([]byte, error) {
	return json.Marshal(InUTC(t).Format(time.RFC3339Nano))
}
