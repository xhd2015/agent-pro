package knowledgesink

import (
	"context"
	"strings"
	"time"
)

const staleError = "stale: no ping for over 1m"

// IsStaleRunning reports whether a running manifest's last_ping is older than StaleAfter.
// Empty last_ping while running is treated as stale only after callers have had a chance
// to write the first ping — if LastPing is empty, use Started-equivalent: treat as stale
// when we cannot prove a recent ping (conservative: stale).
func IsStaleRunning(m *Manifest, now time.Time) bool {
	if m == nil || !strings.EqualFold(strings.TrimSpace(m.Status), statusRunning) {
		return false
	}
	if now.IsZero() {
		now = time.Now()
	}
	ping, err := ParseTime(m.LastPing)
	if err != nil || ping.IsZero() {
		return true
	}
	return now.Sub(ping) > StaleAfter
}

// ReconcileStaleRunning rewrites a stale running manifest to failed. Returns true if rewritten.
func ReconcileStaleRunning(sessionDir string, m *Manifest, now time.Time) (bool, error) {
	if !IsStaleRunning(m, now) {
		return false, nil
	}
	if now.IsZero() {
		now = time.Now()
	}
	m.Status = statusFailed
	m.Error = staleError
	if err := WriteManifest(sessionDir, m); err != nil {
		return false, err
	}
	return true, nil
}

// LatchRunning sets status=running, refreshes last_ping, optional pid (0 = leave / clear).
func LatchRunning(sessionDir, marcusSessionID string, pid int, now time.Time) (*Manifest, error) {
	if now.IsZero() {
		now = time.Now()
	}
	m, _ := LoadManifest(sessionDir)
	if m == nil {
		m = &Manifest{
			Version:       1,
			MarcusSessionID: marcusSessionID,
			NextSinkIndex: 0,
			LastSinkIndex: -1,
		}
	}
	if strings.TrimSpace(m.MarcusSessionID) == "" {
		m.MarcusSessionID = marcusSessionID
	}
	m.Status = statusRunning
	m.Error = ""
	m.LastPing = FormatTime(now)
	if pid > 0 {
		m.Pid = pid
	}
	if err := WriteManifest(sessionDir, m); err != nil {
		return nil, err
	}
	return m, nil
}

// TouchPing updates last_ping (and pid if > 0) while status is still running.
func TouchPing(sessionDir string, pid int, now time.Time) error {
	m, err := LoadManifest(sessionDir)
	if err != nil || m == nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(m.Status), statusRunning) {
		return nil
	}
	if now.IsZero() {
		now = time.Now()
	}
	m.LastPing = FormatTime(now)
	if pid > 0 {
		m.Pid = pid
	}
	return WriteManifest(sessionDir, m)
}

// StartPingLoop refreshes last_ping every PingInterval until ctx is cancelled.
func StartPingLoop(ctx context.Context, sessionDir string, pid int, nowFn func() time.Time) {
	if ctx == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(PingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				if nowFn != nil {
					now = nowFn()
				}
				_ = TouchPing(sessionDir, pid, now)
			}
		}
	}()
}
