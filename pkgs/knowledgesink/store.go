package knowledgesink

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// TimeLayout is human-readable local time with numeric offset + zone name.
const TimeLayout = "2006-01-02 15:04:05 -0700 MST"

const (
	statusIdle    = "idle"
	statusRunning = "running"
	statusFailed  = "failed"
)

// Heartbeat while status=running. ~4s interval ⇒ 1m ≈ 14 missed pings.
const (
	PingInterval = 4 * time.Second
	StaleAfter   = time.Minute
)

// Manifest is session-level cursor/status under knowledge-sink/<id>/manifest.json.
// status (+ last_ping) is the single source of truth for UI and all callers.
type Manifest struct {
	Version                     int      `json:"version"`
	MarcusSessionID             string   `json:"marcus_session_id"`
	GrokSessionID               string   `json:"grok_session_id,omitempty"`
	LastSinkAt                  string   `json:"last_sink_at,omitempty"`
	LastSinkMaxMessageTimestamp string   `json:"last_sink_max_message_timestamp,omitempty"`
	NextSinkIndex               int      `json:"next_sink_index"`
	LastSinkIndex               int      `json:"last_sink_index"`
	Status                      string   `json:"status"`
	Error                       string   `json:"error,omitempty"`
	Pid                         int      `json:"pid,omitempty"`       // info only; not used for liveness
	LastPing                    string   `json:"last_ping,omitempty"` // TimeLayout; heartbeat while running
	LastPaths                   []string `json:"last_paths,omitempty"`
	LastHubPaths                []string `json:"last_hub_paths,omitempty"`
	LastBranch                  string   `json:"last_branch,omitempty"`
	LastMRURL                   string   `json:"last_mr_url,omitempty"`
	LastCommit                  string   `json:"last_commit,omitempty"`
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(time.Local).Format(TimeLayout)
}

func ParseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	return time.Parse(TimeLayout, s)
}

func Root(stateDir string) string {
	return filepath.Join(strings.TrimSpace(stateDir), "knowledge-sink")
}

var sessionIDSafe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func SanitizeSessionID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	out := sessionIDSafe.ReplaceAllString(id, "_")
	out = strings.Trim(out, "._-")
	if out == "" {
		return "session"
	}
	if len(out) > 180 {
		out = out[:180]
	}
	return out
}

func SessionDir(stateDir, marcusSessionID string) string {
	return filepath.Join(Root(stateDir), SanitizeSessionID(marcusSessionID))
}

func ManifestPath(sessionDir string) string {
	return filepath.Join(sessionDir, "manifest.json")
}

func RunDir(sessionDir string, index int) string {
	return filepath.Join(sessionDir, fmt.Sprintf("sink-%d", index))
}

func LoadManifest(sessionDir string) (*Manifest, error) {
	path := ManifestPath(sessionDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &m, nil
}

func WriteManifest(sessionDir string, m *Manifest) error {
	if m == nil {
		return fmt.Errorf("nil manifest")
	}
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return err
	}
	if m.Version == 0 {
		m.Version = 1
	}
	if strings.TrimSpace(m.Status) == "" {
		m.Status = statusIdle
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := ManifestPath(sessionDir) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, ManifestPath(sessionDir))
}

// TipAfterMax reports whether tip is strictly after lastSinkMax (new messages).
func TipAfterMax(tip, lastSinkMax time.Time) bool {
	if tip.IsZero() {
		return true
	}
	if lastSinkMax.IsZero() {
		return true
	}
	return tip.After(lastSinkMax)
}
