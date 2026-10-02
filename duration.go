package chron

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Duration is a sum of calendar and clock offsets.
// Calendar fields are stored collapsed (months, days) and applied via AddDate;
// clock is applied last via Add (fixed nanoseconds).
type Duration struct {
	months int32 // years*12 + months
	days   int32 // weeks*7 + days
	clock  time.Duration
}

func Years(n int) Duration           { return Duration{months: int32(n) * 12} }
func Months(n int) Duration          { return Duration{months: int32(n)} }
func Weeks(n int) Duration           { return Duration{days: int32(n) * 7} }
func Days(n int) Duration            { return Duration{days: int32(n)} }
func Hours(n int) Duration           { return Duration{clock: time.Duration(n) * time.Hour} }
func Minutes(n int) Duration         { return Duration{clock: time.Duration(n) * time.Minute} }
func Seconds(n int) Duration         { return Duration{clock: time.Duration(n) * time.Second} }
func Millis(n int) Duration          { return Duration{clock: time.Duration(n) * time.Millisecond} }
func Micros(n int) Duration          { return Duration{clock: time.Duration(n) * time.Microsecond} }
func Nanos(n int) Duration           { return Duration{clock: time.Duration(n) * time.Nanosecond} }
func Clock(d time.Duration) Duration { return Duration{clock: d} }

func (d Duration) Years(n int) Duration           { d.months += int32(n) * 12; return d }
func (d Duration) Months(n int) Duration          { d.months += int32(n); return d }
func (d Duration) Weeks(n int) Duration           { d.days += int32(n) * 7; return d }
func (d Duration) Days(n int) Duration            { d.days += int32(n); return d }
func (d Duration) Hours(n int) Duration           { d.clock += time.Duration(n) * time.Hour; return d }
func (d Duration) Minutes(n int) Duration         { d.clock += time.Duration(n) * time.Minute; return d }
func (d Duration) Seconds(n int) Duration         { d.clock += time.Duration(n) * time.Second; return d }
func (d Duration) Millis(n int) Duration          { d.clock += time.Duration(n) * time.Millisecond; return d }
func (d Duration) Micros(n int) Duration          { d.clock += time.Duration(n) * time.Microsecond; return d }
func (d Duration) Nanos(n int) Duration           { d.clock += time.Duration(n) * time.Nanosecond; return d }
func (d Duration) Clock(c time.Duration) Duration { d.clock += c; return d }

// Neg flips the sign of all components.
func (d Duration) Neg() Duration {
	return Duration{
		months: -d.months,
		days:   -d.days,
		clock:  -d.clock,
	}
}

// Mul scales all components by n.
func (d Duration) Mul(n int) Duration {
	return Duration{
		months: d.months * int32(n),
		days:   d.days * int32(n),
		clock:  d.clock * time.Duration(n),
	}
}

// IsZero reports whether d has no offset.
func (d Duration) IsZero() bool {
	return d.months == 0 && d.days == 0 && d.clock == 0
}

func (d Duration) clockOnly() bool {
	return d.months == 0 && d.days == 0
}

// MustDuration parses s and panics on error. Prefer ParseDuration at I/O boundaries.
// Spec name Duration(s) is spelled MustDuration here because Go disallows a func and
// type with the same identifier in one package.
func MustDuration(s string) Duration {
	d, err := ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

// ParseDuration parses an ISO 8601 duration string with chron extensions.
// Blank and whitespace-only input error. Explicit zeros such as "P0D", "PT0S",
// and "0d" succeed as a zero Duration (same idea as time.ParseDuration("0s")).
func ParseDuration(s string) (Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Duration{}, errParseDuration
	}

	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = strings.TrimPrefix(s, "-")
	}

	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	if !strings.HasPrefix(lower, "p") {
		if isBareClockUnit(lower) {
			s = "pt" + s
		} else {
			s = "p" + s
		}
		lower = strings.ToLower(s)
	}

	if len(lower) < 2 {
		return Duration{}, errParseDuration
	}

	datePart := ""
	timePart := ""
	body := s[1:]
	lowerBody := lower[1:]

	if tIdx := strings.IndexByte(lowerBody, 't'); tIdx >= 0 {
		datePart = body[:tIdx]
		timePart = body[tIdx+1:]
	} else {
		datePart = body
	}

	// Require at least one component so bare "P"/"PT" stay invalid, while
	// explicit zeros ("P0D", "PT0S") parse successfully.
	if datePart == "" && timePart == "" {
		return Duration{}, errParseDuration
	}

	d := Duration{}
	if err := parseDatePart(datePart, &d); err != nil {
		return Duration{}, err
	}
	if err := parseTimePart(timePart, &d); err != nil {
		return Duration{}, err
	}

	if neg {
		d = d.Neg()
	}
	return d, nil
}

// Precision returns the finest unit after canonical decomposition.
func (d Duration) Precision() Precision {
	if d.clock != 0 {
		abs := d.clock
		if abs < 0 {
			abs = -abs
		}
		switch {
		case abs%time.Hour == 0:
			return Hour
		case abs%time.Minute == 0:
			return Minute
		case abs%time.Second == 0:
			return Second
		case abs%time.Millisecond == 0:
			return Millisecond
		case abs%time.Microsecond == 0:
			return Microsecond
		default:
			return Nanosecond
		}
	}

	months := d.months
	days := d.days
	if months < 0 {
		months = -months
	}
	if days < 0 {
		days = -days
	}

	y, m := months/12, months%12
	w, dayRem := days/7, days%7

	switch {
	case dayRem != 0:
		return Day
	case w != 0:
		return Week
	case m != 0:
		return Month
	case y != 0:
		return Year
	default:
		return Nanosecond
	}
}

