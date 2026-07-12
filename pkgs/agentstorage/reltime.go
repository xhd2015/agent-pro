package agentstorage

import (
	"fmt"
	"strings"
	"time"
)

// FormatRelativeAge formats the age of t relative to now for human session lists.
// Zero t → "-"; age < 1s (after clamping future to 0) → "just now";
// otherwise short units (s/m/h/d), max 2 non-zero units, zero stops chain, " ago".
func FormatRelativeAge(now, t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := now.Sub(t)
	if d < 0 {
		d = 0
	}
	if d < time.Second {
		return "just now"
	}

	totalSecs := int64(d / time.Second)
	days := totalSecs / 86400
	rem := totalSecs % 86400
	hours := rem / 3600
	rem = rem % 3600
	mins := rem / 60
	secs := rem % 60

	type unit struct {
		val   int64
		label string
	}
	units := []unit{
		{days, "d"},
		{hours, "h"},
		{mins, "m"},
		{secs, "s"},
	}

	var parts []string
	started := false
	for _, u := range units {
		if !started {
			if u.val == 0 {
				continue
			}
			started = true
		} else if u.val == 0 {
			// Zero unit stops the chain: omit it and everything smaller.
			break
		}
		parts = append(parts, fmt.Sprintf("%d%s", u.val, u.label))
		if len(parts) == 2 {
			break
		}
	}
	if len(parts) == 0 {
		return "just now"
	}
	return strings.Join(parts, "") + " ago"
}
