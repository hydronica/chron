package chron

import (
	"database/sql/driver"
	"testing"
	"time"

	"tools/dura"

	"github.com/stretchr/testify/assert"
)

// --- year ---

var tyear = time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)
var year = YearOf(tyear)

func TestNewYear(t *testing.T) {
	assert.Equal(t, tyear, NewYear(2018))
	assert.Equal(t, tyear, YearOf(tyear))
}

func TestThisYear(t *testing.T) {
	now := YearOf(time.Now())
	ty := ThisYear()
	assert.Equal(t, now, ty)
}

func TestYearTransfers(t *testing.T) {
	ch := ThisYear()
	assert.Equal(t, YearOf(ch), YearOf(ch))
}

func TestYearIncrement(t *testing.T) {
	y := Increment(year, dura.NewDuration(1, 2, 30, time.Second*500))
	td := tyear.AddDate(1, 2, 30).Add(time.Second * 500)
	assert.Exactly(t, td, y)
}

func TestYearDecrement(t *testing.T) {
	d := Decrement(year, dura.NewDuration(1, 2, 30, time.Second*500))
	td := tyear.AddDate(-1, -2, -30).Add(time.Second * -500)
	assert.Exactly(t, td, d)
}

func TestYearAddN(t *testing.T) {
	assert.Exactly(t, NewYear(2020), AddNYear(year, 2))
}

func TestYearSpan(t *testing.T) {
	sy := SpanOfYear(year)
	assert.Exactly(t, year, sy.Start())
	assert.Exactly(t, tyear.AddDate(1, 0, 0).Add(-time.Nanosecond), sy.End())
}

func TestYearContains(t *testing.T) {
	assert.True(t, SpanContains(SpanOfYear(year), SpanOfMinute(NewMinute(2018, time.February, 1, 12, 45))))
	assert.False(t, SpanContains(SpanOfYear(year), SpanOfYear(AddNYear(year, 1))))
}

func TestYearDurationUnit(t *testing.T) {
	assert.Equal(t, dura.Year.Years(), dura.Year.Years())
}

func TestYearAddFns(t *testing.T) {
	assert.Exactly(t, AddYearsYear(year, 2), NewYear(2020))
	assert.Exactly(t, AddMonthsYear(year, 25), NewMonth(2020, time.February))
	assert.Exactly(t, AddDaysYear(year, 2), NewDay(2018, time.January, 3))
	assert.Exactly(t, AddHoursYear(year, 25), NewHour(2018, time.January, 2, 1))
	assert.Exactly(t, AddMinutesYear(year, 72), NewMinute(2018, time.January, 1, 1, 12))
	assert.Exactly(t, AddSecondsYear(year, 3672), NewSecond(2018, time.January, 1, 1, 1, 12))
	assert.Exactly(t, AddMillisYear(year, 3672001), NewMilli(2018, time.January, 1, 1, 1, 12, 1))
	assert.Exactly(t, AddMicrosYear(year, 3672000001), NewMicro(2018, time.January, 1, 1, 1, 12, 1))
	assert.Exactly(t, AddNanosYear(year, 3672000000001), NewTime(2018, time.January, 1, 1, 1, 12, 1))
}

func TestYearScanValueJSON(t *testing.T) {
	yearZero := YearOf(ZeroValue())
	var y time.Time
	assert.Nil(t, ScanTime(&y, tyear, YearOf))
	assert.Exactly(t, year, y)
	assert.Nil(t, ScanTime(&y, nil, YearOf))
	assert.Exactly(t, yearZero, y)
	assert.Error(t, ScanTime(&y, "wrong value", YearOf))

	v, err := ValueTime(year)
	assert.Nil(t, err)
	assert.Exactly(t, driver.Value(tyear), v)

	assert.Nil(t, UnmarshalTimeJSON(&y, []byte("null"), YearOf))
	assert.Exactly(t, yearZero, y)
	assert.Error(t, UnmarshalTimeJSON(&y, []byte("as;dlkjfd"), YearOf))
	assert.Exactly(t, yearZero, y)
	assert.Nil(t, UnmarshalTimeJSON(&y, []byte("\"2018-02-01T00:00:00Z\""), YearOf))
	assert.Exactly(t, year, y)
}

// --- month ---

var tmonth = time.Date(2018, time.April, 1, 0, 0, 0, 0, time.UTC)
var month = MonthOf(tmonth)

func TestNewMonth(t *testing.T) {
	assert.Equal(t, tmonth, NewMonth(2018, 4))
	assert.Equal(t, tmonth, MonthOf(tmonth))
}

