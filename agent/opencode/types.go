package opencode

import "time"

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
