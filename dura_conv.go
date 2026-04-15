package chron

import "tools/dura"

// AsDuration normalizes any dura.Time to a concrete Duration.
func AsDuration(d dura.Time) dura.Duration {
	return dura.NewDuration(d.Years(), d.Months(), d.Days(), d.Duration())
}
