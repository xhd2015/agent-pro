package types

type ActionType string

const (
	ActionThink      ActionType = "think"
	ActionToolCall   ActionType = "tool_call"
	ActionMessage    ActionType = "message"
	ActionError      ActionType = "error"
	ActionDone       ActionType = "done"
	ActionStepStart  ActionType = "step_start"
	ActionStepFinish ActionType = "step_finish"
	ActionSleep      ActionType = "sleep"
)

type AgentEvent struct {
	ID        string         `json:"id,omitempty"`
	Type      ActionType     `json:"type"`
	Timestamp int64          `json:"timestamp,omitempty"`
	Text      string         `json:"text,omitempty"`
	Tool      string         `json:"tool,omitempty"`
	ToolInput map[string]any `json:"tool_input,omitempty"`
	Output    string         `json:"output,omitempty"`
	Stderr    string         `json:"stderr,omitempty"`
	ExitCode  *int           `json:"exit_code,omitempty"`
	DelayMs   int            `json:"delay_ms,omitempty"`
	Mock      *MockConfig    `json:"mock,omitempty"`
	Changes   []FileChange   `json:"changes,omitempty"`
}

type FileChange struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type MockConfig struct {
	Output   string       `json:"output,omitempty"`
	Stderr   string       `json:"stderr,omitempty"`
	ExitCode *int         `json:"exit_code,omitempty"`
	Content  string       `json:"content,omitempty"`
	Changes  []FileChange `json:"changes,omitempty"`
}
