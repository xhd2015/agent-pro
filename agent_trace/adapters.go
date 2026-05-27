package agent_trace

import (
	"encoding/json"
	"sort"
)

type AgentTraceParsedEvent struct {
	Message  *AgentTraceMessage
	Activity *AgentTraceActivity
}

type AgentTraceAdapter interface {
	Name() string
	Parse(raw json.RawMessage) (AgentTraceParsedEvent, bool)
}

type registeredAgentTraceAdapter struct {
	priority int
	adapter  AgentTraceAdapter
}

var agentTraceAdapters []registeredAgentTraceAdapter

func RegisterAgentTraceAdapter(priority int, adapter AgentTraceAdapter) {
	if adapter == nil {
		return
	}
	agentTraceAdapters = append(agentTraceAdapters, registeredAgentTraceAdapter{
		priority: priority,
		adapter:  adapter,
	})
	sort.SliceStable(agentTraceAdapters, func(i, j int) bool {
		return agentTraceAdapters[i].priority < agentTraceAdapters[j].priority
	})
}

func parseAgentTraceLine(raw json.RawMessage) (AgentTraceParsedEvent, bool) {
	for _, registered := range agentTraceAdapters {
		if parsed, ok := registered.adapter.Parse(raw); ok {
			return parsed, true
		}
	}
	return AgentTraceParsedEvent{}, false
}
