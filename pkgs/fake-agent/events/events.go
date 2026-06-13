package events

import (
	"fmt"
	"strings"

	fakeagent "github.com/xhd2015/agent-pro/pkgs/fake-agent"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
)

type ActionType string

const (
	ActionThink    ActionType = "think"
	ActionToolCall ActionType = "tool_call"
	ActionMessage  ActionType = "message"
	ActionError    ActionType = "error"
	ActionDone     ActionType = "done"
)

type AgentEvent struct {
	ID        string                   `json:"id,omitempty"`
	Type      ActionType               `json:"type"`
	Text      string                   `json:"text,omitempty"`
	Tool      string                   `json:"tool,omitempty"`
	ToolInput map[string]any           `json:"tool_input,omitempty"`
	Output    string                   `json:"output,omitempty"`
	Stderr    string                   `json:"stderr,omitempty"`
	ExitCode  *int                     `json:"exit_code,omitempty"`
	Mock      *faketoolexec.MockConfig `json:"mock,omitempty"`
	Changes   []FileChange             `json:"changes,omitempty"`
}

type FileChange struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

func ToCodex(events []AgentEvent) []fakeagent.Event {
	var result []fakeagent.Event
	for i, e := range events {
		id := e.ID
		if id == "" {
			id = fmt.Sprintf("evt_%d", i+1)
		}
		switch e.Type {
		case ActionThink:
			started := fakeagent.Event{
				Type: fakeagent.EventStarted,
				Item: &fakeagent.EventItem{ID: id, Type: fakeagent.ItemReasoning},
			}
			completed := fakeagent.Event{
				Type: fakeagent.EventCompleted,
				Item: &fakeagent.EventItem{
					ID:     id,
					Type:   fakeagent.ItemReasoning,
					Text:   e.Text,
					Status: "completed",
				},
			}
			result = append(result, started, completed)
		case ActionToolCall:
			result = append(result, convertToolCallToCodex(e, id)...)
		case ActionMessage:
			completed := fakeagent.Event{
				Type: fakeagent.EventCompleted,
				Item: &fakeagent.EventItem{
					ID:     id,
					Type:   fakeagent.ItemMessage,
					Text:   e.Text,
					Status: "completed",
				},
			}
			result = append(result, completed)
		case ActionError:
			errEvent := fakeagent.Event{
				Type:    fakeagent.EventError,
				Message: e.Text,
			}
			result = append(result, errEvent)
		case ActionDone:
		}
	}
	return result
}

func convertToolCallToCodex(e AgentEvent, id string) []fakeagent.Event {
	switch e.Tool {
	case "bash":
		return convertBashToCodex(e, id)
	case "read":
		return convertReadToCodex(e, id)
	case "write":
		return convertWriteToCodex(e, id)
	case "grep":
		return convertGrepToCodex(e, id)
	default:
		return convertBashToCodex(e, id)
	}
}

func convertBashToCodex(e AgentEvent, id string) []fakeagent.Event {
	command, _ := e.ToolInput["command"].(string)

	var stdout, stderr string
	var exitCode int

	if e.Mock != nil {
		stdout, stderr, exitCode = faketoolexec.ExecuteBashMock(command, *e.Mock)
	} else {
		workdir, _ := e.ToolInput["workdir"].(string)
		stdout, stderr, exitCode, _ = faketoolexec.ExecuteBash(command, workdir, nil)
	}
	_ = stderr
	stdout = strings.TrimRight(stdout, "\n")

	started := fakeagent.Event{
		Type: fakeagent.EventStarted,
		Item: &fakeagent.EventItem{ID: id, Type: fakeagent.ItemCommandExecution},
	}
	ec := exitCode
	completed := fakeagent.Event{
		Type: fakeagent.EventCompleted,
		Item: &fakeagent.EventItem{
			ID:               id,
			Type:             fakeagent.ItemCommandExecution,
			AggregatedOutput: stdout,
			ExitCode:         &ec,
			Status:           "completed",
		},
	}
	return []fakeagent.Event{started, completed}
}

