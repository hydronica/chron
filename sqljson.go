package chron

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// MarshalJSON encodes c as a precision-aware ISO string, or JSON null when zero.
func (c Chron) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(c.String())
}

// UnmarshalJSON decodes a JSON string into c using registered parse layouts.
func (c *Chron) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*c = Chron{}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := Parse(s)
	if err != nil {
		parsed, err = ParseFrom(time.RFC3339, s)
		if err != nil {
			parsed, err = ParseFrom(time.RFC3339Nano, s)
			if err != nil {
				return err
			}
		}
	}
	*c = parsed
	return nil
}

// Value implements driver.Valuer for database/sql.
func (c Chron) Value() (driver.Value, error) {
	if c.IsZero() {
		return nil, nil
	}
	return c.String(), nil
}

// Scan implements sql.Scanner for database/sql.
func (c *Chron) Scan(src any) error {
	if src == nil {
		*c = Chron{}
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		*c = FromTime(v)
		return nil
	case string:
		return c.scanString(v)
	case []byte:
		return c.scanString(string(v))
	default:
		return fmt.Errorf("chron: cannot scan %T into Chron", src)
	}
}

func (c *Chron) scanString(s string) error {
	parsed, err := Parse(s)
	if err != nil {
		parsed, err = ParseFrom(time.RFC3339, s)
		if err != nil {
			parsed, err = ParseFrom(time.RFC3339Nano, s)
			if err != nil {
				return err
			}
		}
	}
	*c = parsed
	return nil
}
