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
	case types.ActionSleep:
		continue
	case types.ActionThink:
			evt.Type = EvtReasoning
			evt.Part = ReasoningPart{
				ID:   id,
				Type: PartReasoning,
				Text: e.Text,
			}
		case types.ActionMessage:
			evt.Type = EvtText
			evt.Part = TextPart{
				ID:   id,
				Type: PartText,
				Text: e.Text,
			}
		case types.ActionError:
			evt.Type = EvtError
			evt.Error = &ErrorDetail{
				Name: "Error",
				Data: &ErrorData{
					Message: e.Text,
				},
			}
		case types.ActionDone:
			evt.Type = EvtDone
			evt.Done = true
		case types.ActionStepStart:
			evt.Type = EvtStepStart
			part := StepStartPart{
				ID:   id,
				Type: PartStepStart,
			}
			if e.ToolInput != nil {
				if v, ok := e.ToolInput["sessionID"].(string); ok {
					part.SessionID = v
				}
				if v, ok := e.ToolInput["messageID"].(string); ok {
					part.MessageID = v
				}
				if v, ok := e.ToolInput["snapshot"].(string); ok {
					part.Snapshot = v
				}
			}
			evt.Part = part
		case types.ActionStepFinish:
			evt.Type = EvtStepFinish
			part := StepFinishPart{
				ID:   id,
				Type: PartStepFinish,
			}
			if e.ToolInput != nil {
				if v, ok := e.ToolInput["sessionID"].(string); ok {
					part.SessionID = v
				}
				if v, ok := e.ToolInput["messageID"].(string); ok {
					part.MessageID = v
				}
				if v, ok := e.ToolInput["reason"].(string); ok {
					part.Reason = v
				}
				if v, ok := e.ToolInput["snapshot"].(string); ok {
					part.Snapshot = v
				}
				if v, ok := e.ToolInput["cost"].(float64); ok {
					part.Cost = v
				}
				if tokens, ok := e.ToolInput["tokens"].(map[string]any); ok {
					if v, ok := tokens["input"].(float64); ok {
						part.Tokens.Input = int(v)
					}
					if v, ok := tokens["output"].(float64); ok {
						part.Tokens.Output = int(v)
					}
					if v, ok := tokens["reasoning"].(float64); ok {
						part.Tokens.Reasoning = int(v)
					}
					if cache, ok := tokens["cache"].(map[string]any); ok {
						if v, ok := cache["read"].(float64); ok {
							part.Tokens.Cache.Read = int(v)
						}
						if v, ok := cache["write"].(float64); ok {
							part.Tokens.Cache.Write = int(v)
						}
					}
				}
			}
			evt.Part = part
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
		Type: EvtToolUse,
		Part: ToolUsePart{
			ID:     id,
			Type:   PartTool,
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

func FromOpencode(events []Event, _ string) []types.AgentEvent {
	var result []types.AgentEvent
	for _, e := range events {
		evt := types.AgentEvent{
			Timestamp: e.Timestamp,
		}
		switch e.Type {
		case EvtReasoning:
			evt.Type = types.ActionThink
			if p := lookupReasoningPart(e.Part); p != nil {
				evt.ID = p.ID
				evt.Text = p.Text
			}
		case EvtText:
			evt.Type = types.ActionMessage
			if p := lookupTextPart(e.Part); p != nil {
				evt.ID = p.ID
				evt.Text = p.Text
			}
		case EvtError:
			evt.Type = types.ActionError
			if e.Error != nil && e.Error.Data != nil {
				evt.Text = e.Error.Data.Message
			}
		case EvtDone:
			evt.Type = types.ActionDone
		case EvtStepStart:
			evt.Type = types.ActionStepStart
			if p := lookupStepStartPart(e.Part); p != nil {
				evt.ID = p.ID
				evt.ToolInput = map[string]any{
					"sessionID": p.SessionID,
					"messageID": p.MessageID,
					"snapshot":  p.Snapshot,
				}
			}
		case EvtStepFinish:
			evt.Type = types.ActionStepFinish
			if p := lookupStepFinishPart(e.Part); p != nil {
				evt.ID = p.ID
				tokens := map[string]any{
					"input":     float64(p.Tokens.Input),
					"output":    float64(p.Tokens.Output),
					"reasoning": float64(p.Tokens.Reasoning),
					"cache": map[string]any{
						"read":  float64(p.Tokens.Cache.Read),
						"write": float64(p.Tokens.Cache.Write),
					},
				}
				evt.ToolInput = map[string]any{
					"sessionID": p.SessionID,
					"messageID": p.MessageID,
					"reason":    p.Reason,
					"snapshot":  p.Snapshot,
					"cost":      p.Cost,
					"tokens":    tokens,
				}
			}
		case EvtToolUse:
			evt.Type = types.ActionToolCall
			if p := lookupToolUsePart(e.Part); p != nil {
				evt.ID = p.ID
				evt.Tool = p.Tool
				if p.State.Input != nil {
					evt.ToolInput = p.State.Input
				}
				exitCode := p.State.ExitCode
				evt.Mock = &types.MockConfig{
					Output:   p.State.Output,
					Stderr:   p.State.Stderr,
					ExitCode: &exitCode,
				}
			}
		}
		result = append(result, evt)
	}
	return result
}

func lookupReasoningPart(part any) *ReasoningPart {
	if p, ok := part.(ReasoningPart); ok {
		return &p
	}
	if m, ok := part.(map[string]any); ok {
		p := ReasoningPart{}
		if v, ok := m["id"].(string); ok {
			p.ID = v
		}
		if v, ok := m["type"].(string); ok {
			p.Type = PartType(v)
		}
		if v, ok := m["text"].(string); ok {
			p.Text = v
		}
		return &p
	}
	return nil
}

func lookupTextPart(part any) *TextPart {
	if p, ok := part.(TextPart); ok {
		return &p
	}
	if m, ok := part.(map[string]any); ok {
		p := TextPart{}
		if v, ok := m["id"].(string); ok {
			p.ID = v
		}
		if v, ok := m["type"].(string); ok {
			p.Type = PartType(v)
		}
		if v, ok := m["text"].(string); ok {
			p.Text = v
		}
		return &p
	}
	return nil
}

func lookupStepStartPart(part any) *StepStartPart {
	if p, ok := part.(StepStartPart); ok {
		return &p
	}
	if m, ok := part.(map[string]any); ok {
		p := StepStartPart{}
		if v, ok := m["id"].(string); ok {
			p.ID = v
		}
		if v, ok := m["sessionID"].(string); ok {
			p.SessionID = v
		}
		if v, ok := m["messageID"].(string); ok {
			p.MessageID = v
		}
		if v, ok := m["type"].(string); ok {
			p.Type = PartType(v)
		}
		if v, ok := m["snapshot"].(string); ok {
			p.Snapshot = v
		}
		return &p
	}
	return nil
}

func lookupStepFinishPart(part any) *StepFinishPart {
	if p, ok := part.(StepFinishPart); ok {
		return &p
	}
	if m, ok := part.(map[string]any); ok {
		p := StepFinishPart{}
		if v, ok := m["id"].(string); ok {
			p.ID = v
		}
		if v, ok := m["sessionID"].(string); ok {
			p.SessionID = v
		}
		if v, ok := m["messageID"].(string); ok {
			p.MessageID = v
		}
		if v, ok := m["type"].(string); ok {
			p.Type = PartType(v)
		}
		if v, ok := m["reason"].(string); ok {
			p.Reason = v
		}
		if v, ok := m["snapshot"].(string); ok {
			p.Snapshot = v
		}
		if v, ok := m["cost"].(float64); ok {
			p.Cost = v
		}
		if tokens, ok := m["tokens"].(map[string]any); ok {
			if v, ok := tokens["input"].(float64); ok {
				p.Tokens.Input = int(v)
			}
			if v, ok := tokens["output"].(float64); ok {
				p.Tokens.Output = int(v)
			}
			if v, ok := tokens["reasoning"].(float64); ok {
				p.Tokens.Reasoning = int(v)
			}
			if cache, ok := tokens["cache"].(map[string]any); ok {
				if v, ok := cache["read"].(float64); ok {
					p.Tokens.Cache.Read = int(v)
				}
				if v, ok := cache["write"].(float64); ok {
					p.Tokens.Cache.Write = int(v)
				}
			}
		}
		return &p
	}
	return nil
}

func lookupToolUsePart(part any) *ToolUsePart {
	if p, ok := part.(ToolUsePart); ok {
		return &p
	}
	if m, ok := part.(map[string]any); ok {
		p := ToolUsePart{}
		if v, ok := m["id"].(string); ok {
			p.ID = v
		}
		if v, ok := m["type"].(string); ok {
			p.Type = PartType(v)
		}
		if v, ok := m["callID"].(string); ok {
			p.CallID = v
		}
		if v, ok := m["tool"].(string); ok {
			p.Tool = v
		}
		if state, ok := m["state"].(map[string]any); ok {
			if v, ok := state["input"].(map[string]any); ok {
				p.State.Input = v
			}
			if v, ok := state["output"].(string); ok {
				p.State.Output = v
			}
			if v, ok := state["stderr"].(string); ok {
				p.State.Stderr = v
			}
			if v, ok := state["exit_code"].(float64); ok {
				p.State.ExitCode = int(v)
			}
			if v, ok := state["status"].(string); ok {
				p.State.Status = v
			}
		}
		return &p
	}
	return nil
}
