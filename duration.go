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
// Calendar fields are applied via AddDate (order: years, months, days).
// Clock is applied last via Add (fixed nanoseconds).
type Duration struct {
	years, months, weeks, days int
	clock                      time.Duration
}

func Years(n int) Duration           { return Duration{years: n} }
func Months(n int) Duration          { return Duration{months: n} }
func Weeks(n int) Duration           { return Duration{weeks: n} }
func Days(n int) Duration            { return Duration{days: n} }
func Hours(n int) Duration           { return Duration{clock: time.Duration(n) * time.Hour} }
func Minutes(n int) Duration         { return Duration{clock: time.Duration(n) * time.Minute} }
func Seconds(n int) Duration         { return Duration{clock: time.Duration(n) * time.Second} }
func Millis(n int) Duration          { return Duration{clock: time.Duration(n) * time.Millisecond} }
func Micros(n int) Duration          { return Duration{clock: time.Duration(n) * time.Microsecond} }
func Nanos(n int) Duration           { return Duration{clock: time.Duration(n) * time.Nanosecond} }
func Clock(d time.Duration) Duration { return Duration{clock: d} }

func (d Duration) Years(n int) Duration           { d.years += n; return d }
func (d Duration) Months(n int) Duration          { d.months += n; return d }
func (d Duration) Weeks(n int) Duration           { d.weeks += n; return d }
func (d Duration) Days(n int) Duration            { d.days += n; return d }
func (d Duration) Hours(n int) Duration           { d.clock += time.Duration(n) * time.Hour; return d }
func (d Duration) Min(n int) Duration             { d.clock += time.Duration(n) * time.Minute; return d }
func (d Duration) Sec(n int) Duration             { d.clock += time.Duration(n) * time.Second; return d }
func (d Duration) Millis(n int) Duration          { d.clock += time.Duration(n) * time.Millisecond; return d }
func (d Duration) Micros(n int) Duration          { d.clock += time.Duration(n) * time.Microsecond; return d }
func (d Duration) Nanos(n int) Duration           { d.clock += time.Duration(n) * time.Nanosecond; return d }
func (d Duration) Clock(c time.Duration) Duration { d.clock += c; return d }

// Neg flips the sign of all components.
func (d Duration) Neg() Duration {
	return Duration{
		years:  -d.years,
		months: -d.months,
		weeks:  -d.weeks,
		days:   -d.days,
		clock:  -d.clock,
	}
}

// Mul scales all components by n.
func (d Duration) Mul(n int) Duration {
	return Duration{
		years:  d.years * n,
		months: d.months * n,
		weeks:  d.weeks * n,
		days:   d.days * n,
		clock:  d.clock * time.Duration(n),
	}
}

// IsZero reports whether d has no offset.
func (d Duration) IsZero() bool {
	return d.years == 0 && d.months == 0 && d.weeks == 0 && d.days == 0 && d.clock == 0
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

	d := Duration{}
	if err := parseDatePart(datePart, &d); err != nil {
		return Duration{}, err
	}
	if err := parseTimePart(timePart, &d); err != nil {
		return Duration{}, err
	}

	if d.IsZero() {
		return Duration{}, errParseDuration
	}
	if neg {
		d = d.Neg()
	}
	return d, nil
}

func (d Duration) isClockOnly() bool {
	return d.years == 0 && d.months == 0 && d.weeks == 0 && d.days == 0
}

// Precision
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
			return MicroSecond
		default:
			return Nanosecond
		}
	}
	if d.days != 0 {
		return Day
	}
	if d.weeks != 0 {
		return Week
	}
	if d.months != 0 {
		return Month
	}
	if d.years != 0 {
		return Year
	}

	// duration is zero
	return Nanosecond
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
	if part == "" {
		return nil
	}
	hasW := false
	hasYMD := false
	for len(part) > 0 {
		numStr, rest, ok := readNumber(part)
		if !ok || len(rest) == 0 {
			return errParseDuration
		}
		if strings.ContainsAny(numStr, ".,") {
			return errInvalidDuration
		}
		n, err := strconv.Atoi(numStr)
		if err != nil {
			return errParseDuration
		}
		unit := unicode.ToLower([]rune(rest)[0])
		part = rest[1:]
		switch unit {
		case 'y':
			hasYMD = true
			d.years = n
		case 'm':
			hasYMD = true
			d.months = n
		case 'w':
			hasW = true
			d.weeks = n
		case 'd':
			hasYMD = true
			d.days = n
		default:
			return errParseDuration
		}
	}
	if hasW && hasYMD {
		return errInvalidDuration
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
		lowerRest := strings.ToLower(rest)

		var consumed int
		switch {
		case strings.HasPrefix(lowerRest, "ms"):
			n, err := parseIntUnit(numStr)
			if err != nil {
				return err
			}
			d.clock += time.Duration(n) * time.Millisecond
			consumed = 2
		case strings.HasPrefix(lowerRest, "us"), strings.HasPrefix(lowerRest, "µs"), strings.HasPrefix(lowerRest, "μs"):
			n, err := parseIntUnit(numStr)
			if err != nil {
				return err
			}
			d.clock += time.Duration(n) * time.Microsecond
			consumed = unitLen(rest, "us", "µs", "μs")
		case strings.HasPrefix(lowerRest, "ns"):
			n, err := parseIntUnit(numStr)
			if err != nil {
				return err
			}
			d.clock += time.Duration(n) * time.Nanosecond
			consumed = 2
		case strings.HasPrefix(lowerRest, "h"):
			n, err := parseIntUnit(numStr)
			if err != nil {
				return err
			}
			d.clock += time.Duration(n) * time.Hour
			consumed = 1
		case strings.HasPrefix(lowerRest, "m"):
			n, err := parseIntUnit(numStr)
			if err != nil {
				return err
			}
			d.clock += time.Duration(n) * time.Minute
			consumed = 1
		case strings.HasPrefix(lowerRest, "s"):
			f, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return errParseDuration
			}
			d.clock += time.Duration(f * float64(time.Second))
			consumed = 1
		default:
			return errParseDuration
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

func parseIntUnit(numStr string) (int, error) {
	if strings.Contains(numStr, ".") {
		return 0, errInvalidDuration
	}
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, errParseDuration
	}
	return n, nil
}

func unitLen(rest string, ascii string, runes ...string) int {
	lower := strings.ToLower(rest)
	if strings.HasPrefix(lower, ascii) {
		return len(ascii)
	}
	for _, r := range runes {
		if strings.HasPrefix(rest, r) {
			return len(r)
		}
	}
	return 2
}

// FormatDuration returns the canonical lowercase duration string.
func FormatDuration(d Duration) string {
	if d.IsZero() {
		return "p0d"
	}

	neg := d.years < 0 || d.months < 0 || d.weeks < 0 || d.days < 0 || d.clock < 0
	if neg {
		d = d.Neg()
	}

	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	b.WriteByte('p')

	if d.years != 0 {
		fmt.Fprintf(&b, "%dy", d.years)
	}
	if d.months != 0 {
		fmt.Fprintf(&b, "%dm", d.months)
	}
	if d.weeks != 0 {
		fmt.Fprintf(&b, "%dw", d.weeks)
	}
	if d.days != 0 {
		fmt.Fprintf(&b, "%dd", d.days)
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

// String returns the canonical duration string.
func (d Duration) String() string {
	return FormatDuration(d)
}

// MarshalJSON encodes d as a JSON string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(FormatDuration(d))
}

// UnmarshalJSON decodes a JSON string into d.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseDuration(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
