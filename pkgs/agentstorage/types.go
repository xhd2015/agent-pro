package agentstorage

import (
	"encoding/json"
	"fmt"
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

// SessionMeta is stored in sessions/<session_id>/meta.json.
type SessionMeta struct {
	Runner            string `json:"runner"`
	SessionID         string `json:"session_id"`
	InitialPrompt     string `json:"initial_prompt,omitempty"`
	RunnerSessionID   string `json:"runner_session_id,omitempty"`
	TerminalSessionID string `json:"terminal_session_id,omitempty"`
	Status            string `json:"status"`
	Workspace         string `json:"workspace,omitempty"`
	Model             string `json:"model,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

func (m *SessionMeta) UnmarshalJSON(data []byte) error {
	type alias SessionMeta
	var raw struct {
		alias
		CreatedAt flexibleJSONText `json:"created_at,omitempty"`
		UpdatedAt flexibleJSONText `json:"updated_at,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = SessionMeta(raw.alias)
	m.CreatedAt = raw.CreatedAt.String()
	m.UpdatedAt = raw.UpdatedAt.String()
	return nil
}

type flexibleJSONText string

func (v *flexibleJSONText) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*v = flexibleJSONText(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*v = flexibleJSONText(n.String())
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*v = flexibleJSONText(fmt.Sprint(b))
		return nil
	}
	if string(data) == "null" {
		*v = ""
		return nil
	}
	return fmt.Errorf("unsupported text value: %s", string(data))
}

func (v flexibleJSONText) String() string {
	return string(v)
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
