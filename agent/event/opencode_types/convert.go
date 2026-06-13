package opencode_types

import (
	"fmt"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
)

func ToOpencode(events []types.AgentEvent, sessionID string) []map[string]any {
	var result []map[string]any
	for i, e := range events {
		id := e.ID
		if id == "" {
			id = fmt.Sprintf("evt_%d", i+1)
		}
		switch e.Type {
		case types.ActionThink:
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
		case types.ActionMessage:
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
		case types.ActionError:
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
		case types.ActionDone:
			evt := map[string]any{
				"type": "done",
				"done": true,
			}
			if sessionID != "" {
				evt["sessionID"] = sessionID
			}
			result = append(result, evt)
		case types.ActionToolCall:
			result = append(result, convertToolCallToOpencode(e, id, sessionID))
		}
	}
	return result
}

func convertToolCallToOpencode(e types.AgentEvent, id, sessionID string) map[string]any {
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