func convertReadToCodex(e AgentEvent, id string) []fakeagent.Event {
	var output string
	var exitCode int

	if e.Mock != nil {
		output = faketoolexec.ExecuteReadMock(*e.Mock)
		if e.Mock.ExitCode != nil {
			exitCode = *e.Mock.ExitCode
		}
	} else {
		path, _ := e.ToolInput["path"].(string)
		content, err := faketoolexec.ExecuteRead(path)
		if err != nil {
			exitCode = 1
		} else {
			output = content
		}
	}

	started := fakeagent.Event{
		Type: fakeagent.EventStarted,
		Item: &fakeagent.EventItem{ID: id, Type: fakeagent.ItemCommandExecution},
	}
	ec := exitCode
	completed := fakeagent.Event{
		Type: fakeagent.EventCompleted,
		Item: &fakeagent.EventItem{
			ID:               id,
			Type:             fakeagent.ItemCommandExecution,
			AggregatedOutput: output,
			ExitCode:         &ec,
			Status:           "completed",
		},
	}
	return []fakeagent.Event{started, completed}
}

func convertWriteToCodex(e AgentEvent, id string) []fakeagent.Event {
	if e.Mock != nil {
		if len(e.Mock.Changes) > 0 {
			var changes []fakeagent.FileChange
			for _, c := range e.Mock.Changes {
				changes = append(changes, fakeagent.FileChange{Path: c.Path, Kind: c.Kind})
			}
			faketoolexec.ExecuteWriteMock()
			started := fakeagent.Event{
				Type: fakeagent.EventStarted,
				Item: &fakeagent.EventItem{ID: id, Type: fakeagent.ItemFileChange},
			}
			completed := fakeagent.Event{
				Type: fakeagent.EventCompleted,
				Item: &fakeagent.EventItem{
					ID:      id,
					Type:    fakeagent.ItemFileChange,
					Status:  "completed",
					Changes: changes,
				},
			}
			return []fakeagent.Event{started, completed}
		}
		faketoolexec.ExecuteWriteMock()
		started := fakeagent.Event{
			Type: fakeagent.EventStarted,
			Item: &fakeagent.EventItem{ID: id, Type: fakeagent.ItemFileChange},
		}
		completed := fakeagent.Event{
			Type: fakeagent.EventCompleted,
			Item: &fakeagent.EventItem{
				ID:     id,
				Type:   fakeagent.ItemFileChange,
				Status: "completed",
			},
		}
		return []fakeagent.Event{started, completed}
	}

	path, _ := e.ToolInput["path"].(string)
	content, _ := e.ToolInput["content"].(string)
	faketoolexec.ExecuteWrite(path, content)

	started := fakeagent.Event{
		Type: fakeagent.EventStarted,
		Item: &fakeagent.EventItem{ID: id, Type: fakeagent.ItemFileChange},
	}
	completed := fakeagent.Event{
		Type: fakeagent.EventCompleted,
		Item: &fakeagent.EventItem{
			ID:     id,
			Type:   fakeagent.ItemFileChange,
			Status: "completed",
			Changes: []fakeagent.FileChange{
				{Path: path, Kind: "add"},
			},
		},
	}
	return []fakeagent.Event{started, completed}
}

func convertGrepToCodex(e AgentEvent, id string) []fakeagent.Event {
	var output string
	var exitCode int
	var pattern string

	if e.Mock != nil {
		output, exitCode = faketoolexec.ExecuteGrepMock(*e.Mock)
	} else {
		pattern, _ = e.ToolInput["pattern"].(string)
		searchPath, _ := e.ToolInput["path"].(string)
		output, exitCode, _ = faketoolexec.ExecuteGrep(pattern, searchPath)
	}

	started := fakeagent.Event{
		Type: fakeagent.EventStarted,
		Item: &fakeagent.EventItem{ID: id, Type: fakeagent.ItemCommandExecution},
	}
	ec := exitCode
	completed := fakeagent.Event{
		Type: fakeagent.EventCompleted,
		Item: &fakeagent.EventItem{
			ID:               id,
			Type:             fakeagent.ItemCommandExecution,
			Command:          pattern,
			AggregatedOutput: output,
			ExitCode:         &ec,
			Status:           "completed",
		},
	}
	return []fakeagent.Event{started, completed}
}

