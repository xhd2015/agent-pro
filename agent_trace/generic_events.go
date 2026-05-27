package agent_trace

import (
	"encoding/json"
	"strings"

	"github.com/xhd2015/agent-traces/agent_trace/types"
)

type genericTraceAdapter struct{}

func init() {
	types.RegisterAgentTraceAdapter(1000, genericTraceAdapter{})
}

func (genericTraceAdapter) Name() string {
	return "generic"
}

func (genericTraceAdapter) Parse(raw json.RawMessage) (types.AgentTraceParsedEvent, bool) {
	var event types.TraceEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return types.AgentTraceParsedEvent{}, false
	}
	switch event.Type {
	case "assistant":
		if text := types.TraceMessageText(event.Message); text != "" {
			return types.AgentTraceParsedEvent{Message: &types.AgentTraceMessage{
				Role:    types.RoleAssistant,
				Content: text,
			}}, true
		}
	case "result":
		if text := strings.TrimSpace(event.Result); text != "" {
			return types.AgentTraceParsedEvent{Message: &types.AgentTraceMessage{
				Role:    types.RoleAssistant,
				Content: text,
			}}, true
		}
	}
	return types.AgentTraceParsedEvent{}, false
}
