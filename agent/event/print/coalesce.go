package print

import "github.com/xhd2015/agent-pro/agent/event/types"

// Coalescer tracks consecutive ActionMessage events and skips redundant
// PhaseEnd events when earlier phases (start/update/instant/end) were
// already seen for the same message ID.
//
// Non-ActionMessage events always pass through and reset the tracking state.
type Coalescer struct {
	lastMsgID string
	wasShown  bool
}

// ShouldSkip returns true if the event is a redundant ActionMessage
// with PhaseEnd that was already shown via prior phases for the same
// message ID.
func (c *Coalescer) ShouldSkip(event types.AgentEvent) bool {
	// Non-ActionMessage events always pass through and reset state.
	if event.Type != types.ActionMessage {
		c.lastMsgID = ""
		c.wasShown = false
		return false
	}

	// ActionMessage with PhaseEnd: skip if we've already shown content
	// for the same message ID.
	if event.Phase == types.PhaseEnd {
		if event.ID == c.lastMsgID && c.wasShown {
			return true
		}
		// First time seeing this PhaseEnd (or different ID).
		// Record it so subsequent PhaseEnd for same ID is skipped.
		c.lastMsgID = event.ID
		c.wasShown = true
		return false
	}

	// Any other phase (PhaseStart, PhaseUpdate, PhaseInstant):
	// never skip, and mark this ID as having content shown.
	c.lastMsgID = event.ID
	c.wasShown = true
	return false
}
