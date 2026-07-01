package chron_test

import (
	"encoding/json"
	"testing"
	"time"

	"chron"
)

func durationEqual(a, b chron.Duration) bool {
	anchor := chron.Date(2020, 6, 15)
	return anchor.Add(a).Equal(anchor.Add(b).AsTime())
}

func TestDateAndTruncate(t *testing.T) {
	jan := chron.Date(2026, 1, 15).Truncate(chron.Month)
	if jan.Precision() != chron.Month {
		t.Fatalf("precision = %v, want Month", jan.Precision())
	}
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !jan.Equal(want) {
		t.Fatalf("got %v, want %v", jan.AsTime(), want)
	}

	sixth := jan.Add(chron.Days(5))
	if sixth.Precision() != chron.Day {
		t.Fatalf("precision = %v, want Day", sixth.Precision())
	}
	want6 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	if !sixth.Equal(want6) {
		t.Fatalf("got %v, want %v", sixth.AsTime(), want6)
	}
}

func TestAddPrecisionRules(t *testing.T) {
	event := chron.FromTime(time.Date(2026, 2, 4, 15, 0, 0, 123, time.UTC))
	if event.Add(chron.Hours(2)).Precision() != chron.Nanosecond {
		t.Fatal("clock-only add on nanosecond instant should stay Nanosecond")
	}
	if event.Add(chron.Months(1)).Precision() != chron.Month {
		t.Fatal("calendar add on nanosecond instant should become Month")
	}

	day := chron.Date(2026, 6, 1)
	if day.Add(chron.Hours(3)).Precision() != chron.Hour {
		t.Fatal("day + hours should become Hour precision")
	}
}

func TestCalendarMonthAdd(t *testing.T) {
	got := chron.Date(2026, 1, 31).Add(chron.Months(1))
	// Go AddDate normalizes Feb 31 → Mar 3 (non-leap year).
	want := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("Jan 31 + 1 month = %v, want %v", got.AsTime(), want)
	}
}

