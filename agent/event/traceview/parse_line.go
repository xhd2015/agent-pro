package traceview

import "encoding/json"

var parseTraceLine func(json.RawMessage) (AgentTraceParsedEvent, bool)

func SetTraceLineParser(fn func(json.RawMessage) (AgentTraceParsedEvent, bool)) {
	parseTraceLine = fn
}