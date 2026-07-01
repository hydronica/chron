package chron

import (
	"strings"
	"time"
)

// ParseFormats is the layout registry for Parse and UnmarshalJSON.
// Register narrowest layouts before broader ones when ambiguity matters.
var ParseFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15",
	"2006-01-02",
	"2006-01",
	"2006",
}

// Parse tries registered layouts and sets precision from the matched layout.
func Parse(s string) (Chron, error) {
	for _, layout := range ParseFormats {
		if c, err := ParseFrom(layout, s); err == nil {
			return c, nil
		}
	}
	return Chron{}, errParseChron
}

// ParseFrom parses s with an explicit layout and sets precision from the layout.
func ParseFrom(layout, s string) (Chron, error) {
	t, err := time.Parse(layout, s)
	if err != nil {
		return Chron{}, err
	}
	p := precisionFromLayout(layout)
	t = Truncate(t.UTC(), p)
	return Chron{Time: t, precision: p}, nil
}

func precisionFromLayout(layout string) Precision {
	switch layout {
	case "2006":
		return Year
	case "2006-01":
		return Month
	case "2006-01-02":
		return Day
	case "2006-01-02T15":
		return Hour
	case "2006-01-02T15:04:05":
		return Second
	case time.RFC3339:
		return Second
	case time.RFC3339Nano:
		return Nanosecond
	default:
		return precisionFromLayoutComponents(layout)
	}
}

func precisionFromLayoutComponents(layout string) Precision {
	hasYear := strings.Contains(layout, "2006")
	hasMonth := strings.Contains(layout, "01")
	hasDay := strings.Contains(layout, "02")
	hasHour := strings.Contains(layout, "15")
	hasMinute := strings.Contains(layout, "04")
	hasSecond := strings.Contains(layout, "05")
	hasFrac := strings.Contains(layout, "9") || strings.Contains(layout, "0")

	switch {
	case hasFrac:
		return Nanosecond
	case hasSecond:
		return Second
	case hasMinute:
		return Minute
	case hasHour:
		return Hour
	case hasDay:
		return Day
	case hasMonth:
		return Month
	case hasYear:
		return Year
	default:
		return Nanosecond
	}
}

func layoutForPrecision(p Precision) string {
	switch p {
	case Year:
		return "2006"
	case Month:
		return "2006-01"
	case Day:
		return "2006-01-02"
	case Hour:
		return "2006-01-02T15"
	case Minute, Second:
		return time.RFC3339
	case Millisecond, MicroSecond, Nanosecond, Week, MondayWeek, SundayWeek:
		return time.RFC3339Nano
	default:
		return time.RFC3339Nano
	}
}

func formatChron(c Chron) string {
	if c.IsZero() {
		return ""
	}
	layout := layoutForPrecision(c.precision)
	if layout == time.RFC3339 || layout == time.RFC3339Nano {
		return c.Time.UTC().Format(layout)
	}
	return c.Time.UTC().Format(layout)
}
