// Package cronspec parses schedule specs for `agent-pro usage collect --cron`.
//
// Accepted forms:
//
//	5-field cron   "*/5 * * * *", "0 9 * * 1-5"
//	macros         @hourly @daily @midnight @weekly @monthly
//	interval       "@every 10m", or a bare duration "10m" / "300" (seconds)
//
// Cron fields are minute, hour, day-of-month, month, day-of-week. Each accepts
// "*", "a", "a-b", "*/n", "a-b/n" and comma-separated lists of those.
// Day-of-week is 0-7 with both 0 and 7 meaning Sunday. When day-of-month and
// day-of-week are both restricted, a time matches when either one matches
// (Vixie cron semantics).
package cronspec

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schedule is a parsed spec that can compute its next fire time.
type Schedule interface {
	// Next returns the first fire time strictly after after, in after's
	// location. The zero time means the spec can never fire again.
	Next(after time.Time) time.Time
	// String returns the canonical spec text.
	String() string
}

// Field bounds.
const (
	minMinute, maxMinute   = 0, 59
	minHour, maxHour       = 0, 23
	minDay, maxDay         = 1, 31
	minMonth, maxMonth     = 1, 12
	minWeekday, maxWeekday = 0, 7
)

// acceptedForms describes the dialect for error messages.
const acceptedForms = "expected 5 fields (minute hour day-of-month month day-of-week), " +
	"@hourly/@daily/@midnight/@weekly/@monthly, or @every <duration>"

// Parse parses spec into a Schedule.
func Parse(spec string) (Schedule, error) {
	raw := strings.TrimSpace(spec)
	if raw == "" {
		return nil, fmt.Errorf("%s", acceptedForms)
	}

	if strings.HasPrefix(raw, "@") {
		return parseMacro(raw)
	}

	fields := strings.Fields(raw)
	switch len(fields) {
	case 1:
		interval, err := parseInterval(fields[0])
		if err != nil {
			return nil, fmt.Errorf("%s (single field %q is not a duration)", acceptedForms, fields[0])
		}
		return &intervalSchedule{spec: raw, every: interval}, nil
	case 5:
		return parseCronFields(raw, fields)
	default:
		return nil, fmt.Errorf("%s; got %d fields", acceptedForms, len(fields))
	}
}

func parseMacro(raw string) (Schedule, error) {
	lower := strings.ToLower(raw)
	if rest, ok := cutPrefixFold(lower, "@every"); ok {
		durText := strings.TrimSpace(rest)
		if durText == "" {
			return nil, fmt.Errorf("@every requires a duration, e.g. @every 10m")
		}
		interval, err := parseInterval(durText)
		if err != nil {
			return nil, err
		}
		return &intervalSchedule{spec: raw, every: interval}, nil
	}

	var spec string
	switch lower {
	case "@hourly":
		spec = "0 * * * *"
	case "@daily", "@midnight":
		spec = "0 0 * * *"
	case "@weekly":
		spec = "0 0 * * 0"
	case "@monthly":
		spec = "0 0 1 * *"
	default:
		return nil, fmt.Errorf("unknown macro %q; %s", raw, acceptedForms)
	}
	return parseCronFields(raw, strings.Fields(spec))
}

// cutPrefixFold trims prefix from s case-insensitively, reporting whether it matched.
func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || !strings.EqualFold(s[:len(prefix)], prefix) {
		return s, false
	}
	return s[len(prefix):], true
}

// parseInterval parses a duration the way kool timeout / for-every do: prefer
// time.ParseDuration, else treat a bare integer as seconds. Must be positive.
func parseInterval(text string) (time.Duration, error) {
	s := strings.TrimSpace(text)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if dur, err := time.ParseDuration(s); err == nil {
		if dur <= 0 {
			return 0, fmt.Errorf("duration %q must be greater than 0", s)
		}
		return dur, nil
	}
	if secs, err := strconv.ParseInt(s, 10, 64); err == nil {
		if secs <= 0 {
			return 0, fmt.Errorf("duration %q must be greater than 0", s)
		}
		return time.Duration(secs) * time.Second, nil
	}
	return 0, fmt.Errorf("invalid duration %q", s)
}

// intervalSchedule fires every `every` from the first call, without drift.
type intervalSchedule struct {
	spec  string
	every time.Duration
}

func (s *intervalSchedule) String() string { return s.spec }

func (s *intervalSchedule) Next(after time.Time) time.Time {
	return after.Add(s.every)
}

// cronSchedule matches calendar fields in local time.
type cronSchedule struct {
	spec    string
	minute  field
	hour    field
	day     field
	month   field
	weekday field
}

func (s *cronSchedule) String() string { return s.spec }

