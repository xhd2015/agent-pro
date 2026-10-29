// Package sessions inspects DSH v3 durable logs without restoring or modifying them.
package sessions

import "encoding/json"

// Options selects a storage root. SessionsDir overrides Home, DSH_HOME, and ~/.dsh.
type Options struct {
	Home        string
	SessionsDir string
}

// Summary contains durable facts, not process liveness. Provider and Model are the last used route.
type Summary struct {
	ID              string `json:"id"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
	Cwd             string `json:"cwd"`
	ParentSession   string `json:"parent_session"`
	Origin          string `json:"origin"`
	IsSeeded        bool   `json:"is_seeded"`
	AgentPreset     string `json:"agent_preset"`
	DelegationDepth int64  `json:"delegation_depth"`
	FormatVersion   int    `json:"format_version"`
	Path            string `json:"path"`
	Encoding        string `json:"encoding"`
	SizeBytes       int64  `json:"size_bytes"`
	Outcome         string `json:"outcome"`
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	Title           string `json:"title,omitempty"`
}

// Inspection retains physical JSON records; it does not reconstruct current model history.
type Inspection struct {
	Summary
	Header json.RawMessage `json:"header"`
	Events []Event         `json:"events"`
}

// Event exposes common envelope fields and retains all original extension metadata in Raw.
type Event struct {
	Type string          `json:"type"`
	Seq  int64           `json:"seq"`
	Time int64           `json:"time"`
	Data json.RawMessage `json:"data"`
	Raw  json.RawMessage `json:"-"`
}