func TestThisMonth(t *testing.T) {
	now := MonthOf(time.Now())
	ty := ThisMonth()
	assert.Equal(t, now, ty)
}

func TestMonthIncrement(t *testing.T) {
	y := Increment(month, dura.NewDuration(1, 2, 30, time.Second*500))
	td := tmonth.AddDate(1, 2, 30).Add(time.Second * 500)
	assert.Exactly(t, td, y)
}

func TestMonthAddN(t *testing.T) {
	assert.Exactly(t, NewMonth(2018, time.June), AddNMonth(month, 2))
}

func TestMonthSpan(t *testing.T) {
	sm := SpanOfMonth(month)
	assert.Exactly(t, month, sm.Start())
	assert.Exactly(t, tmonth.AddDate(0, 1, 0).Add(-time.Nanosecond), sm.End())
}

func TestMonthContains(t *testing.T) {
	assert.True(t, SpanContains(SpanOfMonth(month), SpanOfMinute(NewMinute(2018, time.April, 5, 12, 45))))
	assert.False(t, SpanContains(SpanOfMonth(month), SpanOfMonth(AddNMonth(month, 1))))
}

func TestMonthAddFns(t *testing.T) {
	assert.Exactly(t, AddYearsMonth(month, 2), NewMonth(2020, time.April))
	assert.Exactly(t, AddMonthsMonth(month, 25), NewMonth(2020, time.May))
	assert.Exactly(t, AddDaysMonth(month, 2), NewDay(2018, time.April, 3))
	assert.Exactly(t, AddHoursMonth(month, 25), NewHour(2018, time.April, 2, 1))
	assert.Exactly(t, AddMinutesMonth(month, 72), NewMinute(2018, time.April, 1, 1, 12))
	assert.Exactly(t, AddSecondsMonth(month, 3672), NewSecond(2018, time.April, 1, 1, 1, 12))
	assert.Exactly(t, AddMillisMonth(month, 3672001), NewMilli(2018, time.April, 1, 1, 1, 12, 1))
	assert.Exactly(t, AddMicrosMonth(month, 3672000001), NewMicro(2018, time.April, 1, 1, 1, 12, 1))
	assert.Exactly(t, AddNanosMonth(month, 3672000000001), NewTime(2018, time.April, 1, 1, 1, 12, 1))
}

func TestMonthScanValueJSON(t *testing.T) {
	var m time.Time
	assert.Nil(t, ScanTime(&m, tmonth, MonthOf))
	assert.Exactly(t, month, m)
	assert.Nil(t, ScanTime(&m, nil, MonthOf))
	assert.Exactly(t, MonthOf(ZeroValue()), m)
	assert.Nil(t, UnmarshalTimeJSON(&m, []byte("\"2018-04-01T00:00:00Z\""), MonthOf))
	assert.Exactly(t, month, m)
}

// --- day ---

var tday = time.Date(2018, time.February, 1, 0, 0, 0, 0, time.UTC)
var day = DayOf(tday)

func TestNewDay(t *testing.T) {
	assert.Equal(t, tday, NewDay(2018, time.February, 1))
	assert.Equal(t, tday, DayOf(tday))
}

func TestToday(t *testing.T) {
	now := DayOf(time.Now())
	today := Today()
	assert.Equal(t, now, today)
}

func TestDayIncrement(t *testing.T) {
	d := Increment(day, dura.Duration{Yrs: 1, Mons: 2, Dys: 30, Dur: time.Second * 500})
	td := tday.AddDate(1, 2, 30).Add(time.Second * 500)
	assert.Exactly(t, td, d)
}

func TestDayAddN(t *testing.T) {
	assert.Exactly(t, NewDay(2018, time.February, 3), AddNDay(day, 2))
}

func TestDaySpan(t *testing.T) {
	sd := SpanOfDay(day)
	assert.Exactly(t, day, sd.Start())
	assert.Exactly(t, tday.Add(24*time.Hour-time.Nanosecond), sd.End())
}

func TestDayContains(t *testing.T) {
	assert.True(t, SpanContains(SpanOfDay(day), SpanOfMinute(NewMinute(2018, time.February, 1, 12, 45))))
	assert.False(t, SpanContains(SpanOfDay(day), SpanOfDay(AddNDay(day, 1))))
}

