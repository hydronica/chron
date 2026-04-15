package chron

import "time"

// --- constructors (UTC) ---

func NewYear(year int) time.Time {
	return time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func NewMonth(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

func NewDay(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func NewHour(year int, month time.Month, day, hour int) time.Time {
	return time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
}

func NewMinute(year int, month time.Month, day, hour, min int) time.Time {
	return time.Date(year, month, day, hour, min, 0, 0, time.UTC)
}

func NewSecond(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}

func NewMilli(year int, month time.Month, day, hour, min, sec, milli int) time.Time {
	return time.Date(year, month, day, hour, min, sec, milli*1e6, time.UTC)
}

func NewMicro(year int, month time.Month, day, hour, min, sec, micro int) time.Time {
	return time.Date(year, month, day, hour, min, sec, micro*1e3, time.UTC)
}

// --- truncations (UTC) ---

func YearOf(t time.Time) time.Time {
	t = t.UTC()
	return NewYear(t.Year())
}

func MonthOf(t time.Time) time.Time {
	t = t.UTC()
	return NewMonth(t.Year(), t.Month())
}

func DayOf(t time.Time) time.Time {
	t = t.UTC()
	return NewDay(t.Year(), t.Month(), t.Day())
}

func HourOf(t time.Time) time.Time {
	t = t.UTC()
	return NewHour(t.Year(), t.Month(), t.Day(), t.Hour())
}

func MinuteOf(t time.Time) time.Time {
	t = t.UTC()
	return NewMinute(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}

func SecondOf(t time.Time) time.Time {
	t = t.UTC()
	return t.Truncate(time.Second).UTC()
}

func MilliOf(t time.Time) time.Time {
	t = t.UTC()
	return t.Truncate(time.Millisecond).UTC()
}

func MicroOf(t time.Time) time.Time {
	t = t.UTC()
	return t.Truncate(time.Microsecond).UTC()
}

// --- “this” helpers ---

func ThisYear() time.Time  { return YearOf(time.Now()) }
func ThisMonth() time.Time { return MonthOf(time.Now()) }
func Today() time.Time     { return DayOf(time.Now()) }
func ThisHour() time.Time  { return HourOf(time.Now()) }
func ThisMinute() time.Time {
	return MinuteOf(time.Now())
}
func ThisSecond() time.Time {
	return SecondOf(time.Now())
}
func ThisMilli() time.Time {
	return MilliOf(time.Now())
}
func ThisMicro() time.Time {
	return MicroOf(time.Now())
}
