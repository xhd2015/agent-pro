package codex_types

import (
	"fmt"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	fakeagent "github.com/xhd2015/agent-pro/pkgs/fake-agent"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
)

func ToCodex(events []types.AgentEvent) []fakeagent.Event {
	var result []fakeagent.Event
	for i, e := range events {
		id := e.ID
		if id == "" {
			id = fmt.Sprintf("evt_%d", i+1)
		}
		switch e.Type {
		case types.ActionThink:
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
		case types.ActionToolCall:
			result = append(result, convertToolCallToCodex(e, id)...)
		case types.ActionMessage:
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
		case types.ActionError:
			errEvent := fakeagent.Event{
				Type:    fakeagent.EventError,
				Message: e.Text,
			}
			result = append(result, errEvent)
		case types.ActionDone:
		}
	}
	return result
}

func convertToolCallToCodex(e types.AgentEvent, id string) []fakeagent.Event {
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

func convertBashToCodex(e types.AgentEvent, id string) []fakeagent.Event {
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

func convertReadToCodex(e types.AgentEvent, id string) []fakeagent.Event {
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

func convertWriteToCodex(e types.AgentEvent, id string) []fakeagent.Event {
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

func convertGrepToCodex(e types.AgentEvent, id string) []fakeagent.Event {
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