func TestDayAddFns(t *testing.T) {
	assert.Exactly(t, AddYearsDay(day, 2), NewDay(2020, time.February, 1))
	assert.Exactly(t, AddMonthsDay(day, 24), NewDay(2020, time.February, 1))
	assert.Exactly(t, AddDaysDay(day, 2), AddNDay(day, 2))
	assert.Exactly(t, AddHoursDay(day, 25), NewHour(2018, time.February, 2, 1))
	assert.Exactly(t, AddMinutesDay(day, 72), NewMinute(2018, time.February, 1, 1, 12))
	assert.Exactly(t, AddSecondsDay(day, 3672), NewSecond(2018, time.February, 1, 1, 1, 12))
	assert.Exactly(t, AddMillisDay(day, 3672001), NewMilli(2018, time.February, 1, 1, 1, 12, 1))
	assert.Exactly(t, AddMicrosDay(day, 3672000001), NewMicro(2018, time.February, 1, 1, 1, 12, 1))
	assert.Exactly(t, AddNanosDay(day, 3672000000001), NewTime(2018, time.February, 1, 1, 1, 12, 1))
}

func TestDayScanJSON(t *testing.T) {
	var d time.Time
	assert.Nil(t, ScanTime(&d, tday, DayOf))
	assert.Exactly(t, day, d)
	assert.Nil(t, UnmarshalTimeJSON(&d, []byte("\"2018-02-01T00:00:00Z\""), DayOf))
	assert.Exactly(t, day, d)
}

// --- hour ---

var thour = time.Date(2018, time.June, 5, 12, 0, 0, 0, time.UTC)
var hour = HourOf(thour)

func TestNewHour(t *testing.T) {
	assert.Equal(t, thour, NewHour(2018, time.June, 5, 12))
	assert.Equal(t, thour, HourOf(thour))
}

func TestThisHour(t *testing.T) {
	now := HourOf(time.Now())
	ty := ThisHour()
	assert.Equal(t, now, ty)
}

func TestHourAddN(t *testing.T) {
	assert.Exactly(t, NewHour(2018, time.June, 5, 14), AddNHour(hour, 2))
}

func TestHourSpan(t *testing.T) {
	sh := SpanOfHour(hour)
	assert.Exactly(t, hour, sh.Start())
	assert.Exactly(t, thour.Add(time.Hour-time.Nanosecond), sh.End())
}

func TestHourContains(t *testing.T) {
	assert.True(t, SpanContains(SpanOfHour(hour), SpanOfMinute(NewMinute(2018, time.June, 5, 12, 45))))
	assert.False(t, SpanContains(SpanOfHour(hour), SpanOfHour(AddNHour(hour, 1))))
}

func TestHourAddFns(t *testing.T) {
	assert.Exactly(t, AddYearsHour(hour, 2), NewHour(2020, time.June, 5, 12))
	assert.Exactly(t, AddMonthsHour(hour, 25), NewHour(2020, time.July, 5, 12))
	assert.Exactly(t, AddDaysHour(hour, 2), NewHour(2018, time.June, 7, 12))
	assert.Exactly(t, AddHoursHour(hour, 25), NewHour(2018, time.June, 6, 13))
	assert.Exactly(t, AddMinutesHour(hour, 72), NewMinute(2018, time.June, 5, 13, 12))
	assert.Exactly(t, AddSecondsHour(hour, 3672), NewSecond(2018, time.June, 5, 13, 1, 12))
	assert.Exactly(t, AddMillisHour(hour, 3672001), NewMilli(2018, time.June, 5, 13, 1, 12, 1))
	assert.Exactly(t, AddMicrosHour(hour, 3672000001), NewMicro(2018, time.June, 5, 13, 1, 12, 1))
	assert.Exactly(t, AddNanosHour(hour, 3672000000001), NewTime(2018, time.June, 5, 13, 1, 12, 1))
}

func TestHourScanJSON(t *testing.T) {
	var h time.Time
	assert.Nil(t, ScanTime(&h, thour, HourOf))
	assert.Exactly(t, hour, h)
	assert.Nil(t, UnmarshalTimeJSON(&h, []byte("\"2018-06-05T12:00:00Z\""), HourOf))
	assert.Exactly(t, hour, h)
}

// --- minute ---

var tmin = time.Date(2018, time.June, 5, 12, 10, 0, 0, time.UTC)
var min = MinuteOf(tmin)

func TestNewMinute(t *testing.T) {
	assert.Equal(t, tmin, NewMinute(2018, time.June, 5, 12, 10))
	assert.Equal(t, tmin, MinuteOf(tmin))
}

func TestMinuteAddN(t *testing.T) {
	assert.Exactly(t, NewMinute(2018, time.June, 5, 12, 12), AddNMinute(min, 2))
}