func isBareClockUnit(s string) bool {
	units := []string{"h", "m", "s", "ms", "ns", "us", "µs", "μs"}
	for _, u := range units {
		if strings.HasSuffix(s, u) {
			return true
		}
	}
	return false
}
func parseDatePart(part string, d *Duration) error {
	for len(part) > 0 {
		numStr, rest, ok := readNumber(part)
		if !ok || rest == "" {
			return errParseDuration
		}
		unit := unicode.ToLower(rune(rest[0]))
		part = rest[1:]

		if unit == 'y' {
			if strings.ContainsAny(numStr, ".,") {
				f, err := strconv.ParseFloat(strings.ReplaceAll(numStr, ",", "."), 64)
				if err != nil {
					return errParseDuration
				}
				// Fractional years become whole months (P1.5Y → 18).
				d.months += int32(f * 12)
				continue
			}
			n, err := strconv.Atoi(numStr)
			if err != nil {
				return errParseDuration
			}
			d.months += int32(n) * 12
			continue
		}

		n, err := strconv.Atoi(strings.ReplaceAll(numStr, ",", "."))
		if err != nil {
			return err
		}
		switch unit {
		case 'm':
			d.months += int32(n)
		case 'w':
			d.days += int32(n) * 7
		case 'd':
			d.days += int32(n)
		default:
			return errParseDuration
		}
	}
	return nil
}

func parseTimePart(part string, d *Duration) error {
	for len(part) > 0 {
		numStr, rest, ok := readNumber(part)
		if !ok || rest == "" {
			return errParseDuration
		}
		numStr = strings.ReplaceAll(numStr, ",", ".")
		lower := strings.ToLower(rest)

		// Longer prefixes first so "ms" is not read as minutes.
		// µ (U+00B5) and μ (U+03BC) are both 3 bytes and do not case-fold into each other.
		var (
			scale    time.Duration
			consumed int
			frac     bool
		)
		switch {
		case strings.HasPrefix(lower, "ms"):
			scale, consumed = time.Millisecond, len("ms")
		case strings.HasPrefix(lower, "ns"):
			scale, consumed = time.Nanosecond, len("ns")
		case strings.HasPrefix(lower, "us"):
			scale, consumed = time.Microsecond, len("us")
		case strings.HasPrefix(lower, "µs"), strings.HasPrefix(lower, "μs"):
			scale, consumed = time.Microsecond, len("µs")
		case strings.HasPrefix(lower, "h"):
			scale, consumed = time.Hour, len("h")
		case strings.HasPrefix(lower, "m"):
			scale, consumed = time.Minute, len("m")
		case strings.HasPrefix(lower, "s"):
			scale, consumed, frac = time.Second, len("s"), true
		default:
			return errParseDuration
		}

		if frac {
			f, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return errParseDuration
			}
			d.clock += time.Duration(f * float64(scale))
		} else {
			n, err := strconv.Atoi(numStr)
			if err != nil {
				return err
			}
			d.clock += time.Duration(n) * scale
		}
		part = rest[consumed:]
	}
	return nil
}

func readNumber(s string) (num, rest string, ok bool) {
	i := 0
	for i < len(s) && (unicode.IsDigit(rune(s[i])) || s[i] == '.' || s[i] == ',') {
		i++
	}
	if i == 0 {
		return "", "", false
	}
	return s[:i], s[i:], true
}

// String returns the canonical lowercase duration string.
// Zero returns "". Months decompose to y+m; days to w+d.
func (d Duration) String() string {
	if d.IsZero() {
		return ""
	}

	neg := d.months < 0 || d.days < 0 || d.clock < 0
	if neg {
		d = d.Neg()
	}

	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	b.WriteByte('p')

	y := d.months / 12
	m := d.months % 12
	w := d.days / 7
	day := d.days % 7

	if y != 0 {
		fmt.Fprintf(&b, "%dy", y)
	}
	if m != 0 {
		fmt.Fprintf(&b, "%dm", m)
	}
	if w != 0 {
		fmt.Fprintf(&b, "%dw", w)
	}
	if day != 0 {
		fmt.Fprintf(&b, "%dd", day)
	}

	if d.clock != 0 {
		b.WriteByte('t')
		b.WriteString(formatClock(d.clock))
	}

	return b.String()
}

func formatClock(clock time.Duration) string {
	if clock == 0 {
		return "0s"
	}

	h := clock / time.Hour
	clock %= time.Hour
	m := clock / time.Minute
	clock %= time.Minute
	s := clock / time.Second
	clock %= time.Second
	ms := clock / time.Millisecond
	clock %= time.Millisecond
	us := clock / time.Microsecond
	ns := clock % time.Microsecond

	var parts []string
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}
	if s > 0 {
		parts = append(parts, fmt.Sprintf("%ds", s))
	}
	if ms > 0 {
		parts = append(parts, fmt.Sprintf("%dms", ms))
	}
	if us > 0 {
		parts = append(parts, fmt.Sprintf("%dµs", us))
	}
	if ns > 0 {
		parts = append(parts, fmt.Sprintf("%dns", ns))
	}
	if len(parts) == 0 {
		return "0s"
	}
	return strings.Join(parts, "")
}

// MarshalJSON encodes d as a JSON string, or null when zero.
func (d Duration) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.String())
}

// UnmarshalJSON decodes a JSON string or null into d.
func (d *Duration) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*d = Duration{}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*d = Duration{}
		return nil
	}
	parsed, err := ParseDuration(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
