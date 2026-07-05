package queue

import (
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// Batch is the slice consumed for one HTTP serve: leading think events plus one breakpoint.
type Batch struct {
	PrefixThink []types.AgentEvent
	Breakpoint  types.AgentEvent
}

// DequeueToBreakpoint pops leading think events and exactly one tool_call or message
// breakpoint from the front of queue. If the queue holds only think events (no breakpoint),
// nothing is consumed and ok is false.
func DequeueToBreakpoint(queue *[]types.AgentEvent) (batch Batch, ok bool) {
	if queue == nil || len(*queue) == 0 {
		return Batch{}, false
	}

	i := 0
	for i < len(*queue) && (*queue)[i].Type == types.ActionThink {
		batch.PrefixThink = append(batch.PrefixThink, (*queue)[i])
		i++
	}
	if i >= len(*queue) {
		return Batch{}, false
	}

	evt := (*queue)[i]
	if evt.Type != types.ActionToolCall && evt.Type != types.ActionMessage {
		return Batch{}, false
	}
	batch.Breakpoint = evt
	*queue = (*queue)[i+1:]
	return batch, true
}

// ConsumedEvents returns all AgentEvents served for one breakpoint dequeue.
func ConsumedEvents(batch Batch) []types.AgentEvent {
	out := make([]types.AgentEvent, 0, len(batch.PrefixThink)+1)
	out = append(out, batch.PrefixThink...)
	out = append(out, batch.Breakpoint)
	return out
}

// CollapsedThinkText joins prefix think event text with spaces.
func CollapsedThinkText(thinks []types.AgentEvent) string {
	var parts []string
	for _, evt := range thinks {
		if evt.Text != "" {
			parts = append(parts, evt.Text)
		}
	}
	return strings.Join(parts, " ")
}