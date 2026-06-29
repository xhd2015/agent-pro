// Package agent_trace is deprecated: trace types and adapter registration moved into
// agent/event (traceview, traceparse, summary, codex_types).
//
// This package remains a thin backward-compatible wrapper. New code should import the
// event packages directly instead of relying on these aliases.
//
// Migration:
//   - AgentTraceMetadata, AgentTraceMessage, Message, … → agent/event/traceview
//   - FileChange → agent/event/types or agent/event/traceview
//   - Adapter registration → blank import agent/event/traceparse
package agent_trace

import (
	"github.com/xhd2015/agent-pro/agent/event/traceview"
	"github.com/xhd2015/agent-pro/agent_trace/types"

	_ "github.com/xhd2015/agent-pro/agent/event/traceparse"
)

type AgentTraceMetadata = traceview.AgentTraceMetadata
type AgentTraceSummary = traceview.AgentTraceSummary
type AgentTraceDetail = traceview.AgentTraceDetail
type AgentTraceUpdate = traceview.AgentTraceUpdate
type AgentTraceChild = traceview.AgentTraceChild
type AgentTraceMessage = traceview.AgentTraceMessage
type AgentTraceActivity = traceview.AgentTraceActivity
type AgentTraceFileChange = traceview.AgentTraceFileChange
type Message = traceview.AgentTraceMessage
type ToolCallEvent = traceview.AgentTraceActivity
type FileChange = types.FileChange