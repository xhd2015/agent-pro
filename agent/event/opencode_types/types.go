package opencode_types

type Event struct {
	Type      string       `json:"type"`
	SessionID string       `json:"sessionID,omitempty"`
	Part      any          `json:"part,omitempty"`
	Error     *ErrorDetail `json:"error,omitempty"`
	Done      bool         `json:"done,omitempty"`
}

type ReasoningPart struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text"`
}

type TextPart struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text"`
}

type ErrorDetail struct {
	Name string     `json:"name"`
	Data *ErrorData `json:"data"`
}

type ErrorData struct {
	Message string `json:"message"`
}

type ToolUsePart struct {
	ID    string       `json:"id"`
	Type  string       `json:"type"`
	Tool  string       `json:"tool"`
	State ToolUseState `json:"state"`
}

type ToolUseState struct {
	Input    map[string]any `json:"input,omitempty"`
	Output   string         `json:"output,omitempty"`
	Stderr   string         `json:"stderr,omitempty"`
	ExitCode int            `json:"exit_code"`
	Error    string         `json:"error,omitempty"`
	Status   string         `json:"status,omitempty"`
}
