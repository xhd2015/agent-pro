package sessions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	traceTypes "github.com/xhd2015/agent-pro/agent_trace/types"
)

var rolloutUUIDPattern = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

type rolloutLine struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type sessionMetaPayload struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	CWD       string `json:"cwd"`
}

type eventMsgPayload struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Phase   string `json:"phase"`
}

type responseItemPayload struct {
	Type              string `json:"type"`
	Name              string `json:"name"`
	Arguments         string `json:"arguments"`
	Input             string `json:"input"`
	CallID            string `json:"call_id"`
	Output            string `json:"output"`
	EncryptedContent  string `json:"encrypted_content"`
	Text              string `json:"text"`
	Summary           string `json:"summary"`
	Role              string `json:"role"`
}

func parseSessionMeta(line string) (sessionMetaPayload, bool) {
	var row rolloutLine
	if err := json.Unmarshal([]byte(line), &row); err != nil || row.Type != "session_meta" {
		return sessionMetaPayload{}, false
	}
	var meta sessionMetaPayload
	if err := json.Unmarshal(row.Payload, &meta); err != nil {
		return sessionMetaPayload{}, false
	}
	return meta, meta.ID != ""
}

func uuidFromFilename(name string) string {
	base := strings.TrimSuffix(name, ".jsonl")
	match := rolloutUUIDPattern.FindString(base)
	return match
}

func rolloutToTraceLines(line string) []string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}

	var row rolloutLine
	if err := json.Unmarshal([]byte(trimmed), &row); err != nil {
		return nil
	}

	switch row.Type {
	case "session_meta", "turn_context":
		return nil
	case "event_msg":
		return eventMsgToTraceLines(row.Payload)
	case "response_item":
		return responseItemToTraceLines(row.Payload)
	default:
		return nil
	}
}

func eventMsgToTraceLines(payload json.RawMessage) []string {
	var p eventMsgPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil
	}
	switch p.Type {
	case "agent_message":
		if p.Phase != "commentary" && p.Phase != "final" {
			return nil
		}
		text := strings.TrimSpace(p.Message)
		if text == "" {
			return nil
		}
		return []string{marshalTraceEvent(traceTypes.TraceEvent{
			Type: "item.completed",
			Item: &traceTypes.TraceItem{
				Type: "agent_message",
				Text: text,
			},
		})}
	case "token_count", "task_started", "task_complete", "user_message", "patch_apply_end":
		return nil
	default:
		return nil
	}
}

func responseItemToTraceLines(payload json.RawMessage) []string {
	var p responseItemPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil
	}

	switch p.Type {
	case "function_call":
		if p.Name != "exec_command" {
			return nil
		}
		cmd := extractExecCommand(p.Arguments)
		if cmd == "" {
			return nil
		}
		return []string{marshalTraceEvent(traceTypes.TraceEvent{
			Type: "item.started",
			Item: &traceTypes.TraceItem{
				ID:      p.CallID,
				Type:    "command_execution",
				Command: cmd,
			},
		})}
	case "function_call_output":
		output := p.Output
		if output == "" {
			return nil
		}
		exitCode := 0
		return []string{marshalTraceEvent(traceTypes.TraceEvent{
			Type: "item.completed",
			Item: &traceTypes.TraceItem{
				ID:               p.CallID,
				Type:             "command_execution",
				AggregatedOutput: output,
				ExitCode:         &exitCode,
			},
		})}
	case "custom_tool_call":
		if p.Name != "apply_patch" {
			return nil
		}
		path := extractPatchPath(p.Input)
		if path == "" {
			return nil
		}
		return []string{marshalTraceEvent(traceTypes.TraceEvent{
			Type: "item.completed",
			Item: &traceTypes.TraceItem{
				ID:   p.CallID,
				Type: "patch",
				Changes: []traceTypes.FileChange{
					{Path: path, Kind: "edit"},
				},
			},
		})}
	case "reasoning":
		text := reasoningText(p)
		if text == "" {
			return nil
		}
		return []string{marshalTraceEvent(traceTypes.TraceEvent{
			Type: "item.completed",
			Item: &traceTypes.TraceItem{
				ID:   p.CallID,
				Type: "reasoning",
				Text: text,
			},
		})}
	case "message":
		role := strings.ToLower(strings.TrimSpace(p.Role))
		if role == "developer" || role == "system" {
			return nil
		}
		return nil
	default:
		return nil
	}
}

func reasoningText(p responseItemPayload) string {
	for _, candidate := range []string{p.Text, p.Summary} {
		if s := strings.TrimSpace(candidate); s != "" {
			return s
		}
	}
	if strings.TrimSpace(p.EncryptedContent) != "" {
		return "[Redacted]"
	}
	return ""
}

func extractExecCommand(arguments string) string {
	arguments = strings.TrimSpace(arguments)
	if arguments == "" {
		return ""
	}
	var args struct {
		Cmd string `json:"cmd"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return ""
	}
	return strings.TrimSpace(args.Cmd)
}

func extractPatchPath(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}
	var patch struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &patch); err != nil {
		return ""
	}
	return strings.TrimSpace(patch.Path)
}

func marshalTraceEvent(event traceTypes.TraceEvent) string {
	data, err := json.Marshal(event)
	if err != nil {
		return ""
	}
	return string(data)
}

func inferSessionStatus(lines []string) string {
	hasStarted := false
	hasComplete := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		var row rolloutLine
		if err := json.Unmarshal([]byte(trimmed), &row); err != nil || row.Type != "event_msg" {
			continue
		}
		var p eventMsgPayload
		if err := json.Unmarshal(row.Payload, &p); err != nil {
			continue
		}
		switch p.Type {
		case "task_started":
			hasStarted = true
		case "task_complete":
			hasComplete = true
		}
	}
	if hasComplete {
		return "completed"
	}
	if hasStarted {
		return "running"
	}
	return ""
}

func displayEventFromTraceLine(traceLine string) (DisplayEvent, bool) {
	formatted := formatTraceLine(traceLine)
	if strings.TrimSpace(formatted) == "" {
		return DisplayEvent{}, false
	}
	text := displayTextFromTraceLine(traceLine)
	return DisplayEvent{
		Kind:      displayKind(formatted),
		Text:      text,
		Formatted: formatted,
	}, true
}

func displayKind(formatted string) string {
	for _, kind := range []string{"ASSISTANT", "RUN", "EDIT", "REASONING"} {
		if strings.Contains(formatted, kind) {
			return kind
		}
	}
	if idx := strings.Index(formatted, " "); idx > 0 {
		return strings.TrimSpace(formatted[:idx])
	}
	return formatted
}

func displayTextFromTraceLine(traceLine string) string {
	var event traceTypes.TraceEvent
	if err := json.Unmarshal([]byte(traceLine), &event); err != nil || event.Item == nil {
		return ""
	}
	item := event.Item
	switch item.Type {
	case "agent_message", "message", "assistant_message", "output_text":
		return strings.TrimSpace(traceTypes.TraceItemText(item))
	case "reasoning":
		return strings.TrimSpace(item.Text)
	case "command_execution":
		if strings.TrimSpace(item.Command) != "" {
			return strings.TrimSpace(item.Command)
		}
		return strings.TrimSpace(item.AggregatedOutput)
	case "patch", "file_change":
		if len(item.Changes) > 0 {
			return strings.TrimSpace(item.Changes[0].Path)
		}
	}
	return strings.TrimSpace(traceTypes.TraceItemText(item))
}

func sessionNotFoundError(id string) error {
	return fmt.Errorf("codex session not found: %s", id)
}