func TestParseMonth(t *testing.T) {
	c, err := chron.Parse("2026-02")
	if err != nil {
		t.Fatal(err)
	}
	if c.Precision() != chron.Month {
		t.Fatalf("precision = %v, want Month", c.Precision())
	}
	want := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !c.Equal(want) {
		t.Fatalf("got %v, want %v", c.AsTime(), want)
	}

	data, err := c.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"2026-02"` {
		t.Fatalf("marshal = %s, want %q", data, `"2026-02"`)
	}
}

func TestSpanMonth(t *testing.T) {
	feb := chron.Date(2026, 2, 15).Span(chron.Month)
	event := chron.FromTime(time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC))
	if !feb.Contains(event) {
		t.Fatal("event should be inside February span")
	}
	before := chron.FromTime(time.Date(2026, 1, 31, 23, 0, 0, 0, time.UTC))
	if feb.Contains(before) {
		t.Fatal("Jan 31 should not be inside February span")
	}
	if !feb.End.Equal(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("end = %v", feb.End.AsTime())
	}
}

func TestWeekBoundaries(t *testing.T) {
	anchor := chron.Date(2026, 2, 4)
	iso := anchor.Span(chron.MondayWeek)
	wantStart := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	if !iso.Start.Equal(wantStart) {
		t.Fatalf("iso start = %v, want %v", iso.Start.AsTime(), wantStart)
	}
	if iso.Start.Precision() != chron.MondayWeek {
		t.Fatalf("stored precision = %v, want MondayWeek", iso.Start.Precision())
	}

	us := anchor.Span(chron.SundayWeek)
	wantUS := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !us.Start.Equal(wantUS) {
		t.Fatalf("us start = %v, want %v", us.Start.AsTime(), wantUS)
	}
	if us.Start.Precision() != chron.SundayWeek {
		t.Fatalf("stored precision = %v, want SundayWeek", us.Start.Precision())
	}
}

func TestSpanOverlap(t *testing.T) {
	a := chron.MustSpan(
		chron.Date(2026, 1, 1),
		chron.Date(2026, 2, 1),
	)
	b := chron.MustSpan(
		chron.Date(2026, 1, 15),
		chron.Date(2026, 3, 1),
	)
	if !a.Overlaps(b) {
		t.Fatal("spans should overlap")
	}
	adj := chron.MustSpan(chron.Date(2026, 2, 1), chron.Date(2026, 3, 1))
	if !a.Adjacent(adj) {
		t.Fatal("spans should be adjacent")
	}
}

func TestParseDuration(t *testing.T) {
	cases := []struct {
		in   string
		want chron.Duration
	}{
		{"P14D", chron.Days(14)},
		{"14d", chron.Days(14)},
		{"PT12H", chron.Hours(12)},
		{"12h", chron.Hours(12)},
		{"P2W", chron.Weeks(2)},
		{"P1Y3M4DT12H", chron.Years(1).Months(3).Days(4).Hours(12)},
		{"PT0.001S", chron.Millis(1)},
		{"pt4h500ms", chron.Hours(4).Millis(500)},
		{"250us", chron.Micros(250)},
		{"250µs", chron.Micros(250)},
		{"-P1D", chron.Days(-1)},
	}
	for _, tc := range cases {
		got, err := chron.ParseDuration(tc.in)
		if err != nil {
			t.Fatalf("ParseDuration(%q): %v", tc.in, err)
		}
		anchor := chron.Date(2026, 1, 15)
		if !anchor.Add(got).Equal(anchor.Add(tc.want).AsTime()) {
			t.Fatalf("ParseDuration(%q) apply mismatch", tc.in)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	if got := chron.FormatDuration(chron.Duration{}); got != "p0d" {
		t.Fatalf("zero = %q, want p0d", got)
	}
	if got := chron.FormatDuration(chron.Days(14)); got != "p14d" {
		t.Fatalf("14d = %q", got)
	}
	if got := chron.FormatDuration(chron.Micros(250)); got != "pt250µs" {
		t.Fatalf("250us = %q", got)
	}
}

func TestDurationJSON(t *testing.T) {
	type cfg struct {
		Length chron.Duration `json:"length"`
	}
	in := cfg{Length: chron.Days(14)}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"length":"p14d"}` {
		t.Fatalf("marshal = %s", data)
	}
	var out cfg
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if !durationEqual(out.Length, in.Length) {
		t.Fatal("round-trip mismatch")
	}
}

func TestChronJSONNull(t *testing.T) {
	var c chron.Chron
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Fatalf("zero marshal = %s, want null", data)
	}
}

func TestWeeksDurationUsesAddDate(t *testing.T) {
	start := chron.Date(2026, 1, 31)
	end := start.Add(chron.Weeks(1))
	want := start.Add(chron.Days(7))
	if !end.Equal(want.AsTime()) {
		t.Fatalf("Weeks(1) = %v, want %v", end.AsTime(), want.AsTime())
	}
}

func TestInvalidSpan(t *testing.T) {
	_, err := chron.NewSpan(chron.Date(2026, 2, 1), chron.Date(2026, 1, 1))
	if err == nil {
		t.Fatal("expected error for invalid span")
	}
}

func TestInvalidDuration(t *testing.T) {
	_, err := chron.ParseDuration("P1.5Y")
	if err == nil {
		t.Fatal("expected error for fractional year")
	}
	_, err = chron.ParseDuration("P1Y1W")
	if err == nil {
		t.Fatal("expected error for W combined with Y")
	}
}

func TestParseFromPrecision(t *testing.T) {
	c, err := chron.ParseFrom("2006-01-02T15", "2026-02-15T14")
	if err != nil {
		t.Fatal(err)
	}
	if c.Precision() != chron.Hour {
		t.Fatalf("precision = %v, want Hour", c.Precision())
	}
}

func TestSub(t *testing.T) {
	c := chron.Date(2026, 3, 1).Sub(chron.Months(1))
	want := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !c.Equal(want) {
		t.Fatalf("got %v, want %v", c.AsTime(), want)
	}
}

func TestDurationNegMul(t *testing.T) {
	d := chron.Months(2).Mul(3)
	if d.String() != "p6m" {
		t.Fatalf("Mul = %s", d.String())
	}
	if !durationEqual(d.Neg(), chron.Months(-6)) {
		t.Fatal("Neg mismatch")
	}
}
