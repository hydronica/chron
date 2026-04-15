package chron

import (
	"time"

	"tools/dura"
)

// --- Chron-style (nanosecond instant) additions — same as legacy Chron.Add* ---

func AddYearsChron(t time.Time, y int) time.Time {
	return Increment(t, dura.Years(y))
}

func AddMonthsChron(t time.Time, m int) time.Time {
	return Increment(t, dura.Months(m))
}

func AddDaysChron(t time.Time, d int) time.Time {
	return Increment(t, dura.Days(d))
}

func AddHoursChron(t time.Time, h int) time.Time {
	return Increment(t, dura.Hours(h))
}

func AddMinutesChron(t time.Time, m int) time.Time {
	return Increment(t, dura.Mins(m))
}

func AddSecondsChron(t time.Time, s int) time.Time {
	return Increment(t, dura.Secs(s))
}

func AddMillisChron(t time.Time, m int) time.Time {
	return Increment(t, dura.Millis(m))
}

func AddMicrosChron(t time.Time, m int) time.Time {
	return Increment(t, dura.Micros(m))
}

func AddNanosChron(t time.Time, n int) time.Time {
	return AddN(t, n)
}

// --- Year bucket ---

func AddYearsYear(t time.Time, n int) time.Time {
	return YearOf(t).AddDate(n, 0, 0)
}

func AddMonthsYear(t time.Time, m int) time.Time {
	return MonthOf(YearOf(t)).AddDate(0, m, 0)
}

func AddDaysYear(t time.Time, d int) time.Time {
	return DayOf(YearOf(t)).AddDate(0, 0, d)
}

func AddHoursYear(t time.Time, h int) time.Time {
	return HourOf(YearOf(t)).Add(time.Duration(int(time.Hour) * h))
}

func AddMinutesYear(t time.Time, m int) time.Time {
	return MinuteOf(YearOf(t)).Add(time.Duration(int(time.Minute) * m))
}

func AddSecondsYear(t time.Time, s int) time.Time {
	return SecondOf(YearOf(t)).Add(time.Duration(int(time.Second) * s))
}

func AddMillisYear(t time.Time, m int) time.Time {
	return MilliOf(YearOf(t)).Add(time.Duration(int(time.Millisecond) * m))
}

func AddMicrosYear(t time.Time, m int) time.Time {
	return MicroOf(YearOf(t)).Add(time.Duration(int(time.Microsecond) * m))
}

func AddNanosYear(t time.Time, n int) time.Time {
	return AddN(YearOf(t), n)
}

// --- Month bucket ---

func AddYearsMonth(t time.Time, y int) time.Time {
	return MonthOf(Increment(MonthOf(t), dura.Years(y)))
}

func AddMonthsMonth(t time.Time, m int) time.Time {
	return MonthOf(t).AddDate(0, m, 0)
}

func AddDaysMonth(t time.Time, d int) time.Time {
	return DayOf(MonthOf(t)).AddDate(0, 0, d)
}

func AddHoursMonth(t time.Time, h int) time.Time {
	return HourOf(MonthOf(t)).Add(time.Duration(int(time.Hour) * h))
}

func AddMinutesMonth(t time.Time, m int) time.Time {
	return MinuteOf(MonthOf(t)).Add(time.Duration(int(time.Minute) * m))
}

func AddSecondsMonth(t time.Time, s int) time.Time {
	return SecondOf(MonthOf(t)).Add(time.Duration(int(time.Second) * s))
}

func AddMillisMonth(t time.Time, m int) time.Time {
	return MilliOf(MonthOf(t)).Add(time.Duration(int(time.Millisecond) * m))
}

func AddMicrosMonth(t time.Time, m int) time.Time {
	return MicroOf(MonthOf(t)).Add(time.Duration(int(time.Microsecond) * m))
}

func AddNanosMonth(t time.Time, n int) time.Time {
	return AddN(MonthOf(t), n)
}

// --- Day bucket ---

func AddYearsDay(t time.Time, y int) time.Time {
	return DayOf(Increment(DayOf(t), dura.Years(y)))
}

func AddMonthsDay(t time.Time, m int) time.Time {
	return DayOf(Increment(DayOf(t), dura.Months(m)))
}

func AddDaysDay(t time.Time, d int) time.Time {
	return DayOf(t).AddDate(0, 0, d)
}

