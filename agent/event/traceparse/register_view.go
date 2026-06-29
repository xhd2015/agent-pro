package traceparse

import "github.com/xhd2015/agent-pro/agent/event/traceview"

func init() {
	traceview.SetTraceLineParser(ParseTraceLine)
}