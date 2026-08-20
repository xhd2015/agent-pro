package agenttty

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// openBindState is durable bind progress at sessions/<id>/bind.json
// (same shape as agentui.OpenGrokBindState so status can report "binding").
type openBindState struct {
	State           string `json:"state"` // in_progress|ok|failed
	StartedAt       string `json:"started_at,omitempty"`
	FinishedAt      string `json:"finished_at,omitempty"`
	Error           string `json:"error,omitempty"`
	RunnerSessionID string `json:"runner_session_id,omitempty"`
}

func writeOpenBindJSON(home, sessionID string, st openBindState) {
	home = strings.TrimSpace(home)
	sessionID = strings.TrimSpace(sessionID)
	if home == "" || sessionID == "" {
		return
	}
	dir := filepath.Join(home, "sessions", sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	data, err := json.Marshal(st)
	if err != nil {
		return
	}
	path := filepath.Join(dir, "bind.json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}