func AddHoursDay(t time.Time, h int) time.Time {
	return HourOf(DayOf(t)).Add(time.Duration(int(time.Hour) * h))
}

func AddMinutesDay(t time.Time, m int) time.Time {
	return MinuteOf(DayOf(t)).Add(time.Duration(int(time.Minute) * m))
}

func AddSecondsDay(t time.Time, s int) time.Time {
	return SecondOf(DayOf(t)).Add(time.Duration(int(time.Second) * s))
}

func AddMillisDay(t time.Time, m int) time.Time {
	return MilliOf(DayOf(t)).Add(time.Duration(int(time.Millisecond) * m))
}

func AddMicrosDay(t time.Time, m int) time.Time {
	return MicroOf(DayOf(t)).Add(time.Duration(int(time.Microsecond) * m))
}

func AddNanosDay(t time.Time, n int) time.Time {
	return AddN(DayOf(t), n)
}

// --- Hour bucket ---

func AddYearsHour(t time.Time, y int) time.Time {
	return HourOf(Increment(HourOf(t), dura.Years(y)))
}

func AddMonthsHour(t time.Time, m int) time.Time {
	return HourOf(Increment(HourOf(t), dura.Months(m)))
}

func AddDaysHour(t time.Time, d int) time.Time {
	return HourOf(Increment(HourOf(t), dura.Days(d)))
}

func AddHoursHour(t time.Time, h int) time.Time {
	return HourOf(t).Add(time.Duration(int(time.Hour) * h))
}

func AddMinutesHour(t time.Time, m int) time.Time {
	return MinuteOf(HourOf(t)).Add(time.Duration(int(time.Minute) * m))
}

func AddSecondsHour(t time.Time, s int) time.Time {
	return SecondOf(HourOf(t)).Add(time.Duration(int(time.Second) * s))
}

func AddMillisHour(t time.Time, m int) time.Time {
	return MilliOf(HourOf(t)).Add(time.Duration(int(time.Millisecond) * m))
}

func AddMicrosHour(t time.Time, m int) time.Time {
	return MicroOf(HourOf(t)).Add(time.Duration(int(time.Microsecond) * m))
}

func AddNanosHour(t time.Time, n int) time.Time {
	return AddN(HourOf(t), n)
}

// --- Minute bucket ---

func AddYearsMinute(t time.Time, y int) time.Time {
	return MinuteOf(Increment(MinuteOf(t), dura.Years(y)))
}

func AddMonthsMinute(t time.Time, m int) time.Time {
	return MinuteOf(Increment(MinuteOf(t), dura.Months(m)))
}

func AddDaysMinute(t time.Time, d int) time.Time {
	return MinuteOf(Increment(MinuteOf(t), dura.Days(d)))
}

func AddHoursMinute(t time.Time, h int) time.Time {
	return MinuteOf(Increment(MinuteOf(t), dura.Hours(h)))
}

func AddMinutesMinute(t time.Time, m int) time.Time {
	return MinuteOf(t).Add(time.Duration(int(time.Minute) * m))
}

func AddSecondsMinute(t time.Time, s int) time.Time {
	return SecondOf(MinuteOf(t)).Add(time.Duration(int(time.Second) * s))
}

func AddMillisMinute(t time.Time, m int) time.Time {
	return MilliOf(MinuteOf(t)).Add(time.Duration(int(time.Millisecond) * m))
}

func AddMicrosMinute(t time.Time, m int) time.Time {
	return MicroOf(MinuteOf(t)).Add(time.Duration(int(time.Microsecond) * m))
}

func AddNanosMinute(t time.Time, n int) time.Time {
	return AddN(MinuteOf(t), n)
}

// --- Second bucket ---

func AddYearsSecond(t time.Time, y int) time.Time {
	return SecondOf(Increment(SecondOf(t), dura.Years(y)))
}

func AddMonthsSecond(t time.Time, m int) time.Time {
	return SecondOf(Increment(SecondOf(t), dura.Months(m)))
}

func AddDaysSecond(t time.Time, d int) time.Time {
	return SecondOf(Increment(SecondOf(t), dura.Days(d)))
}

func AddHoursSecond(t time.Time, h int) time.Time {
	return SecondOf(Increment(SecondOf(t), dura.Hours(h)))
}

func AddMinutesSecond(t time.Time, m int) time.Time {
	return SecondOf(Increment(SecondOf(t), dura.Mins(m)))
}

