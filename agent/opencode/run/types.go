package run

type SessionOpts struct {
	Model       string   `json:"model,omitempty"`
	Agent       string   `json:"agent,omitempty"`
	Dir         string   `json:"dir,omitempty"`
	File        []string `json:"file,omitempty"`
	Command     string   `json:"command,omitempty"`
	Variant     string   `json:"variant,omitempty"`
	NoSubAgents bool     `json:"noSubAgents,omitempty"`
	Thinking    bool     `json:"thinking,omitempty"`
}

type StreamEventTime struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type StreamEvent struct {
	Type      string           `json:"type"`
	Timestamp int64            `json:"timestamp,omitempty"`
	SessionID string           `json:"sessionID,omitempty"`
	Text      string           `json:"text,omitempty"`
	ToolUse   *ToolUseEvent    `json:"toolUse,omitempty"`
	Error     string           `json:"error,omitempty"`
	Reasoning string           `json:"reasoning,omitempty"`
	ReasoningTime *StreamEventTime `json:"reasoningTime,omitempty"`
	File      *FileChange      `json:"file,omitempty"`
	Done      bool             `json:"done,omitempty"`
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