func TestMinuteSpan(t *testing.T) {
	sm := SpanOfMinute(min)
	assert.Exactly(t, tmin.Add(time.Minute-time.Nanosecond), sm.End())
}

func TestMinuteContains(t *testing.T) {
	assert.True(t, SpanContains(SpanOfMinute(min), SpanOfSecond(NewSecond(2018, time.June, 5, 12, 10, 10))))
	assert.False(t, SpanContains(SpanOfMinute(min), SpanOfMinute(AddNMinute(min, 1))))
}

func TestMinuteAddFns(t *testing.T) {
	assert.Exactly(t, AddYearsMinute(min, 2), NewMinute(2020, time.June, 5, 12, 10))
	assert.Exactly(t, AddMonthsMinute(min, 25), NewMinute(2020, time.July, 5, 12, 10))
	assert.Exactly(t, AddDaysMinute(min, 2), NewMinute(2018, time.June, 7, 12, 10))
	assert.Exactly(t, AddHoursMinute(min, 25), NewMinute(2018, time.June, 6, 13, 10))
	assert.Exactly(t, AddMinutesMinute(min, 72), NewMinute(2018, time.June, 5, 13, 22))
	assert.Exactly(t, AddSecondsMinute(min, 3672), NewSecond(2018, time.June, 5, 13, 11, 12))
	assert.Exactly(t, AddMillisMinute(min, 3672001), NewMilli(2018, time.June, 5, 13, 11, 12, 1))
	assert.Exactly(t, AddMicrosMinute(min, 3672000001), NewMicro(2018, time.June, 5, 13, 11, 12, 1))
	assert.Exactly(t, AddNanosMinute(min, 3672000000001), NewTime(2018, time.June, 5, 13, 11, 12, 1))
}

func TestMinuteScanJSON(t *testing.T) {
	var m time.Time
	assert.Nil(t, ScanTime(&m, tmin, MinuteOf))
	assert.Nil(t, UnmarshalTimeJSON(&m, []byte("\"2018-06-05T12:10:00Z\""), MinuteOf))
	assert.Exactly(t, min, m)
}

// --- second ---

var tsec = time.Date(2018, time.June, 5, 12, 10, 6, 0, time.UTC)
var sec = SecondOf(tsec)

func TestSecondAddN(t *testing.T) {
	assert.Exactly(t, NewSecond(2018, time.June, 5, 12, 10, 8), AddNSecond(sec, 2))
}

func TestSecondContains(t *testing.T) {
	assert.True(t, SpanContains(SpanOfSecond(sec), InstantSpan(NewTime(2018, time.June, 5, 12, 10, 6, 234973))))
	assert.False(t, SpanContains(SpanOfSecond(sec), SpanOfSecond(AddNSecond(sec, 1))))
}

func TestSecondAddFns(t *testing.T) {
	assert.Exactly(t, AddYearsSecond(sec, 2), NewSecond(2020, time.June, 5, 12, 10, 6))
	assert.Exactly(t, AddMonthsSecond(sec, 25), NewSecond(2020, time.July, 5, 12, 10, 6))
	assert.Exactly(t, AddDaysSecond(sec, 2), NewSecond(2018, time.June, 7, 12, 10, 6))
	assert.Exactly(t, AddHoursSecond(sec, 25), NewSecond(2018, time.June, 6, 13, 10, 6))
	assert.Exactly(t, AddMinutesSecond(sec, 72), NewSecond(2018, time.June, 5, 13, 22, 6))
	assert.Exactly(t, AddSecondsSecond(sec, 3672), NewSecond(2018, time.June, 5, 13, 11, 18))
	assert.Exactly(t, AddMillisSecond(sec, 3672001), NewMilli(2018, time.June, 5, 13, 11, 18, 1))
	assert.Exactly(t, AddMicrosSecond(sec, 3672000001), NewMicro(2018, time.June, 5, 13, 11, 18, 1))
	assert.Exactly(t, AddNanosSecond(sec, 3672000000001), NewTime(2018, time.June, 5, 13, 11, 18, 1))
}

func TestSecondScanJSON(t *testing.T) {
	var m time.Time
	assert.Nil(t, ScanTime(&m, tsec, SecondOf))
	assert.Nil(t, UnmarshalTimeJSON(&m, []byte("\"2018-06-05T12:10:06Z\""), SecondOf))
	assert.Exactly(t, sec, m)
}

// --- milli ---

var tmilli = time.Date(2018, time.June, 5, 12, 10, 6, 55000000, time.UTC)
var milli = MilliOf(tmilli)

