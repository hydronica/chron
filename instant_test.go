package chron

import (
	"database/sql/driver"
	"testing"
	"time"

	"tools/dura"

	"github.com/stretchr/testify/assert"
)

var tnano = time.Date(2018, time.June, 5, 12, 10, 6, 55, time.UTC)
var chr = TimeOf(tnano)

func TestNewNano(t *testing.T) {
	assert.Equal(t, tnano, NewTime(2018, time.June, 5, 12, 10, 6, 55))
	assert.Equal(t, tnano, TimeOf(tnano))
}

func TestChronTransfers(t *testing.T) {
	ch := Now()
	assert.Equal(t, YearOf(ch), YearOf(ch))
	assert.Equal(t, MonthOf(ch), MonthOf(ch))
	assert.Equal(t, DayOf(ch), DayOf(ch))
	assert.Equal(t, HourOf(ch), HourOf(ch))
	assert.Equal(t, MinuteOf(ch), MinuteOf(ch))
	assert.Equal(t, SecondOf(ch), SecondOf(ch))
	assert.Equal(t, MilliOf(ch), MilliOf(ch))
	assert.Equal(t, MicroOf(ch), MicroOf(ch))
	var c Chron = ch
	_ = c
}

func TestChronIncrement(t *testing.T) {
	y := Increment(chr, dura.NewDuration(1, 2, 30, time.Nanosecond*500))
	td := tnano.AddDate(1, 2, 30).Add(time.Nanosecond * 500)
	assert.Exactly(t, td, y)
}

func TestChronDecrement(t *testing.T) {
	d := Decrement(chr, dura.NewDuration(1, 2, 30, time.Nanosecond*500))
	td := tnano.AddDate(-1, -2, -30).Add(time.Nanosecond * -500)
	assert.Exactly(t, td, d)
}

func TestChronAddN(t *testing.T) {
	assert.Exactly(t, NewTime(2018, time.June, 5, 12, 10, 6, 58), AddN(chr, 3))
}

func TestChronInstantSpan(t *testing.T) {
	assert.Exactly(t, chr, InstantSpan(chr).Start())
	assert.Exactly(t, chr, InstantSpan(chr).End())
}

func TestChronSpanContains(t *testing.T) {
	assert.True(t, SpanContains(InstantSpan(chr), InstantSpan(NewTime(2018, time.June, 5, 12, 10, 6, 55))))
	assert.False(t, SpanContains(InstantSpan(chr), InstantSpan(AddN(chr, 1))))
}

func TestChronAddFns(t *testing.T) {
	assert.Exactly(t, AddYearsChron(chr, 2), NewTime(2020, time.June, 5, 12, 10, 6, 55))
	assert.Exactly(t, AddMonthsChron(chr, 25), NewTime(2020, time.July, 5, 12, 10, 6, 55))
	assert.Exactly(t, AddDaysChron(chr, 2), NewTime(2018, time.June, 7, 12, 10, 6, 55))
	assert.Exactly(t, AddHoursChron(chr, 25), NewTime(2018, time.June, 6, 13, 10, 6, 55))
	assert.Exactly(t, AddMinutesChron(chr, 72), NewTime(2018, time.June, 5, 13, 22, 6, 55))
	assert.Exactly(t, AddSecondsChron(chr, 3672), NewTime(2018, time.June, 5, 13, 11, 18, 55))
	assert.Exactly(t, AddMillisChron(chr, 3672001), NewTime(2018, time.June, 5, 13, 11, 18, 1000055))
	assert.Exactly(t, AddMicrosChron(chr, 3672000001), NewTime(2018, time.June, 5, 13, 11, 18, 1055))
	assert.Exactly(t, AddNanosChron(chr, 3672000000001), NewTime(2018, time.June, 5, 13, 11, 18, 56))
}

func TestChronScanValueJSON(t *testing.T) {
	var m time.Time
	assert.Nil(t, ScanTime(&m, tnano))
	assert.Exactly(t, chr, m)
	assert.Nil(t, ScanTime(&m, nil))
	assert.Exactly(t, ZeroValue(), m)

	assert.Error(t, ScanTime(&m, "wrong value"))

	v, err := ValueTime(chr)
	assert.Nil(t, err)
	assert.Exactly(t, driver.Value(tnano), v)

	assert.Nil(t, UnmarshalTimeJSON(&m, []byte("null"), nil))
	assert.Exactly(t, ZeroValue(), m)
	assert.Error(t, UnmarshalTimeJSON(&m, []byte("as;dlkjfd"), nil))
	assert.Exactly(t, ZeroValue(), m)
	assert.Nil(t, UnmarshalTimeJSON(&m, []byte("\"2018-06-05T12:10:06.000000055Z\""), nil))
	assert.Exactly(t, chr, m)
}

func TestZeroYear(t *testing.T) {
	zy := ZeroYear()
	assert.Exactly(t, 0, zy.Year())
	assert.Exactly(t, time.January, zy.Month())
	assert.Exactly(t, 1, zy.Day())
	assert.Exactly(t, 0, zy.Hour())
	assert.Exactly(t, 0, zy.Minute())
	assert.Exactly(t, 0, zy.Second())
	assert.Exactly(t, 0, zy.Nanosecond())
}

func TestZeroUnix(t *testing.T) {
	assert.Exactly(t, TimeOf(time.Unix(0, 0)), ZeroUnix())
}
