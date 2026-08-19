package queue

import (
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// DelayFor is the sum of DelayMs on events (sleep events and delay_ms on others).
func DelayFor(events []types.AgentEvent) time.Duration {
	var d time.Duration
	for _, evt := range events {
		if evt.DelayMs > 0 {
			d += time.Duration(evt.DelayMs) * time.Millisecond
		}
	}
	return d
}

// SleepFor sleeps DelayFor(events). No-op when the sum is 0.
func SleepFor(events []types.AgentEvent) {
	if d := DelayFor(events); d > 0 {
		time.Sleep(d)
	}
}
