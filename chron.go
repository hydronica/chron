package chron

import (
	"errors"
	"time"
)

var (
	errInvalidSpan     = errors.New("chron: invalid span: start must be before end")
	errParseChron      = errors.New("chron: unable to parse time")
	errParseDuration   = errors.New("chron: unable to parse duration")
	errInvalidDuration = errors.New("chron: invalid duration")
)

// Chron is an instant in time with nanosecond storage precision.
// It embeds time.Time and adds chron-specific behavior.
type Chron struct {
	time.Time
	precision Precision
}

// Now returns the current instant in UTC with Nanosecond precision.
func Now() Chron {
	return FromTime(time.Now().UTC())
}

// FromTime wraps t, normalizing to UTC with Nanosecond precision.
func FromTime(t time.Time) Chron {
	return Chron{Time: t.UTC(), precision: Nanosecond}
}

// Date returns a calendar date at midnight UTC with Day precision.
func Date(y int, m time.Month, d int) Chron {
	return Chron{
		Time:      time.Date(y, m, d, 0, 0, 0, 0, time.UTC),
		precision: Day,
	}
}

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

// Precision returns the current precision metadata.
func (c Chron) Precision() Precision {
	return c.precision
}

// AsTime returns the underlying time.Time for third-party APIs.
func (c Chron) AsTime() time.Time {
	return c.Time
}

// InLocation returns the instant in loc for display only.
func (c Chron) InLocation(loc *time.Location) time.Time {
	return c.Time.In(loc)
}

// Truncate returns the inclusive start of the unit containing c and sets
// precision on the result to p (including MondayWeek and SundayWeek).
func (c Chron) Truncate(p Precision) Chron {
	return Chron{Time: Truncate(c.Time, p), precision: p}
}

// Add applies d relative to c and updates precision per offset rules.
func (c Chron) Add(d Duration) Chron {
	t := c.Time.AddDate(d.years, d.months, d.weeks*7+d.days).Add(d.clock)
	p := d.Precision()
	if c.precision == Nanosecond && d.clockOnly() {
		p = Nanosecond
	}
	return Chron{Time: t, precision: p}
}

// Sub subtracts d from c.
func (c Chron) Sub(d Duration) Chron {
	return c.Add(d.Neg())
}

// Span returns the half-open interval [start, end) containing c for precision p.
func (c Chron) Span(p Precision) Span {
	start := c.Truncate(p)
	end := start.Add(p.Duration())
	return Span{
		Start: start,
		End:   end,
	}
}

// Truncate returns the inclusive start of the unit containing t for precision p.
func Truncate(t time.Time, p Precision) time.Time {
	t = t.UTC()

	switch weekBoundary(p) {
	case Year:
		return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	case Month:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	case MondayWeek:
		return mondayWeekStart(t)
	case SundayWeek:
		return sundayWeekStart(t)
	case Day:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	case Hour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
	case Minute:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
	case Second:
		return t.Truncate(time.Second)
	case MicroSecond:
		return t.Truncate(time.Microsecond)
	case Millisecond:
		return t.Truncate(time.Millisecond)
	case Nanosecond:
		return t
	default:
		return t
	}
}

func mondayWeekStart(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	daysSinceMonday := (int(t.Weekday()) + 6) % 7
	return t.AddDate(0, 0, -daysSinceMonday)
}

func sundayWeekStart(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return t.AddDate(0, 0, -int(t.Weekday()))
}

// String returns the precision-aware canonical string form of c.
func (c Chron) String() string {
	if c.IsZero() {
		return ""
	}
	layout := layoutForPrecision(c.precision)
	if layout == time.RFC3339 || layout == time.RFC3339Nano {
		return c.Time.UTC().Format(layout)
	}
	return c.Time.UTC().Format(layout)
}
