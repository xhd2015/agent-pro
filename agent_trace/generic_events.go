package agent_trace

import (
	"encoding/json"
	"strings"
)

type genericTraceAdapter struct{}

func init() {
	RegisterAgentTraceAdapter(1000, genericTraceAdapter{})
}

func (genericTraceAdapter) Name() string {
	return "generic"
}

func (genericTraceAdapter) Parse(raw json.RawMessage) (AgentTraceParsedEvent, bool) {
	var event traceEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return AgentTraceParsedEvent{}, false
	}
	switch event.Type {
	case "assistant":
		if text := traceMessageText(event.Message); text != "" {
			return AgentTraceParsedEvent{Message: &AgentTraceMessage{
				Role:    "assistant",
				Content: text,
			}}, true
		}
	case "result":
		if text := strings.TrimSpace(event.Result); text != "" {
			return AgentTraceParsedEvent{Message: &AgentTraceMessage{
				Role:    "assistant",
				Content: text,
			}}, true
		}
	}
	return AgentTraceParsedEvent{}, false
}
