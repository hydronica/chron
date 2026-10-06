package chron

import (
	"encoding/json"
)

// Period names a closed reporting window relative to a Chron.
// It is not a Precision scale: there is no Less or Duration, and iota order
// is not fine-to-coarse. MonthToDate is the zero value.
type Period uint8

const (
	MonthToDate Period = iota // zero value; an unset Period is month-to-date
	YearToDate
	PrevMonth
	PrevMonthToDate
	PrevYearMonthToDate
	PrevYearToDate
	LastFullWeek
)

func (p Period) String() string {
	switch p {
	case MonthToDate:
		return "month_to_date"
	case YearToDate:
		return "year_to_date"
	case PrevMonth:
		return "prev_month"
	case PrevMonthToDate:
		return "prev_month_to_date"
	case PrevYearMonthToDate:
		return "prev_year_month_to_date"
	case PrevYearToDate:
		return "prev_year_to_date"
	case LastFullWeek:
		return "last_full_week"
	default:
		return ""
	}
}

// MarshalJSON encodes p as a JSON string (e.g. "month_to_date"), not an integer.
func (p Period) MarshalJSON() ([]byte, error) {
	s := p.String()
	if s == "" {
		return nil, errInvalidPeriod
	}
	return json.Marshal(s)
}

// UnmarshalJSON decodes a JSON string into p. Null maps to MonthToDate.
func (p *Period) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*p = MonthToDate
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, ok := periodFromString(s)
	if !ok {
		return errInvalidPeriod
	}
	*p = parsed
	return nil
}

func periodFromString(s string) (Period, bool) {
	switch s {
	case "month_to_date":
		return MonthToDate, true
	case "year_to_date":
		return YearToDate, true
	case "prev_month":
		return PrevMonth, true
	case "prev_month_to_date":
		return PrevMonthToDate, true
	case "prev_year_month_to_date":
		return PrevYearMonthToDate, true
	case "prev_year_to_date":
		return PrevYearToDate, true
	case "last_full_week":
		return LastFullWeek, true
	default:
		return MonthToDate, false
	}
}
