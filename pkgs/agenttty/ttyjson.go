package agenttty

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

// TTYSnapshot is the denormalized TTY cross-ref written to sessions/.../tty.json.
type TTYSnapshot struct {
	RunnerID          string `json:"runner_id"`
	AgentSessionID    string `json:"agent_session_id"`
	TerminalSessionID string `json:"terminal_session_id"`
	ListenAddr        string `json:"listen_addr"`
	PID               int    `json:"pid"`
	CreatedAt         string `json:"created_at"`
	ScreenStatus      string `json:"screen_status,omitempty"`
	Alive             bool   `json:"alive"`
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func ttyJSONPath(home, runner, agentSessionID string) string {
	return filepath.Join(home, "sessions", runner, agentSessionID, "tty.json")
}

// WriteTTYJSON dual-writes the denormalized TTY snapshot.
func WriteTTYJSON(home string, snap TTYSnapshot) error {
	if snap.AgentSessionID == "" {
		return nil
	}
	dir := filepath.Dir(ttyJSONPath(home, snap.RunnerID, snap.AgentSessionID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if snap.CreatedAt == "" {
		snap.CreatedAt = nowRFC3339()
	}
	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return os.WriteFile(ttyJSONPath(home, snap.RunnerID, snap.AgentSessionID), data, 0644)
}

func writeTTYJSONOnStart(home, runnerID, agentSessionID string, entry ttywatch.RegistryEntry, provider Provider) error {
	screenStatus := "unknown"
	if provider.DetectScreenStatus != nil {
		screenStatus = provider.DetectScreenStatus(nil)
	}
	return WriteTTYJSON(home, TTYSnapshot{
		RunnerID:          runnerID,
		AgentSessionID:    agentSessionID,
		TerminalSessionID: entry.SessionID,
		ListenAddr:        entry.ListenAddr,
		PID:               entry.PID,
		CreatedAt:         entry.CreatedAt,
		ScreenStatus:      screenStatus,
		Alive:             true,
	})
}

func dualWriteAfterRegistry(home, runnerID, agentSessionID, terminalSessionID string, provider Provider) {
	if agentSessionID == "" || terminalSessionID == "" {
		return
	}
	cfg := ttywatch.RegistryConfig{Home: home, Subdir: provider.RegistryDir}
	for i := 0; i < 20; i++ {
		entry, err := ttywatch.ReadRegistry(cfg, terminalSessionID)
		if err == nil && entry != nil {
			_ = writeTTYJSONOnStart(home, runnerID, agentSessionID, *entry, provider)
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
}