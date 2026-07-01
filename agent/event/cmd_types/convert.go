package cmd_types

import (
	"encoding/json"
	"fmt"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// FromCmd converts a sequence of cmd session events into the canonical AgentEvent stream.
// It walks the native events in order, emitting zero or more canonical events per native event.
// Tool results are merged into the preceding tool-call by matching toolCallId.
func FromCmd(events []Event, sessionID string) []types.AgentEvent {
	result := []types.AgentEvent{}

	// Track pending tool-calls by toolCallId so tool results can be merged.
	type pendingTool struct {
		idx        int // index into result
		toolCallID string
	}
	var pendingToolCalls []*pendingTool

	for _, e := range events {
		content, err := parseContent(e.Content)
		if err != nil {
			// Skip events with unparseable content
			continue
		}
		for _, block := range content {
			switch e.Role {
			case RoleAssistant:
				switch block.Type {
				case BlockTypeReasoning:
					result = append(result, types.AgentEvent{
						Type: types.ActionThink,
						Text: block.Text,
					})
				case BlockTypeText:
					result = append(result, types.AgentEvent{
						Type: types.ActionMessage,
						Text: block.Text,
					})
				case BlockTypeToolCall:
					toolInput := make(map[string]any)
					if len(block.Input) > 0 {
						if err := json.Unmarshal(block.Input, &toolInput); err != nil {
							toolInput = nil
						}
					}
					canonicalIdx := len(result)
					result = append(result, types.AgentEvent{
						Type:      types.ActionToolCall,
						Tool:      block.ToolName,
						ToolInput: toolInput,
					})
					pendingToolCalls = append(pendingToolCalls, &pendingTool{
						idx:        canonicalIdx,
						toolCallID: block.ToolCallID,
					})
				default:
					// unknown block types are skipped
				}
			case RoleTool:
				switch block.Type {
				case BlockTypeToolResult:
					// Find matching pending tool-call by toolCallId
					for _, pt := range pendingToolCalls {
						if pt.toolCallID == block.ToolCallID {
							if block.Output != nil {
								result[pt.idx].Output = block.Output.Value
							}
							break
						}
					}
				default:
					// unknown block types are skipped
				}
			case RoleUser:
				// User messages mark the start of a step
				if block.Type == BlockTypeText {
					result = append(result, types.AgentEvent{
						Type: types.ActionStepStart,
					})
				}
			}
		}
	}
	return result
}

// ToCmd converts a canonical AgentEvent stream into cmd session events.
func ToCmd(events []types.AgentEvent, sessionID string) []Event {
	result := []Event{}
	for _, e := range events {
		var ev Event
		switch e.Type {
		case types.ActionThink:
			content := []ContentBlock{
				{Type: BlockTypeReasoning, Text: e.Text},
			}
			contentRaw, _ := json.Marshal(content)
			ev = Event{
				Role:    RoleAssistant,
				Content: contentRaw,
			}
		case types.ActionMessage:
			content := []ContentBlock{
				{Type: BlockTypeText, Text: e.Text},
			}
			contentRaw, _ := json.Marshal(content)
			ev = Event{
				Role:    RoleAssistant,
				Content: contentRaw,
			}
		case types.ActionToolCall:
			toolInputRaw, _ := json.Marshal(e.ToolInput)
			var output *ToolOutput
			if e.Output != "" {
				output = &ToolOutput{
					Type:  "text",
					Value: e.Output,
				}
			}
			content := []ContentBlock{
				{
					Type:       BlockTypeToolCall,
					ToolCallID: e.ID,
					ToolName:   e.Tool,
					Input:      toolInputRaw,
					Output:     output,
				},
			}
			contentRaw, _ := json.Marshal(content)
			ev = Event{
				Role:    RoleAssistant,
				Content: contentRaw,
			}
		case types.ActionStepStart:
			content := []ContentBlock{}
			contentRaw, _ := json.Marshal(content)
			ev = Event{
				Role:    RoleUser,
				Content: contentRaw,
			}
		case types.ActionStepFinish:
			// Map to a minimal user event
			content := []ContentBlock{}
			contentRaw, _ := json.Marshal(content)
			ev = Event{
				Role:    RoleUser,
				Content: contentRaw,
			}
		case types.ActionError:
			content := []ContentBlock{
				{Type: BlockTypeText, Text: fmt.Sprintf("error: %s", e.Text)},
			}
			contentRaw, _ := json.Marshal(content)
			ev = Event{
				Role:    RoleAssistant,
				Content: contentRaw,
			}
		case types.ActionDone:
			content := []ContentBlock{
				{Type: BlockTypeText, Text: e.Text},
			}
			contentRaw, _ := json.Marshal(content)
			ev = Event{
				Role:    RoleAssistant,
				Content: contentRaw,
			}
		default:
			// unknown action types are skipped
			continue
		}
		if sessionID != "" {
			ev.SessionID = sessionID
		}
		result = append(result, ev)
	}
	return result
}

// parseContent parses the content field which can be either a string or a JSON array of ContentBlock.
func parseContent(raw json.RawMessage) ([]ContentBlock, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	// Try as array first
	var blocks []ContentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		return blocks, nil
	}
	// Try as string
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		if str == "" {
			return nil, nil
		}
		return []ContentBlock{
			{Type: BlockTypeText, Text: str},
		}, nil
	}
	return nil, fmt.Errorf("cannot parse content: %s", string(raw))
}