func ToOpencode(events []AgentEvent, sessionID string) []map[string]any {
	var result []map[string]any
	for i, e := range events {
		id := e.ID
		if id == "" {
			id = fmt.Sprintf("evt_%d", i+1)
		}
		switch e.Type {
		case ActionThink:
			evt := map[string]any{
				"type": "reasoning",
				"part": map[string]any{
					"id":   id,
					"type": "reasoning",
					"text": e.Text,
				},
			}
			if sessionID != "" {
				evt["sessionID"] = sessionID
			}
			result = append(result, evt)
		case ActionMessage:
			evt := map[string]any{
				"type": "text",
				"part": map[string]any{
					"id":   id,
					"type": "text",
					"text": e.Text,
				},
			}
			if sessionID != "" {
				evt["sessionID"] = sessionID
			}
			result = append(result, evt)
		case ActionError:
			evt := map[string]any{
				"type": "error",
				"error": map[string]any{
					"name": "Error",
					"data": map[string]any{
						"message": e.Text,
					},
				},
			}
			if sessionID != "" {
				evt["sessionID"] = sessionID
			}
			result = append(result, evt)
		case ActionDone:
			evt := map[string]any{
				"type": "done",
				"done": true,
			}
			if sessionID != "" {
				evt["sessionID"] = sessionID
			}
			result = append(result, evt)
		case ActionToolCall:
			result = append(result, convertToolCallToOpencode(e, id, sessionID))
		}
	}
	return result
}

func convertToolCallToOpencode(e AgentEvent, id, sessionID string) map[string]any {
	tool := e.Tool

	state := make(map[string]any)

	if e.ToolInput != nil {
		input := make(map[string]any)
		for k, v := range e.ToolInput {
			input[k] = v
		}
		state["input"] = input
	}

	if e.Mock != nil {
		mock := *e.Mock
		switch tool {
		case "bash":
			stdout, stderr, ec := faketoolexec.ExecuteBashMock("", mock)
			state["output"] = stdout
			state["exit_code"] = ec
			if stderr != "" {
				state["stderr"] = stderr
			}
		case "read":
			content := faketoolexec.ExecuteReadMock(mock)
			state["output"] = content
			ec := 0
			if mock.ExitCode != nil {
				ec = *mock.ExitCode
			}
			state["exit_code"] = ec
		case "write":
			faketoolexec.ExecuteWriteMock()
			state["output"] = mock.Output
			ec := 0
			if mock.ExitCode != nil {
				ec = *mock.ExitCode
			}
			state["exit_code"] = ec
		case "grep":
			output, ec := faketoolexec.ExecuteGrepMock(mock)
			state["output"] = output
			state["exit_code"] = ec
		}
	} else {
		switch tool {
		case "bash":
			command, _ := e.ToolInput["command"].(string)
			workdir, _ := e.ToolInput["workdir"].(string)
			stdout, stderr, ec, _ := faketoolexec.ExecuteBash(command, workdir, nil)
			stdout = strings.TrimRight(stdout, "\n")
			state["output"] = stdout
			state["exit_code"] = ec
			if stderr != "" {
				state["stderr"] = stderr
			}
		case "read":
			path, _ := e.ToolInput["path"].(string)
			content, err := faketoolexec.ExecuteRead(path)
			if err != nil {
				state["error"] = err.Error()
				state["exit_code"] = 1
			} else {
				state["output"] = content
				state["exit_code"] = 0
			}
		case "write":
			path, _ := e.ToolInput["path"].(string)
			content, _ := e.ToolInput["content"].(string)
			err := faketoolexec.ExecuteWrite(path, content)
			if err != nil {
				state["error"] = err.Error()
				state["exit_code"] = 1
			} else {
				state["exit_code"] = 0
			}
		case "grep":
			pattern, _ := e.ToolInput["pattern"].(string)
			searchPath, _ := e.ToolInput["path"].(string)
			output, ec, _ := faketoolexec.ExecuteGrep(pattern, searchPath)
			state["output"] = output
			state["exit_code"] = ec
		}
	}
	state["status"] = "completed"

	evt := map[string]any{
		"type": "tool_use",
		"part": map[string]any{
			"id":    id,
			"type":  "tool",
			"tool":  tool,
			"state": state,
		},
	}
	if sessionID != "" {
		evt["sessionID"] = sessionID
	}
	return evt
}
