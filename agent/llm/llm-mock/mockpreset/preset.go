package mockpreset

import (
	"fmt"
	"io"
	"sort"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// Entry describes one named mock-events preset in the catalog.
type Entry struct {
	Name        string
	Description string
}

const (
	textThinkMessage      = "preset:think:think-message"
	textMessageThinkMsg   = "preset:message:think-message"
	textMessageSimple     = "preset:message:simple"
	textThinkMulti1       = "preset:think:multi-think-1"
	textThinkMulti2       = "preset:think:multi-think-2"
	textMessageMulti      = "preset:message:multi-think"
	textThinkToolMessage  = "preset:think:think-tool-message"
	textMessageToolMsg    = "preset:message:think-tool-message"
)

var catalog = []Entry{
	{Name: "simple", Description: "one message event"},
	{Name: "think-message", Description: "think then message"},
	{Name: "multi-think", Description: "think, think, then message"},
	{Name: "tool-bash", Description: "one bash tool_call"},
	{Name: "tool-read", Description: "one read tool_call"},
	{Name: "think-tool-message", Description: "think, tool_call, then message"},
}

var presets = map[string][]types.AgentEvent{
	"simple": {
		{Type: types.ActionMessage, Text: textMessageSimple},
	},
	"think-message": {
		{Type: types.ActionThink, Text: textThinkMessage},
		{Type: types.ActionMessage, Text: textMessageThinkMsg},
	},
	"multi-think": {
		{Type: types.ActionThink, Text: textThinkMulti1},
		{Type: types.ActionThink, Text: textThinkMulti2},
		{Type: types.ActionMessage, Text: textMessageMulti},
	},
	"tool-bash": {
		{
			ID:        "preset-tool-bash",
			Type:      types.ActionToolCall,
			Tool:      "bash",
			ToolInput: map[string]any{"command": "echo preset-bash"},
		},
	},
	"tool-read": {
		{
			ID:        "preset-tool-read",
			Type:      types.ActionToolCall,
			Tool:      "read",
			ToolInput: map[string]any{"path": "preset-read-target.txt"},
		},
	},
	"think-tool-message": {
		{Type: types.ActionThink, Text: textThinkToolMessage},
		{
			ID:        "preset-tool-bash-inline",
			Type:      types.ActionToolCall,
			Tool:      "bash",
			ToolInput: map[string]any{"command": "echo preset-inline-bash"},
		},
		{Type: types.ActionMessage, Text: textMessageToolMsg},
	},
}

// List returns the MVP preset catalog in stable name order.
func List() []Entry {
	out := make([]Entry, len(catalog))
	copy(out, catalog)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

// Resolve returns a copy of the AgentEvent sequence for a named preset.
func Resolve(name string) ([]types.AgentEvent, error) {
	events, ok := presets[name]
	if !ok {
		return nil, fmt.Errorf("unknown mock-events-preset: %s", name)
	}
	out := make([]types.AgentEvent, len(events))
	copy(out, events)
	return out, nil
}

// PrintList writes the preset catalog to w (one line per preset).
func PrintList(w io.Writer) {
	for _, entry := range List() {
		fmt.Fprintf(w, "%s\t%s\n", entry.Name, entry.Description)
	}
}