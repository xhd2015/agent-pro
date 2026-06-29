package traceparse

import (
	"encoding/json"
	"strings"

	"github.com/xhd2015/agent-pro/agent/event/codex_types"
	"github.com/xhd2015/agent-pro/agent/event/traceview"
)

type genericTraceAdapter struct{}

func init() {
	RegisterAdapter(1000, genericTraceAdapter{})
}

func (genericTraceAdapter) Name() string {
	return "generic"
}

func (genericTraceAdapter) Parse(raw json.RawMessage) (traceview.AgentTraceParsedEvent, bool) {
	var event codex_types.TraceEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return traceview.AgentTraceParsedEvent{}, false
	}
	switch event.Type {
	case "assistant":
		if text := codex_types.TraceMessageText(event.Message); text != "" {
			return traceview.AgentTraceParsedEvent{Message: &traceview.AgentTraceMessage{
				Role:    traceview.RoleAssistant,
				Content: text,
			}}, true
		}
	case "result":
		if text := strings.TrimSpace(event.Result); text != "" {
			return traceview.AgentTraceParsedEvent{Message: &traceview.AgentTraceMessage{
				Role:    traceview.RoleAssistant,
				Content: text,
			}}, true
		}
	}
	return traceview.AgentTraceParsedEvent{}, false
}