// Package types is deprecated: all symbols are aliases or delegates to agent/event.
//
// Migration:
//   - Timeline types (AgentTraceMessage, Message, ActivitySubtype, …) → agent/event/traceview
//   - ParseAgentTraceLine, RegisterAgentTraceAdapter → agent/event/traceparse
//   - TraceEvent, TraceItem, TraceMessage, … → agent/event/codex_types
//   - CompactTraceOutput, TitleFromIdentifier → agent/event/summary
//   - FileChange → agent/event/types
package types

import (
	"encoding/json"

	"github.com/xhd2015/agent-pro/agent/event/codex_types"
	"github.com/xhd2015/agent-pro/agent/event/summary"
	"github.com/xhd2015/agent-pro/agent/event/traceparse"
	"github.com/xhd2015/agent-pro/agent/event/traceview"
)

type AgentTraceMetadata = traceview.AgentTraceMetadata
type AgentTraceChild = traceview.AgentTraceChild
type AgentTraceSummary = traceview.AgentTraceSummary
type AgentTraceDetail = traceview.AgentTraceDetail
type AgentTraceUpdate = traceview.AgentTraceUpdate
type MessageRole = traceview.MessageRole
type ActivitySubtype = traceview.ActivitySubtype
type ActivityStatus = traceview.ActivityStatus
type AgentTraceMessage = traceview.AgentTraceMessage
type AgentTraceActivity = traceview.AgentTraceActivity
type AgentTraceFileChange = traceview.AgentTraceFileChange
type FileChange = traceview.FileChange
type Message = traceview.AgentTraceMessage
type ToolCallEvent = traceview.AgentTraceActivity
type AgentTraceParsedEvent = traceview.AgentTraceParsedEvent

const (
	RoleAssistant = traceview.RoleAssistant
	RoleToolCall  = traceview.RoleToolCall
	SubtypeStarted   = traceview.SubtypeStarted
	SubtypeUpdated   = traceview.SubtypeUpdated
	SubtypeCompleted = traceview.SubtypeCompleted
	StatusCompleted  = traceview.StatusCompleted
	StatusFailed     = traceview.StatusFailed
	StatusInProgress = traceview.StatusInProgress
	StatusWarning    = traceview.StatusWarning
	StatusPending    = traceview.StatusPending
)

type TraceEvent = codex_types.TraceEvent
type TraceMessage = codex_types.TraceMessage
type TraceContent = codex_types.TraceContent
type TraceItem = codex_types.TraceItem
type TraceTodoItem = codex_types.TraceTodoItem
type TracePlanItem = codex_types.TracePlanItem

type AgentTraceAdapter interface {
	Name() string
	Parse(raw json.RawMessage) (AgentTraceParsedEvent, bool)
}

func RegisterAgentTraceAdapter(priority int, adapter AgentTraceAdapter) {
	traceparse.RegisterAdapter(priority, adapter)
}

func ParseAgentTraceLine(raw json.RawMessage) (AgentTraceParsedEvent, bool) {
	return traceparse.ParseTraceLine(raw)
}

func TraceIsAssistantItem(itemType string) bool {
	return codex_types.TraceIsAssistantItem(itemType)
}

func TraceItemText(item *TraceItem) string {
	return codex_types.TraceItemText(item)
}

func TraceMessageText(message *TraceMessage) string {
	return codex_types.TraceMessageText(message)
}

func CompactTraceOutput(output string) string {
	return summary.CompactTraceOutput(output)
}

func TitleFromIdentifier(id string) string {
	return summary.TitleFromIdentifier(id)
}