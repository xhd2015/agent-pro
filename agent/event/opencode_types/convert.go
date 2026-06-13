package opencode_types

import (
	"fmt"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
)

func ToOpencode(events []types.AgentEvent, sessionID string) []Event {
	var result []Event
	for i, e := range events {
		id := e.ID
		if id == "" {
			id = fmt.Sprintf("evt_%d", i+1)
		}
		evt := Event{}
		if sessionID != "" {
			evt.SessionID = sessionID
		}
		evt.Timestamp = e.Timestamp
		switch e.Type {
		case types.ActionThink:
			evt.Type = "reasoning"
			evt.Part = ReasoningPart{
				ID:   id,
				Type: "reasoning",
				Text: e.Text,
			}
		case types.ActionMessage:
			evt.Type = "text"
			evt.Part = TextPart{
				ID:   id,
				Type: "text",
				Text: e.Text,
			}
		case types.ActionError:
			evt.Type = "error"
			evt.Error = &ErrorDetail{
				Name: "Error",
				Data: &ErrorData{
					Message: e.Text,
				},
			}
		case types.ActionDone:
			evt.Type = "done"
			evt.Done = true
		case types.ActionStepStart:
			evt.Type = "step_start"
			evt.Part = StepStartPart{
				ID:   id,
				Type: "step-start",
			}
		case types.ActionStepFinish:
			evt.Type = "step_finish"
			evt.Part = StepFinishPart{
				ID:   id,
				Type: "step-finish",
			}
		case types.ActionToolCall:
			result = append(result, convertToolCallToOpencode(e, id, sessionID))
			continue
		}
		result = append(result, evt)
	}
	return result
}

func convertToolCallToOpencode(e types.AgentEvent, id, sessionID string) Event {
	tool := e.Tool

	state := ToolUseState{}

	if e.ToolInput != nil {
		input := make(map[string]any)
		for k, v := range e.ToolInput {
			input[k] = v
		}
		state.Input = input
	}

	if e.Mock != nil {
		mock := *e.Mock
		switch tool {
		case "bash":
			stdout, stderr, ec := faketoolexec.ExecuteBashMock("", mock)
			state.Output = stdout
			state.ExitCode = ec
			if stderr != "" {
				state.Stderr = stderr
			}
		case "read":
			content := faketoolexec.ExecuteReadMock(mock)
			state.Output = content
			ec := 0
			if mock.ExitCode != nil {
				ec = *mock.ExitCode
			}
			state.ExitCode = ec
		case "write":
			faketoolexec.ExecuteWriteMock()
			state.Output = mock.Output
			ec := 0
			if mock.ExitCode != nil {
				ec = *mock.ExitCode
			}
			state.ExitCode = ec
		case "grep":
			output, ec := faketoolexec.ExecuteGrepMock(mock)
			state.Output = output
			state.ExitCode = ec
		}
	} else {
		switch tool {
		case "bash":
			command, _ := e.ToolInput["command"].(string)
			workdir, _ := e.ToolInput["workdir"].(string)
			stdout, stderr, ec, _ := faketoolexec.ExecuteBash(command, workdir, nil)
			stdout = strings.TrimRight(stdout, "\n")
			state.Output = stdout
			state.ExitCode = ec
			if stderr != "" {
				state.Stderr = stderr
			}
		case "read":
			path, _ := e.ToolInput["path"].(string)
			content, err := faketoolexec.ExecuteRead(path)
			if err != nil {
				state.Error = err.Error()
				state.ExitCode = 1
			} else {
				state.Output = content
				state.ExitCode = 0
			}
		case "write":
			path, _ := e.ToolInput["path"].(string)
			content, _ := e.ToolInput["content"].(string)
			err := faketoolexec.ExecuteWrite(path, content)
			if err != nil {
				state.Error = err.Error()
				state.ExitCode = 1
			} else {
				state.ExitCode = 0
			}
		case "grep":
			pattern, _ := e.ToolInput["pattern"].(string)
			searchPath, _ := e.ToolInput["path"].(string)
			output, ec, _ := faketoolexec.ExecuteGrep(pattern, searchPath)
			state.Output = output
			state.ExitCode = ec
		}
	}
	state.Status = "completed"

	evt := Event{
		Type: "tool_use",
		Part: ToolUsePart{
			ID:     id,
			Type:   "tool",
			CallID: id,
			Tool:   tool,
			State:  state,
		},
	}
	if sessionID != "" {
		evt.SessionID = sessionID
	}
	return evt
}
