package opencode

import (
	"encoding/json"
	"time"
)

type Model struct {
	ID         string `json:"id"`
	ProviderID string `json:"providerID,omitempty"`
	Name       string `json:"name,omitempty"`
}

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	ProjectID string    `json:"projectId,omitempty"`
	Directory string    `json:"directory,omitempty"`
}

type SessionOpts struct {
	Model    string   `json:"model,omitempty"`
	Agent    string   `json:"agent,omitempty"`
	Dir      string   `json:"dir,omitempty"`
	File     []string `json:"file,omitempty"`
	Command  string   `json:"command,omitempty"`
	Variant  string   `json:"variant,omitempty"`
	NoSubAgents bool `json:"noSubAgents,omitempty"`
}

type StreamEvent struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp,omitempty"`
	SessionID string `json:"sessionID,omitempty"`

	Text      string         `json:"text,omitempty"`
	ToolUse   *ToolUseEvent  `json:"toolUse,omitempty"`
	Error     string         `json:"error,omitempty"`
	Reasoning string         `json:"reasoning,omitempty"`
	File      *FileChange    `json:"file,omitempty"`
	Done      bool           `json:"done,omitempty"`
}

type ToolUseEvent struct {
	ID      string `json:"id"`
	Tool    string `json:"tool"`
	Status  string `json:"status"`
	Summary string `json:"summary,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

type FileChange struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type StreamCallback func(event StreamEvent)

type SessionExport struct {
	Messages []SessionExportMessage `json:"messages"`
}

type SessionExportMessage struct {
	Info  SessionExportMessageInfo `json:"info"`
	Parts []SessionExportPart      `json:"parts"`
}

type SessionExportMessageInfo struct {
	Role string `json:"role"`
}

type SessionExportPart struct {
	Type      string              `json:"type"`
	Text      string              `json:"text,omitempty"`
	Tool      string              `json:"tool,omitempty"`
	CallID    string              `json:"callID,omitempty"`
	ID        string              `json:"id,omitempty"`
	SessionID string              `json:"sessionID,omitempty"`
	MessageID string              `json:"messageID,omitempty"`
	State     *SessionExportState `json:"state,omitempty"`
	Snapshot  string              `json:"snapshot,omitempty"`
	Reason    string              `json:"reason,omitempty"`
	Hash      string              `json:"hash,omitempty"`
	Files     []string            `json:"files,omitempty"`
	Time      *SessionExportTime  `json:"time,omitempty"`
	Tokens    *SessionExportTokens `json:"tokens,omitempty"`
	Cost      float64             `json:"cost,omitempty"`
}

type SessionExportState struct {
	Status   string            `json:"status,omitempty"`
	Title    string            `json:"title,omitempty"`
	Output   string            `json:"output,omitempty"`
	Input    json.RawMessage   `json:"input,omitempty"`
	Metadata json.RawMessage   `json:"metadata,omitempty"`
	Time     *SessionExportTime `json:"time,omitempty"`
}

type SessionExportTime struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type SessionExportTokens struct {
	Total     int64                  `json:"total"`
	Input     int64                  `json:"input"`
	Output    int64                  `json:"output"`
	Reasoning int64                  `json:"reasoning"`
	Cache     *SessionExportCache    `json:"cache,omitempty"`
}

type SessionExportCache struct {
	Write int64 `json:"write"`
	Read  int64 `json:"read"`
}
