// Deprecated: ParseMessages lives in agent/event/traceview.
//
// Migration:
//
//	import traceview "github.com/xhd2015/agent-pro/agent/event/traceview"
//	messages := traceview.ParseMessages(lines, createdAt)
package agent_trace

import (
	traceview "github.com/xhd2015/agent-pro/agent/event/traceview"
	"github.com/xhd2015/agent-pro/agent_trace/types"
)

func ParseMessages(lines []string, createdAt string) []types.Message {
	return traceview.ParseMessages(lines, createdAt)
}