func TestMilliAddN(t *testing.T) {
	assert.Exactly(t, NewMilli(2018, time.June, 5, 12, 10, 6, 58), AddNMilli(milli, 3))
}

func TestMilliContains(t *testing.T) {
	assert.True(t, SpanContains(SpanOfMilli(milli), InstantSpan(NewTime(2018, time.June, 5, 12, 10, 6, 55000456))))
	assert.False(t, SpanContains(SpanOfMilli(milli), SpanOfMilli(AddNMilli(milli, 1))))
}

func TestMilliAddFns(t *testing.T) {
	assert.Exactly(t, AddYearsMilli(milli, 2), NewMilli(2020, time.June, 5, 12, 10, 6, 55))
	assert.Exactly(t, AddMonthsMilli(milli, 25), NewMilli(2020, time.July, 5, 12, 10, 6, 55))
	assert.Exactly(t, AddDaysMilli(milli, 2), NewMilli(2018, time.June, 7, 12, 10, 6, 55))
	assert.Exactly(t, AddHoursMilli(milli, 25), NewMilli(2018, time.June, 6, 13, 10, 6, 55))
	assert.Exactly(t, AddMinutesMilli(milli, 72), NewMilli(2018, time.June, 5, 13, 22, 6, 55))
	assert.Exactly(t, AddSecondsMilli(milli, 3672), NewMilli(2018, time.June, 5, 13, 11, 18, 55))
	assert.Exactly(t, AddMillisMilli(milli, 3672001), NewMilli(2018, time.June, 5, 13, 11, 18, 56))
	assert.Exactly(t, AddMicrosMilli(milli, 3672000001), NewMicro(2018, time.June, 5, 13, 11, 18, 55001))
	assert.Exactly(t, AddNanosMilli(milli, 3672000000001), NewTime(2018, time.June, 5, 13, 11, 18, 55000001))
}

func TestMilliScanJSON(t *testing.T) {
	var m time.Time
	assert.Nil(t, ScanTime(&m, tmilli, MilliOf))
	assert.Nil(t, UnmarshalTimeJSON(&m, []byte("\"2018-06-05T12:10:06.055Z\""), MilliOf))
	assert.Exactly(t, milli, m)
}

// --- micro ---

var tmicro = time.Date(2018, time.June, 5, 12, 10, 6, 55000, time.UTC)
var micro = MicroOf(tmicro)

func TestMicroAddN(t *testing.T) {
	assert.Exactly(t, NewMicro(2018, time.June, 5, 12, 10, 6, 58), AddNMicro(micro, 3))
}

func TestMicroContains(t *testing.T) {
	assert.True(t, SpanContains(SpanOfMicro(micro), InstantSpan(NewTime(2018, time.June, 5, 12, 10, 6, 55456))))
	assert.False(t, SpanContains(SpanOfMicro(micro), SpanOfMicro(AddNMicro(micro, 1))))
}

func TestMicroAddFns(t *testing.T) {
	assert.Exactly(t, AddYearsMicro(micro, 2), NewMicro(2020, time.June, 5, 12, 10, 6, 55))
	assert.Exactly(t, AddMonthsMicro(micro, 25), NewMicro(2020, time.July, 5, 12, 10, 6, 55))
	assert.Exactly(t, AddDaysMicro(micro, 2), NewMicro(2018, time.June, 7, 12, 10, 6, 55))
	assert.Exactly(t, AddHoursMicro(micro, 25), NewMicro(2018, time.June, 6, 13, 10, 6, 55))
	assert.Exactly(t, AddMinutesMicro(micro, 72), NewMicro(2018, time.June, 5, 13, 22, 6, 55))
	assert.Exactly(t, AddSecondsMicro(micro, 3672), NewMicro(2018, time.June, 5, 13, 11, 18, 55))
	assert.Exactly(t, AddMillisMicro(micro, 3672001), NewMicro(2018, time.June, 5, 13, 11, 18, 1055))
	assert.Exactly(t, AddMicrosMicro(micro, 3672000001), NewMicro(2018, time.June, 5, 13, 11, 18, 56))
	assert.Exactly(t, AddNanosMicro(micro, 3672000000001), NewTime(2018, time.June, 5, 13, 11, 18, 55001))
}

func TestMicroScanJSON(t *testing.T) {
	var m time.Time
	assert.Nil(t, ScanTime(&m, tmicro, MicroOf))
	assert.Nil(t, UnmarshalTimeJSON(&m, []byte("\"2018-06-05T12:10:06.000055Z\""), MicroOf))
	assert.Exactly(t, micro, m)
}
