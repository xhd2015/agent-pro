package knowledgesink

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Auto-sink window and timing (daemon scheduler; CLI list uses the same window).
const (
	AutoSinkWindow   = 7 * 24 * time.Hour
	AutoSinkGrace    = 10 * time.Minute
	AutoSinkInterval = time.Hour
)

// SessionMeta is one Marcus/local-bot session for auto-sink selection.
type SessionMeta struct {
	ID        string
	UpdatedAt time.Time
	Archived  bool
}

// AutoSinkableRow is one session the hourly auto-sink would consider.
type AutoSinkableRow struct {
	SessionID string    `json:"session_id"`
	UpdatedAt time.Time `json:"updated_at"`
	State     string    `json:"state"`
	Why       string    `json:"why,omitempty"`
}

// IsAutoSinkable reports whether StatusView is eligible for an automatic sink.
// Enabled covers ready / behind / failed-with-work; sunk / running / unavailable are not.
func IsAutoSinkable(v *StatusView) bool {
	return v != nil && v.Enabled
}

// AutoSinkWhy is a short reason column for --show-auto-sinkable-sessions.
func AutoSinkWhy(v *StatusView) string {
	if v == nil {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(v.State)) {
	case StateReady:
		return "never sunk"
	case StateBehind:
		return firstNonEmpty(v.Help, "new messages since last sink")
	case StateFailed:
		return firstNonEmpty(v.Error, v.Help, "previous sink failed")
	default:
		return firstNonEmpty(v.Help, v.Error, v.State)
	}
}

// WithinAutoSinkWindow reports whether updatedAt is non-zero and within window of now.
func WithinAutoSinkWindow(updatedAt, now time.Time, window time.Duration) bool {
	if updatedAt.IsZero() || now.IsZero() {
		return false
	}
	if window <= 0 {
		window = AutoSinkWindow
	}
	age := now.Sub(updatedAt)
	return age >= 0 && age <= window
}

// FilterAndSortAutoSinkable returns sinkable sessions within the window, oldest UpdatedAt first.
// statusFor is called only for non-archived sessions inside the window; nil views or errors skip the row.
func FilterAndSortAutoSinkable(
	sessions []SessionMeta,
	now time.Time,
	window time.Duration,
	statusFor func(id string) (*StatusView, error),
) ([]AutoSinkableRow, error) {
	if statusFor == nil {
		return nil, fmt.Errorf("statusFor is required")
	}
	if window <= 0 {
		window = AutoSinkWindow
	}
	if now.IsZero() {
		now = time.Now()
	}

	type ranked struct {
		row AutoSinkableRow
	}
	var out []ranked
	for _, s := range sessions {
		id := strings.TrimSpace(s.ID)
		if id == "" || s.Archived {
			continue
		}
		if !WithinAutoSinkWindow(s.UpdatedAt, now, window) {
			continue
		}
		view, err := statusFor(id)
		if err != nil || !IsAutoSinkable(view) {
			continue
		}
		out = append(out, ranked{row: AutoSinkableRow{
			SessionID: id,
			UpdatedAt: s.UpdatedAt,
			State:     view.State,
			Why:       AutoSinkWhy(view),
		}})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].row.UpdatedAt.Equal(out[j].row.UpdatedAt) {
			return out[i].row.UpdatedAt.Before(out[j].row.UpdatedAt)
		}
		return out[i].row.SessionID < out[j].row.SessionID
	})
	rows := make([]AutoSinkableRow, len(out))
	for i := range out {
		rows[i] = out[i].row
	}
	return rows, nil
}

// PickOldestAutoSinkable returns the first row from FilterAndSortAutoSinkable, or nil when none.
func PickOldestAutoSinkable(
	sessions []SessionMeta,
	now time.Time,
	window time.Duration,
	statusFor func(id string) (*StatusView, error),
) (*AutoSinkableRow, error) {
	rows, err := FilterAndSortAutoSinkable(sessions, now, window, statusFor)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	row := rows[0]
	return &row, nil
}
