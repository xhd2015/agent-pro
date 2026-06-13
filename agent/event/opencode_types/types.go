package opencode_types

type Event struct {
	Type      string       `json:"type"`
	Timestamp int64        `json:"timestamp,omitempty"`
	SessionID string       `json:"sessionID,omitempty"`
	Part      any          `json:"part,omitempty"`
	Error     *ErrorDetail `json:"error,omitempty"`
	Done      bool         `json:"done,omitempty"`
}

type ReasoningPart struct {
	ID   string     `json:"id"`
	Type string     `json:"type"`
	Text string     `json:"text"`
	Time *TimeRange `json:"time,omitempty"`
}

type TextPart struct {
	ID   string     `json:"id"`
	Type string     `json:"type"`
	Text string     `json:"text"`
	Time *TimeRange `json:"time,omitempty"`
}

type ErrorDetail struct {
	Name string     `json:"name"`
	Data *ErrorData `json:"data"`
}

type ErrorData struct {
	Message string `json:"message"`
}

type ToolUsePart struct {
	ID     string       `json:"id"`
	Type   string       `json:"type"`
	CallID string       `json:"callID"`
	Tool   string       `json:"tool"`
	State  ToolUseState `json:"state"`
}

type ToolUseState struct {
	Input    map[string]any `json:"input,omitempty"`
	Output   string         `json:"output,omitempty"`
	Stderr   string         `json:"stderr,omitempty"`
	ExitCode int            `json:"exit_code"`
	Error    string         `json:"error,omitempty"`
	Status   string         `json:"status,omitempty"`
	Title    string         `json:"title,omitempty"`
	Time     *TimeRange     `json:"time,omitempty"`
}

type StepStartPart struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionID"`
	MessageID string `json:"messageID"`
	Type      string `json:"type"`
	Snapshot  string `json:"snapshot,omitempty"`
}

type StepFinishPart struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionID"`
	MessageID string `json:"messageID"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Snapshot  string `json:"snapshot,omitempty"`
	Cost      float64 `json:"cost"`
	Tokens    Tokens `json:"tokens"`
}

type Tokens struct {
	Input     int        `json:"input"`
	Output    int        `json:"output"`
	Reasoning int        `json:"reasoning"`
	Cache     CacheTokens `json:"cache"`
}

type CacheTokens struct {
	Read  int `json:"read"`
	Write int `json:"write"`
}

type TimeRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end,omitempty"`
}
