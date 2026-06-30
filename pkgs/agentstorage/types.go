package agentstorage

import (
	"os"
	"path/filepath"
	"time"
)

// Config persisted at config.json under the store home.
type Config struct {
	DefaultAgentRunner string `json:"default_agent_runner"`
	DefaultModel       string `json:"default_model"`
	LastSession        string `json:"last_session"`
}

// SessionMeta is stored in sessions/<runner>/<session_id>/meta.json.
type SessionMeta struct {
	Runner          string `json:"runner"`
	SessionID       string `json:"session_id"`
	RunnerSessionID string `json:"runner_session_id,omitempty"`
	Status          string `json:"status"`
	Workspace       string `json:"workspace,omitempty"`
	Model           string `json:"model,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// Session wraps session metadata for API consumers.
type Session struct {
	Meta SessionMeta
}

// Message is a queued user follow-up in messages.jsonl.
type Message struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	SessionID string `json:"session_id"`
	CreatedAt string `json:"created_at"`
}

func resolveHome(constructorHome string) (string, error) {
	if v := os.Getenv("AGENT_RUN_HOME"); v != "" {
		return filepath.Clean(v), nil
	}
	home := constructorHome
	if home == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		home = filepath.Join(dir, ".agent-run")
	}
	return filepath.Clean(home), nil
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}