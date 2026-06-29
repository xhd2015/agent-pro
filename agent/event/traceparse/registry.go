package traceparse

import (
	"encoding/json"
	"sort"

	"github.com/xhd2015/agent-pro/agent/event/traceview"
)

type AgentTraceAdapter interface {
	Name() string
	Parse(raw json.RawMessage) (traceview.AgentTraceParsedEvent, bool)
}

type registeredAgentTraceAdapter struct {
	priority int
	adapter  AgentTraceAdapter
}

var agentTraceAdapters []registeredAgentTraceAdapter

func RegisterAdapter(priority int, adapter AgentTraceAdapter) {
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

func ParseTraceLine(raw json.RawMessage) (traceview.AgentTraceParsedEvent, bool) {
	for _, registered := range agentTraceAdapters {
		if parsed, ok := registered.adapter.Parse(raw); ok {
			return parsed, true
		}
	}
	return traceview.AgentTraceParsedEvent{}, false
}