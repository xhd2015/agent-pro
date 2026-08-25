package knowledgesink

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SinkedRow is one session that has completed a sink (manifest last_sink_at set).
type SinkedRow struct {
	SessionID  string    `json:"session_id"`
	LastSinkAt time.Time `json:"last_sink_at"`
	Status     string    `json:"status"` // idle|running|failed (manifest)
	LastMRURL  string    `json:"last_mr_url,omitempty"`
}

// HasSinkHistory reports whether a manifest records a completed sink.
func HasSinkHistory(m *Manifest) bool {
	return m != nil && strings.TrimSpace(m.LastSinkAt) != ""
}

// ListSinkedSessions scans stateDir/knowledge-sink/*/manifest.json and returns
// sessions with last_sink_at set, newest last_sink_at first.
func ListSinkedSessions(stateDir string) ([]SinkedRow, error) {
	root := Root(stateDir)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var rows []SinkedRow
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sessionDir := filepath.Join(root, e.Name())
		m, lerr := LoadManifest(sessionDir)
		if lerr != nil || !HasSinkHistory(m) {
			continue
		}
		id := strings.TrimSpace(m.MarcusSessionID)
		if id == "" {
			id = e.Name()
		}
		var at time.Time
		if t, perr := ParseTime(m.LastSinkAt); perr == nil {
			at = t
		}
		status := strings.TrimSpace(m.Status)
		if status == "" {
			status = statusIdle
		}
		rows = append(rows, SinkedRow{
			SessionID:  id,
			LastSinkAt: at,
			Status:     status,
			LastMRURL:  strings.TrimSpace(m.LastMRURL),
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if !rows[i].LastSinkAt.Equal(rows[j].LastSinkAt) {
			return rows[i].LastSinkAt.After(rows[j].LastSinkAt)
		}
		return rows[i].SessionID < rows[j].SessionID
	})
	return rows, nil
}
