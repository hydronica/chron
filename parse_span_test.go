package chron

import (
	"testing"
	"time"

	"tools/dura"

	"github.com/stretchr/testify/assert"
)

func TestParseUnixSeconds(t *testing.T) {
	d := NewDay(2016, time.March, 17)
	ti, err := ParseUnixSeconds("1458172800")
	assert.True(t, err == nil)
	assert.Exactly(t, d, ti)
}

func TestParseWithFormats(t *testing.T) {
	tt, err := ParseWithFormats("03-Feb-18")
	day := NewDay(2018, time.February, 3)
	assert.Nil(t, err)
	assert.Exactly(t, day, tt)
	assert.Exactly(t, day, DayOf(tt))

	tt, err = ParseWithFormats("03-Feb-2018")
	assert.Nil(t, err)
	assert.Exactly(t, day, tt)
	assert.Exactly(t, day, DayOf(tt))

	tt, err = ParseWithFormats("02-03-18")
	assert.Nil(t, err)
	assert.Exactly(t, day, tt)
	assert.Exactly(t, day, DayOf(tt))

	tt, err = ParseWithFormats("02-03-2018")
	assert.Nil(t, err)
	assert.Exactly(t, day, tt)
	assert.Exactly(t, day, DayOf(tt))

	tt, err = ParseWithFormats("02/03/18")
	assert.Nil(t, err)
	assert.Exactly(t, day, tt)
	assert.Exactly(t, day, DayOf(tt))

	tt, err = ParseWithFormats("02/03/2018")
	assert.Nil(t, err)
	assert.Exactly(t, day, tt)
	assert.Exactly(t, day, DayOf(tt))

	tt, err = ParseWithFormats("02/03/2018 3:13 PM")
	min := NewMinute(2018, time.February, 3, 15, 13)
	assert.Nil(t, err)
	assert.Exactly(t, min, tt)
	assert.Exactly(t, min, MinuteOf(tt))

	tt, err = ParseWithFormats("02/03/2018 3:13:52 PM")
	sec := NewSecond(2018, time.February, 3, 15, 13, 52)
	assert.Nil(t, err)
	assert.Exactly(t, sec, tt)
	assert.Exactly(t, sec, SecondOf(tt))

	tt, err = ParseWithFormats("02/03/2018 15:13")
	assert.Nil(t, err)
	assert.Exactly(t, min, tt)
	assert.Exactly(t, min, MinuteOf(tt))

	tt, err = ParseWithFormats("02/03/2018 15:13:52")
	assert.Nil(t, err)
	assert.Exactly(t, sec, tt)
	assert.Exactly(t, sec, SecondOf(tt))

	month := NewMonth(2018, time.February)
	tt, err = ParseWithFormats("Feb-2018")
	assert.Nil(t, err)
	assert.Exactly(t, month, tt)
	assert.Exactly(t, month, MonthOf(tt))

	tt, err = ParseWithFormats("Feb-18")
	assert.Nil(t, err)
	assert.Exactly(t, month, tt)
	assert.Exactly(t, month, MonthOf(tt))

	tt, err = ParseWithFormats("02-2018")
	assert.Nil(t, err)
	assert.Exactly(t, month, tt)
	assert.Exactly(t, month, MonthOf(tt))

	tt, err = ParseWithFormats("02-18")
	assert.Nil(t, err)
	assert.Exactly(t, month, tt)
	assert.Exactly(t, month, MonthOf(tt))

	tt, err = ParseWithFormats("02/18")
	assert.Nil(t, err)
	assert.Exactly(t, month, tt)
	assert.Exactly(t, month, MonthOf(tt))

	year := NewYear(2018)
	tt, err = ParseWithFormats("2018")
	assert.Nil(t, err)
	assert.Exactly(t, year, tt)
	assert.Exactly(t, year, YearOf(tt))
}

func TestNewInterval(t *testing.T) {
	tchron := NewTime(2018, time.February, 27, 4, 0, 0, 0)
	tinterval := NewInterval(tchron, dura.Week)
	zeroInterval := NewInterval(tchron, dura.Nano)

	assert.Exactly(t, NewInterval(tchron, dura.Week), tinterval)
	assert.Exactly(t, NewInterval(tchron, dura.Nano), zeroInterval)
}

func TestIntervalContains(t *testing.T) {
	tchron := NewTime(2018, time.February, 27, 4, 0, 0, 0)
	tinterval := NewInterval(tchron, dura.Week)
	zeroInterval := NewInterval(tchron, dura.Nano)

	assert.True(t, tinterval.Contains(SpanOfHour(NewHour(2018, time.March, 2, 23))))
	assert.True(t, zeroInterval.Contains(InstantSpan(tchron)))
}

func TestIntervalDuration(t *testing.T) {
	tchron := NewTime(2018, time.February, 27, 4, 0, 0, 0)
	tinterval := NewInterval(tchron, dura.Week)
	assert.Exactly(t, dura.Week, tinterval.Duration())
}

func TestIntervalString(t *testing.T) {
	tchron := NewTime(2018, time.February, 27, 4, 0, 0, 0)
	tinterval := NewInterval(tchron, dura.Week)
	assert.Exactly(t, "start:2018-02-27 04:00:00 +0000 UTC, end:2018-03-06 03:59:59.999999999 +0000 UTC, len:Week", tinterval.String())
}
