package cronspec

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func mustParse(t *testing.T, spec string) Schedule {
	t.Helper()
	s, err := Parse(spec)
	if err != nil {
		t.Fatalf("Parse(%q): %v", spec, err)
	}
	return s
}

func localTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation("2006-01-02 15:04", value, time.Local)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return parsed
}

func TestParseCronNextFire(t *testing.T) {
	cases := []struct {
		spec  string
		after string
		want  string
	}{
		{"*/5 * * * *", "2026-09-15 10:02", "2026-09-15 10:05"},
		{"*/5 * * * *", "2026-09-15 10:05", "2026-09-15 10:10"},
		{"0 * * * *", "2026-09-15 10:02", "2026-09-15 11:00"},
		{"30 9 * * 1-5", "2026-09-15 10:02", "2026-09-16 09:30"}, // Tue 10:02 → Wed 09:30
		{"0 0 1 * *", "2026-09-15 10:02", "2026-10-01 00:00"},
		{"15 3 * * *", "2026-12-31 23:59", "2027-01-01 03:15"},
		{"0 0 * * 0", "2026-09-15 10:02", "2026-09-20 00:00"}, // next Sunday
		{"0 0 * * 7", "2026-09-15 10:02", "2026-09-20 00:00"}, // 7 == Sunday
		{"@hourly", "2026-09-15 10:02", "2026-09-15 11:00"},
		{"@daily", "2026-09-15 10:02", "2026-09-16 00:00"},
		{"@midnight", "2026-09-15 10:02", "2026-09-16 00:00"},
		{"@weekly", "2026-09-15 10:02", "2026-09-20 00:00"},
		{"@monthly", "2026-09-15 10:02", "2026-10-01 00:00"},
	}
	for _, tc := range cases {
		t.Run(tc.spec+" from "+tc.after, func(t *testing.T) {
			s := mustParse(t, tc.spec)
			got := s.Next(localTime(t, tc.after))
			want := localTime(t, tc.want)
			if !got.Equal(want) {
				t.Fatalf("Next = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
			}
		})
	}
}

func TestParseCronDayOfMonthOrWeekday(t *testing.T) {
	// Vixie cron: with both restricted, either one matching fires the job.
	s := mustParse(t, "0 0 15 * 1")
	got := s.Next(localTime(t, "2026-09-16 00:00"))
	// Next Monday is 2026-09-21, before the next 15th (2026-10-15).
	want := localTime(t, "2026-09-21 00:00")
	if !got.Equal(want) {
		t.Fatalf("Next = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}

func TestParseIntervalForms(t *testing.T) {
	// Interval fire times are offsets from the previous fire, so they carry the
	// sub-minute part rather than snapping to a minute boundary.
	cases := []struct {
		spec  string
		after string
		want  string
	}{
		{"@every 10m", "2026-09-15 10:00:00", "2026-09-15 10:10:00"},
		{"@every 90s", "2026-09-15 10:00:00", "2026-09-15 10:01:30"},
		{"@every 90s", "2026-09-15 10:00:30", "2026-09-15 10:02:00"},
		{"10m", "2026-09-15 10:00:00", "2026-09-15 10:10:00"},
		{"300", "2026-09-15 10:00:00", "2026-09-15 10:05:00"},
	}
	for _, tc := range cases {
		t.Run(tc.spec+" from "+tc.after, func(t *testing.T) {
			s := mustParse(t, tc.spec)
			after, err := time.ParseInLocation("2006-01-02 15:04:05", tc.after, time.Local)
			if err != nil {
				t.Fatal(err)
			}
			want, err := time.ParseInLocation("2006-01-02 15:04:05", tc.want, time.Local)
			if err != nil {
				t.Fatal(err)
			}
			if got := s.Next(after); !got.Equal(want) {
				t.Fatalf("Next = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
			}
		})
	}
}

func TestParseIntervalKeepsCadence(t *testing.T) {
	// Intervals chain from the previous scheduled fire, so a slow cycle does not
	// add drift.
	s := mustParse(t, "@every 60s")
	first := localTime(t, "2026-09-15 10:00")
	second := s.Next(first)
	third := s.Next(second)
	wantThird := first.Add(2 * time.Minute)
	if !third.Equal(wantThird) {
		t.Fatalf("third fire = %s, want %s", third.Format(time.RFC3339), wantThird.Format(time.RFC3339))
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		spec string
		want string
	}{
		{"", "5 fields"},
		{"every 5 min", "5 fields"},
		{"* * * * * *", "got 6 fields"},
		{"60 * * * *", "out of range 0-59"},
		{"* 24 * * *", "out of range 0-23"},
		{"* * 0 * *", "out of range 1-31"},
		{"* * * 13 *", "out of range 1-12"},
		{"* * * * 8", "out of range 0-7"},
		{"*/0 * * * *", "step must be a positive integer"},
		{"5-1 * * * *", "range start is greater than end"},
		{"x * * * *", "is not a number"},
		{"@yearly", "unknown macro"},
		{"@every", "@every requires a duration"},
		{"@every 0s", "must be greater than 0"},
		{"5x", "not a duration"},
	}
	for _, tc := range cases {
		t.Run(tc.spec, func(t *testing.T) {
			_, err := Parse(tc.spec)
			if err == nil {
				t.Fatalf("Parse(%q) succeeded, want error", tc.spec)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Parse(%q) error = %q, want it to contain %q", tc.spec, err.Error(), tc.want)
			}
		})
	}
}

func TestNextNeverFires(t *testing.T) {
	s := mustParse(t, "0 0 30 2 *") // 30 February
	if got := s.Next(localTime(t, "2026-09-15 10:00")); !got.IsZero() {
		t.Fatalf("Next = %s, want zero time", got.Format(time.RFC3339))
	}
}

func TestRunMaxRuns(t *testing.T) {
	calls := 0
	s := mustParse(t, "@every 1ms")
	err := Run(context.Background(), s, RunOptions{MaxRuns: 3}, func(context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if calls != 3 {
		t.Fatalf("fn called %d times, want 3", calls)
	}
}

func TestRunContinuesAfterCycleError(t *testing.T) {
	calls := 0
	logs := []string{}
	s := mustParse(t, "@every 1ms")
	err := Run(context.Background(), s, RunOptions{
		MaxRuns: 2,
		Logf:    func(format string, args ...any) { logs = append(logs, format) },
	}, func(context.Context) error {
		calls++
		return errFake
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if calls != 2 {
		t.Fatalf("fn called %d times, want 2", calls)
	}
	found := false
	for _, line := range logs {
		if strings.Contains(line, "cycle failed") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a cycle-failed log line, got %v", logs)
	}
}

func TestRunStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	s := mustParse(t, "@every 1h")
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, s, RunOptions{}, func(context.Context) error {
			calls++
			cancel() // cancel from inside the first cycle
			return nil
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}
	if calls != 1 {
		t.Fatalf("fn called %d times, want 1", calls)
	}
}

var errFake = errors.New("fake failure")
