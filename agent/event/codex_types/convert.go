package codex_types

import (
	"fmt"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
)

func ToCodex(events []types.AgentEvent) []Event {
	var result []Event
	for i, e := range events {
		id := e.ID
		if id == "" {
			id = fmt.Sprintf("evt_%d", i+1)
		}
		switch e.Type {
		case types.ActionThink:
			started := Event{
				Type: EventStarted,
				Item: &EventItem{ID: id, Type: ItemReasoning},
			}
			completed := Event{
				Type: EventCompleted,
				Item: &EventItem{
					ID:     id,
					Type:   ItemReasoning,
					Text:   e.Text,
					Status: "completed",
				},
			}
			result = append(result, started, completed)
		case types.ActionToolCall:
			result = append(result, convertToolCallToCodex(e, id)...)
		case types.ActionMessage:
			completed := Event{
				Type: EventCompleted,
				Item: &EventItem{
					ID:     id,
					Type:   ItemMessage,
					Text:   e.Text,
					Status: "completed",
				},
			}
			result = append(result, completed)
		case types.ActionError:
			errEvent := Event{
				Type:    EventError,
				Message: e.Text,
			}
			result = append(result, errEvent)
		case types.ActionDone:
		}
	}
	return result
}

func convertToolCallToCodex(e types.AgentEvent, id string) []Event {
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

func convertBashToCodex(e types.AgentEvent, id string) []Event {
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

	started := Event{
		Type: EventStarted,
		Item: &EventItem{ID: id, Type: ItemCommandExecution},
	}
	ec := exitCode
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:               id,
			Type:             ItemCommandExecution,
			AggregatedOutput: stdout,
			ExitCode:         &ec,
			Status:           "completed",
		},
	}
	return []Event{started, completed}
}

func convertReadToCodex(e types.AgentEvent, id string) []Event {
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

	started := Event{
		Type: EventStarted,
		Item: &EventItem{ID: id, Type: ItemCommandExecution},
	}
	ec := exitCode
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:               id,
			Type:             ItemCommandExecution,
			AggregatedOutput: output,
			ExitCode:         &ec,
			Status:           "completed",
		},
	}
	return []Event{started, completed}
}

func convertWriteToCodex(e types.AgentEvent, id string) []Event {
	if e.Mock != nil {
		if len(e.Mock.Changes) > 0 {
			var changes []FileChange
			for _, c := range e.Mock.Changes {
				changes = append(changes, FileChange{Path: c.Path, Kind: c.Kind})
			}
			faketoolexec.ExecuteWriteMock()
			started := Event{
				Type: EventStarted,
				Item: &EventItem{ID: id, Type: ItemFileChange},
			}
			completed := Event{
				Type: EventCompleted,
				Item: &EventItem{
					ID:      id,
					Type:    ItemFileChange,
					Status:  "completed",
					Changes: changes,
				},
			}
			return []Event{started, completed}
		}
		faketoolexec.ExecuteWriteMock()
		started := Event{
			Type: EventStarted,
			Item: &EventItem{ID: id, Type: ItemFileChange},
		}
		completed := Event{
			Type: EventCompleted,
			Item: &EventItem{
				ID:     id,
				Type:   ItemFileChange,
				Status: "completed",
			},
		}
		return []Event{started, completed}
	}

	path, _ := e.ToolInput["path"].(string)
	content, _ := e.ToolInput["content"].(string)
	faketoolexec.ExecuteWrite(path, content)

	started := Event{
		Type: EventStarted,
		Item: &EventItem{ID: id, Type: ItemFileChange},
	}
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:     id,
			Type:   ItemFileChange,
			Status: "completed",
			Changes: []FileChange{
				{Path: path, Kind: "add"},
			},
		},
	}
	return []Event{started, completed}
}

func convertGrepToCodex(e types.AgentEvent, id string) []Event {
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

	started := Event{
		Type: EventStarted,
		Item: &EventItem{ID: id, Type: ItemCommandExecution},
	}
	ec := exitCode
	completed := Event{
		Type: EventCompleted,
		Item: &EventItem{
			ID:               id,
			Type:             ItemCommandExecution,
			Command:          pattern,
			AggregatedOutput: output,
			ExitCode:         &ec,
			Status:           "completed",
		},
	}
	return []Event{started, completed}
}
