package chron

import (
	"errors"
	"time"
)

var (
	errParseChron       = errors.New("chron: unable to parse time")
	errParseDuration    = errors.New("chron: unable to parse duration")
	errInvalidPrecision = errors.New("chron: invalid precision")
	errInvalidPeriod    = errors.New("chron: invalid period")
	errInvalidWeekday   = errors.New("chron: invalid weekday")
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

// NthWeekday returns the nth occurrence of weekday in month of year (1-based).
// n < 1, or an n past the last matching weekday in that month, returns
// errInvalidWeekday. The result is midnight UTC with Day precision.
func NthWeekday(year int, month time.Month, weekday time.Weekday, n int) (Chron, error) {
	if n < 1 {
		return Chron{}, errInvalidWeekday
	}
	d := Date(year, month, 1)
	for d.Month() == month {
		if d.Weekday() == weekday {
			n--
			if n == 0 {
				return d, nil
			}
		}
		d = d.Add(Days(1))
	}
	return Chron{}, errInvalidWeekday
}

// LastWeekday returns the last occurrence of weekday in month of year.
// The result is midnight UTC with Day precision.
func LastWeekday(year int, month time.Month, weekday time.Weekday) Chron {
	d := Date(year, month, 1).Add(Months(1)).Sub(Days(1))
	for d.Weekday() != weekday {
		d = d.Sub(Days(1))
	}
	return d
}

// ParseFormats is the layout registry for Parse and UnmarshalJSON.
// Register narrowest layouts before broader ones when ambiguity matters.
// Mutate only during package init, before concurrent Parse calls.
var ParseFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
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

// Truncate returns the inclusive start of the unit containing c and sets
// precision on the result to p (including MondayWeek and SundayWeek).
func (c Chron) Truncate(p Precision) Chron {
	return Chron{Time: Truncate(c.Time, p), precision: p}
}

// EndOf returns the inclusive end of the unit containing c for precision p.
// The result is the last nanosecond of that unit (UTC).
// p selects the unit; the result keeps c's precision.
func (c Chron) EndOf(p Precision) Chron {
	return Chron{Time: endOf(c.Time, p), precision: c.precision}
}

// Add applies d relative to c and updates precision per offset rules.
// Months are clamped to the last day of the target month (Jan 31 + 1 month →
// Feb 28/29), unlike time.Time.AddDate which overflows into the following month.
// Days and clock are applied after months.
// Result precision is the finer of c's precision and d.Precision() (e.g. Day +
// Months keeps Day; Day + Seconds becomes Second).
func (c Chron) Add(d Duration) Chron {
	t := c.Time
	if d.months != 0 {
		t = addMonthsClamp(t, int(d.months))
	}
	if d.days != 0 {
		t = t.AddDate(0, 0, int(d.days))
	}
	if d.clock != 0 {
		t = t.Add(d.clock)
	}
	p := d.Precision()
	if c.precision.Less(p) {
		p = c.precision
	}
	return Chron{Time: t, precision: p}
}

// addMonthsClamp shifts t by months, clamping the day to the last day of the
// destination month when the source day does not exist there.
func addMonthsClamp(t time.Time, months int) time.Time {
	year, month, day := t.Date()
	h, min, sec := t.Clock()
	nsec := t.Nanosecond()

	am := year*12 + int(month-1) + months
	year = am / 12
	mo := am % 12
	if mo < 0 {
		mo += 12
		year--
	}
	dest := time.Month(mo + 1)
	last := time.Date(year, dest+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > last {
		day = last
	}
	return time.Date(year, dest, day, h, min, sec, nsec, t.Location())
}

// Sub subtracts d from c.
func (c Chron) Sub(d Duration) Chron {
	return c.Add(d.Neg())
}

// Span returns the closed interval [start, end] containing c for precision p.
func (c Chron) Span(p Precision) Span {
	return NewSpan(c.Truncate(p), c.EndOf(p))
}

// Period returns the closed reporting window for p relative to c.
// To-date windows end at EndOf(Day). An out-of-range Period uses MonthToDate.
// Week windows follow DefaultWeekStart, the same as Truncate(Week).
func (c Chron) Period(p Period) Span {
	switch p {
	case YearToDate:
		return NewSpan(c.Truncate(Year), c.EndOf(Day))
	case PrevMonth:
		return c.Truncate(Month).Sub(Months(1)).Span(Month)
	case PrevMonthToDate:
		start := c.Truncate(Month).Sub(Months(1))
		end := c.Truncate(Day).Sub(Months(1)).EndOf(Day)
		return NewSpan(start, end)
	case PrevYearMonthToDate:
		end := c.Truncate(Day).Sub(Years(1)).EndOf(Day)
		return NewSpan(end.Truncate(Month), end)
	case PrevYearToDate:
		end := c.Truncate(Day).Sub(Years(1)).EndOf(Day)
		return NewSpan(end.Truncate(Year), end)
	case LastFullWeek:
		return c.Truncate(Week).Sub(Weeks(1)).Span(Week)
	default: // MonthToDate and unknown Period values
		return NewSpan(c.Truncate(Month), c.EndOf(Day))
	}
}

// Truncate returns the inclusive start of the unit containing t for precision p.
func Truncate(t time.Time, p Precision) time.Time {
	t = t.UTC()

	switch p.weekBoundary() {
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
	case Millisecond:
		return t.Truncate(time.Millisecond)
	case Microsecond:
		return t.Truncate(time.Microsecond)
	default: // Nanosecond: or unknown Precision values
		return t
	}
}

// endOf returns the last nanosecond of the unit containing t for precision p.
func endOf(t time.Time, p Precision) time.Time {
	start := Truncate(t, p)
	switch p.weekBoundary() {
	case Year:
		return time.Date(start.Year(), time.December, 31, 23, 59, 59, 999999999, time.UTC)
	case Month:
		return start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case MondayWeek, SundayWeek:
		return start.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case Day:
		return time.Date(start.Year(), start.Month(), start.Day(), 23, 59, 59, 999999999, time.UTC)
	case Hour:
		return start.Add(time.Hour).Add(-time.Nanosecond)
	case Minute:
		return start.Add(time.Minute).Add(-time.Nanosecond)
	case Second:
		return start.Add(time.Second).Add(-time.Nanosecond)
	case Millisecond:
		return start.Add(time.Millisecond).Add(-time.Nanosecond)
	case Microsecond:
		return start.Add(time.Microsecond).Add(-time.Nanosecond)
	default: //  Nanosecond or unknown Precision values
		return start
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
	return c.Time.UTC().Format(c.precision.layout())
}