func parseCronFields(spec string, fields []string) (Schedule, error) {
	minute, err := parseField(fields[0], "minute", minMinute, maxMinute)
	if err != nil {
		return nil, err
	}
	hour, err := parseField(fields[1], "hour", minHour, maxHour)
	if err != nil {
		return nil, err
	}
	day, err := parseField(fields[2], "day-of-month", minDay, maxDay)
	if err != nil {
		return nil, err
	}
	month, err := parseField(fields[3], "month", minMonth, maxMonth)
	if err != nil {
		return nil, err
	}
	weekday, err := parseField(fields[4], "day-of-week", minWeekday, maxWeekday)
	if err != nil {
		return nil, err
	}
	return &cronSchedule{
		spec:    spec,
		minute:  minute,
		hour:    hour,
		day:     day,
		month:   month,
		weekday: weekday,
	}, nil
}

func (s *cronSchedule) Next(after time.Time) time.Time {
	t := after.Truncate(time.Minute).Add(time.Minute)
	limit := t.AddDate(5, 0, 0)
	for t.Before(limit) {
		if !s.month.has(int(t.Month())) {
			t = startOfNextMonth(t)
			continue
		}
		if !s.dayMatches(t) {
			t = startOfNextDay(t)
			continue
		}
		if !s.hour.has(t.Hour()) {
			t = startOfNextHour(t)
			continue
		}
		if !s.minute.has(t.Minute()) {
			t = t.Add(time.Minute)
			continue
		}
		return t
	}
	return time.Time{}
}

// dayMatches applies Vixie cron's day rule: when both day-of-month and
// day-of-week are restricted, a time matches if either matches.
func (s *cronSchedule) dayMatches(t time.Time) bool {
	domMatch := s.day.has(t.Day())
	dowMatch := s.weekday.has(int(t.Weekday()))
	if s.day.any {
		return dowMatch
	}
	if s.weekday.any {
		return domMatch
	}
	return domMatch || dowMatch
}

func startOfNextHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location()).Add(time.Hour)
}

func startOfNextDay(t time.Time) time.Time {
	next := t.AddDate(0, 0, 1)
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, t.Location())
}

func startOfNextMonth(t time.Time) time.Time {
	first := t.AddDate(0, 1, 1)
	return time.Date(first.Year(), first.Month(), 1, 0, 0, 0, 0, t.Location())
}

// field is one parsed cron field.
type field struct {
	any    bool
	values []bool
	// bounds carry the declared range for error messages and lookups.
	min, max int
}

func (f field) has(v int) bool {
	if v < f.min || v > f.max {
		return false
	}
	return f.values[v]
}

func parseField(raw, name string, min, max int) (field, error) {
	f := field{min: min, max: max, values: make([]bool, max+1)}
	text := strings.TrimSpace(raw)
	if text == "" {
		return field{}, fmt.Errorf("%s field is empty", name)
	}
	if text == "*" {
		f.any = true
		for i := min; i <= max; i++ {
			f.values[i] = true
		}
		return f, nil
	}
	for _, part := range strings.Split(text, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return field{}, fmt.Errorf("%s field %q has an empty list item", name, raw)
		}
		step := 1
		span := part
		if slash := strings.Index(part, "/"); slash >= 0 {
			span = strings.TrimSpace(part[:slash])
			stepText := strings.TrimSpace(part[slash+1:])
			parsed, err := strconv.Atoi(stepText)
			if err != nil || parsed <= 0 {
				return field{}, fmt.Errorf("%s field %q: step must be a positive integer", name, part)
			}
			step = parsed
		}

		lo, hi := min, max
		switch {
		case span == "*":
			// full range
		case strings.Contains(span, "-"):
			bits := strings.SplitN(span, "-", 2)
			parsedLo, errLo := parseFieldValue(bits[0], name, min, max)
			parsedHi, errHi := parseFieldValue(bits[1], name, min, max)
			if errLo != nil {
				return field{}, errLo
			}
			if errHi != nil {
				return field{}, errHi
			}
			if parsedLo > parsedHi {
				return field{}, fmt.Errorf("%s field %q: range start is greater than end", name, part)
			}
			lo, hi = parsedLo, parsedHi
		default:
			value, err := parseFieldValue(span, name, min, max)
			if err != nil {
				return field{}, err
			}
			lo, hi = value, value
		}

		if step > 1 && lo == min && hi == max {
			// "*/n" walks from the field minimum.
			lo = min
		}
		for v := lo; v <= hi; v += step {
			f.values[normalizeWeekday(v, max)] = true
		}
	}
	return f, nil
}

func parseFieldValue(text, name string, min, max int) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		return 0, fmt.Errorf("%s field %q: %q is not a number", name, text, text)
	}
	if v < min || v > max {
		return 0, fmt.Errorf("%s field value %d out of range %d-%d", name, v, min, max)
	}
	return v, nil
}

// normalizeWeekday folds 7 (Sunday) onto 0.
func normalizeWeekday(v, max int) int {
	if max == maxWeekday && v == 7 {
		return 0
	}
	return v
}