func AddSecondsSecond(t time.Time, s int) time.Time {
	return SecondOf(t).Add(time.Duration(int(time.Second) * s))
}

func AddMillisSecond(t time.Time, m int) time.Time {
	return MilliOf(SecondOf(t)).Add(time.Duration(int(time.Millisecond) * m))
}

func AddMicrosSecond(t time.Time, m int) time.Time {
	return MicroOf(SecondOf(t)).Add(time.Duration(int(time.Microsecond) * m))
}

func AddNanosSecond(t time.Time, n int) time.Time {
	return AddN(SecondOf(t), n)
}

// --- Milli bucket ---

func AddYearsMilli(t time.Time, y int) time.Time {
	return MilliOf(Increment(MilliOf(t), dura.Years(y)))
}

func AddMonthsMilli(t time.Time, m int) time.Time {
	return MilliOf(Increment(MilliOf(t), dura.Months(m)))
}

func AddDaysMilli(t time.Time, d int) time.Time {
	return MilliOf(Increment(MilliOf(t), dura.Days(d)))
}

func AddHoursMilli(t time.Time, h int) time.Time {
	return MilliOf(Increment(MilliOf(t), dura.Hours(h)))
}

func AddMinutesMilli(t time.Time, m int) time.Time {
	return MilliOf(Increment(MilliOf(t), dura.Mins(m)))
}

func AddSecondsMilli(t time.Time, s int) time.Time {
	return MilliOf(Increment(MilliOf(t), dura.Secs(s)))
}

func AddMillisMilli(t time.Time, m int) time.Time {
	return MilliOf(t).Add(time.Duration(int(time.Millisecond) * m))
}

func AddMicrosMilli(t time.Time, m int) time.Time {
	return MicroOf(MilliOf(t)).Add(time.Duration(int(time.Microsecond) * m))
}

func AddNanosMilli(t time.Time, n int) time.Time {
	return AddN(MilliOf(t), n)
}

// --- Micro bucket ---

func AddYearsMicro(t time.Time, y int) time.Time {
	return MicroOf(Increment(MicroOf(t), dura.Years(y)))
}

func AddMonthsMicro(t time.Time, m int) time.Time {
	return MicroOf(Increment(MicroOf(t), dura.Months(m)))
}

func AddDaysMicro(t time.Time, d int) time.Time {
	return MicroOf(Increment(MicroOf(t), dura.Days(d)))
}

func AddHoursMicro(t time.Time, h int) time.Time {
	return MicroOf(Increment(MicroOf(t), dura.Hours(h)))
}

func AddMinutesMicro(t time.Time, m int) time.Time {
	return MicroOf(Increment(MicroOf(t), dura.Mins(m)))
}

func AddSecondsMicro(t time.Time, s int) time.Time {
	return MicroOf(Increment(MicroOf(t), dura.Secs(s)))
}

func AddMillisMicro(t time.Time, m int) time.Time {
	return MicroOf(Increment(MicroOf(t), dura.Millis(m)))
}

func AddMicrosMicro(t time.Time, m int) time.Time {
	return MicroOf(t).Add(time.Duration(int(time.Microsecond) * m))
}

func AddNanosMicro(t time.Time, n int) time.Time {
	return AddN(MicroOf(t), n)
}

// AddN* functions step one unit of the bucket (legacy AddN on Year, Month, …).

func AddNYear(t time.Time, n int) time.Time {
	return YearOf(t).AddDate(n, 0, 0)
}

func AddNMonth(t time.Time, n int) time.Time {
	return MonthOf(t).AddDate(0, n, 0)
}

func AddNDay(t time.Time, n int) time.Time {
	return DayOf(t).AddDate(0, 0, n)
}

func AddNHour(t time.Time, n int) time.Time {
	return HourOf(t).Add(time.Duration(int(time.Hour) * n))
}

func AddNMinute(t time.Time, n int) time.Time {
	return MinuteOf(t).Add(time.Duration(int(time.Minute) * n))
}

func AddNSecond(t time.Time, n int) time.Time {
	return SecondOf(t).Add(time.Duration(int(time.Second) * n))
}

func AddNMilli(t time.Time, n int) time.Time {
	return MilliOf(t).Add(time.Duration(int(time.Millisecond) * n))
}

func AddNMicro(t time.Time, n int) time.Time {
	return MicroOf(t).Add(time.Duration(int(time.Microsecond) * n))
}
